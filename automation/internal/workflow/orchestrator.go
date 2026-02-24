package workflow

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/TomasBack2Future/Kinetik/automation/internal/claude"
	contextmgr "github.com/TomasBack2Future/Kinetik/automation/internal/context"
	"github.com/TomasBack2Future/Kinetik/automation/internal/github"
	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/config"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/sirupsen/logrus"
)

type Orchestrator struct {
	config         *config.Config
	claudeClient   *claude.CLIClient
	promptBuilder  *claude.PromptBuilder
	contextManager *contextmgr.Manager
	githubClient   *github.Client
	validator      *IssueValidator
	issueQueue     *IssueQueue
}

func NewOrchestrator(
	cfg *config.Config,
	claudeClient *claude.CLIClient,
	promptBuilder *claude.PromptBuilder,
	contextManager *contextmgr.Manager,
	githubClient *github.Client,
) *Orchestrator {
	o := &Orchestrator{
		config:         cfg,
		claudeClient:   claudeClient,
		promptBuilder:  promptBuilder,
		contextManager: contextManager,
		githubClient:   githubClient,
		validator:      NewIssueValidator(),
	}

	// Initialize queue with processor function
	o.issueQueue = NewIssueQueue(o.processQueuedIssue)

	return o
}

// HandleNewIssue processes a new issue creation event
func (o *Orchestrator) HandleNewIssue(event *types.IssueEvent) {
	logger.WithFields(logrus.Fields{
		"repo":         event.Repository.FullName,
		"issue_number": event.Issue.Number,
		"issue_url":    event.Issue.HTMLURL,
		"title":        event.Issue.Title,
		"author":       event.Issue.User.Login,
	}).Info("Handling new issue")

	// Validate issue has required information
	validation := o.validator.ValidateIssue(&event.Issue)

	if !validation.Valid {
		logger.WithFields(logrus.Fields{
			"issue_number":   event.Issue.Number,
			"missing_fields": validation.MissingFields,
		}).Info("Issue missing required information, requesting details")

		// Post comment asking for missing info
		owner, repo := github.ParseRepoOwner(event.Repository.FullName)
		comment := o.validator.BuildRequestInfoComment(validation)

		if err := o.githubClient.CreateIssueComment(owner, repo, event.Issue.Number, comment); err != nil {
			logger.Error("Failed to post info request comment", err)
		}

		// Add label to track incomplete issues
		if err := o.githubClient.AddIssueLabel(owner, repo, event.Issue.Number, "needs-info"); err != nil {
			logger.Error("Failed to add needs-info label", err)
		}

		return
	}

	// Add to queue for processing
	branchName := o.issueQueue.Enqueue(event)

	logger.WithFields(logrus.Fields{
		"issue_number": event.Issue.Number,
		"branch_name":  branchName,
		"queue_length": o.issueQueue.GetQueueLength(),
	}).Info("Issue queued for processing")
}

// processQueuedIssue handles the actual processing of a queued issue
func (o *Orchestrator) processQueuedIssue(ctx context.Context, queued *QueuedIssue) error {
	event := queued.Event

	logger.WithFields(logrus.Fields{
		"issue_number": event.Issue.Number,
		"branch_name":  queued.BranchName,
	}).Info("Processing issue from queue")

	// Create branch for this issue
	owner, repo := github.ParseRepoOwner(event.Repository.FullName)
	if err := o.githubClient.CreateBranch(owner, repo, queued.BranchName, "main"); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create branch, continuing anyway")
		// Don't return error - continue with processing
	}

	// Get or create conversation
	conv, err := o.contextManager.GetOrCreateConversation(ctx, event.Repository.FullName, event.Issue.Number)
	if err != nil {
		return err
	}

	// Store branch name in context
	if err := o.contextManager.AddToContext(ctx, conv, "branch_name", queued.BranchName); err != nil {
		logger.Error("Failed to add branch name to context", err)
	}

	// Build prompt
	contextStr := o.contextManager.BuildContextString(conv)
	prompt := o.promptBuilder.BuildIssueAnalysisPrompt(event, contextStr)

	// Execute Claude with NEW session ID
	logger.WithField("conversation_id", conv.ID).Info("Executing Claude for issue analysis")

	// Use conversation-specific work directory for isolated stdout/stderr logs
	workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
	result, err := o.claudeClient.Execute(ctx, prompt, queued.BranchName, workDir)
	if err != nil {
		logger.Error("Failed to execute Claude", err)
		if updateErr := o.contextManager.UpdateState(ctx, conv, contextmgr.StateFailed); updateErr != nil {
			logger.Error("Failed to update state to failed", updateErr)
		}
		return err
	}

	// Update conversation
	if err := o.contextManager.UpdateSessionID(ctx, conv, result.SessionID); err != nil {
		logger.Error("Failed to update session ID", err)
	}
	if err := o.contextManager.AddTokenUsage(ctx, conv, result.TotalTokens, result.InputTokens, result.OutputTokens, result.ToolUses, result.DurationMs); err != nil {
		logger.Error("Failed to add token usage", err)
	}
	if err := o.contextManager.UpdateState(ctx, conv, contextmgr.StatePendingApproval); err != nil {
		logger.Error("Failed to update state", err)
	}
	if err := o.contextManager.AddToContext(ctx, conv, "analysis", result.Output); err != nil {
		logger.Error("Failed to add analysis to context", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"session_id":      result.SessionID,
		"branch_name":     queued.BranchName,
	}).Info("Issue analysis completed")

	return nil
}

// HandleIssueApproval processes an approval comment on an issue
func (o *Orchestrator) HandleIssueApproval(event *types.IssueCommentEvent) {
	ctx := context.Background()

	logger.WithFields(logrus.Fields{
		"repo":         event.Repository.FullName,
		"issue_number": event.Issue.Number,
		"issue_url":    event.Issue.HTMLURL,
		"commenter":    event.Comment.User.Login,
		"comment_body": event.Comment.Body,
	}).Info("Handling issue approval")

	// Get conversation
	conv, err := o.contextManager.GetOrCreateConversation(ctx, event.Repository.FullName, event.Issue.Number)
	if err != nil {
		logger.Error("Failed to get conversation", err)
		return
	}

	// Check if conversation is in pending approval state
	if conv.State != contextmgr.StatePendingApproval {
		logger.WithFields(logrus.Fields{
			"conversation_id": conv.ID,
			"current_state":   conv.State,
		}).Warn("Conversation not in pending approval state")
		return
	}

	// Update state to executing
	if err := o.contextManager.UpdateState(ctx, conv, contextmgr.StateExecuting); err != nil {
		logger.Error("Failed to update state to executing", err)
		return
	}

	// Build implementation prompt
	contextStr := o.contextManager.BuildContextString(conv)
	prompt := o.promptBuilder.BuildIssueImplementationPrompt(event, contextStr)

	// Execute Claude with previous session ID for continuity
	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"session_id":      conv.ClaudeSessionID,
	}).Info("Executing Claude for implementation")

	// Use conversation-specific work directory for isolated stdout/stderr logs
	workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
	result, err := o.claudeClient.Execute(ctx, prompt, "", workDir)
	if err != nil {
		logger.Error("Failed to execute Claude", err)
		if updateErr := o.contextManager.UpdateState(ctx, conv, contextmgr.StateFailed); updateErr != nil {
			logger.Error("Failed to update state to failed", updateErr)
		}
		return
	}

	// Update conversation
	if err := o.contextManager.AddTokenUsage(ctx, conv, result.TotalTokens, result.InputTokens, result.OutputTokens, result.ToolUses, result.DurationMs); err != nil {
		logger.Error("Failed to add token usage", err)
	}
	if err := o.contextManager.UpdateState(ctx, conv, contextmgr.StateCompleted); err != nil {
		logger.Error("Failed to update state to completed", err)
	}
	if err := o.contextManager.AddToContext(ctx, conv, "implementation", result.Output); err != nil {
		logger.Error("Failed to add implementation to context", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"session_id":      result.SessionID,
	}).Info("Implementation completed")
}

// HandleIssueMention processes a @bot mention in an issue comment
func (o *Orchestrator) HandleIssueMention(event *types.IssueCommentEvent) {
	ctx := context.Background()

	logger.WithFields(logrus.Fields{
		"repo":         event.Repository.FullName,
		"issue_number": event.Issue.Number,
		"issue_url":    event.Issue.HTMLURL,
		"commenter":    event.Comment.User.Login,
		"comment_url":  event.Comment.HTMLURL,
		"comment_body": event.Comment.Body,
	}).Info("Handling bot mention in issue")

	// Check if this is a response to a "needs-info" request
	// Re-validate the issue with updated information
	validation := o.validator.ValidateIssue(&event.Issue)

	if !validation.Valid {
		// Still missing info, remind them
		owner, repo := github.ParseRepoOwner(event.Repository.FullName)
		comment := "I still need some additional information:\n\n" + o.validator.BuildRequestInfoComment(validation)

		if err := o.githubClient.CreateIssueComment(owner, repo, event.Issue.Number, comment); err != nil {
			logger.Error("Failed to post reminder comment", err)
		}
		return
	}

	// Info is now complete, check if already queued/processed
	conv, err := o.contextManager.GetOrCreateConversation(ctx, event.Repository.FullName, event.Issue.Number)
	if err != nil {
		logger.Error("Failed to get/create conversation", err)
		return
	}

	// If conversation exists and is not in initial state, just respond to the mention
	if conv.State != "" && conv.State != contextmgr.StateFailed {
		// Build prompt for general mention response
		contextStr := o.contextManager.BuildContextString(conv)
		prompt := o.promptBuilder.BuildIssueMentionPrompt(event, contextStr)

		// Execute Claude with NEW session ID
		workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
		result, err := o.claudeClient.Execute(ctx, prompt, "", workDir)
		if err != nil {
			logger.Error("Failed to execute Claude", err)
			return
		}

		// Update conversation
		if err := o.contextManager.UpdateSessionID(ctx, conv, result.SessionID); err != nil {
			logger.Error("Failed to update session ID", err)
		}
		if err := o.contextManager.AddTokenUsage(ctx, conv, result.TotalTokens, result.InputTokens, result.OutputTokens, result.ToolUses, result.DurationMs); err != nil {
			logger.Error("Failed to add token usage", err)
		}
		if err := o.contextManager.AddToContext(ctx, conv, "mention_response", result.Output); err != nil {
			logger.Error("Failed to add mention response to context", err)
		}

		logger.WithField("conversation_id", conv.ID).Info("Bot mention handled")
		return
	}

	// New issue with complete info - add to queue
	issueEvent := &types.IssueEvent{
		Action:     "opened",
		Issue:      event.Issue,
		Repository: event.Repository,
		Sender:     event.Sender,
	}

	branchName := o.issueQueue.Enqueue(issueEvent)

	logger.WithFields(logrus.Fields{
		"issue_number": event.Issue.Number,
		"branch_name":  branchName,
		"queue_length": o.issueQueue.GetQueueLength(),
	}).Info("Issue queued for processing after info provided")

	// Post acknowledgment
	owner, repo := github.ParseRepoOwner(event.Repository.FullName)
	comment := fmt.Sprintf("Thanks for the info! I've added this to my queue. Currently processing %d issue(s). I'll work on this on branch `%s`.",
		o.issueQueue.GetQueueLength(), branchName)

	if err := o.githubClient.CreateIssueComment(owner, repo, event.Issue.Number, comment); err != nil {
		logger.Error("Failed to post queue acknowledgment", err)
	}
}

// HandlePullRequest processes a pull request event
func (o *Orchestrator) HandlePullRequest(event *types.PullRequestEvent) {
	ctx := context.Background()

	logger.WithFields(logrus.Fields{
		"repo":      event.Repository.FullName,
		"pr_number": event.PullRequest.Number,
		"action":    event.Action,
	}).Info("Handling pull request event")

	// Get or create conversation for PR
	conv, err := o.contextManager.GetConversationByPR(ctx, event.Repository.FullName, event.PullRequest.Number)
	if err != nil {
		logger.Error("Failed to get/create PR conversation", err)
		return
	}

	// Store PR info in context
	if err := o.contextManager.AddToContext(ctx, conv, "pr_title", event.PullRequest.Title); err != nil {
		logger.Error("Failed to add PR title to context", err)
	}
	if err := o.contextManager.AddToContext(ctx, conv, "pr_body", event.PullRequest.Body); err != nil {
		logger.Error("Failed to add PR body to context", err)
	}

	logger.WithField("conversation_id", conv.ID).Info("PR event processed")
}

// HandlePullRequestReview processes a pull request review event
func (o *Orchestrator) HandlePullRequestReview(event *types.PullRequestReviewEvent) {
	ctx := context.Background()

	logger.WithFields(logrus.Fields{
		"repo":      event.Repository.FullName,
		"pr_number": event.PullRequest.Number,
		"reviewer":  event.Review.User.Login,
		"state":     event.Review.State,
	}).Info("Handling pull request review")

	// Get conversation
	conv, err := o.contextManager.GetConversationByPR(ctx, event.Repository.FullName, event.PullRequest.Number)
	if err != nil {
		logger.Error("Failed to get PR conversation", err)
		return
	}

	// Only process changes_requested reviews
	if event.Review.State != "changes_requested" {
		logger.Debug("Review state not changes_requested, skipping")
		return
	}

	// Build prompt
	contextStr := o.contextManager.BuildContextString(conv)
	prompt := o.promptBuilder.BuildPRGeneralReviewPrompt(event, contextStr)

	// Execute Claude with conversation-specific work directory for isolated stdout/stderr logs
	workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
	result, err := o.claudeClient.Execute(ctx, prompt, "", workDir)
	if err != nil {
		logger.Error("Failed to execute Claude", err)
		return
	}

	// Update conversation
	if err := o.contextManager.UpdateSessionID(ctx, conv, result.SessionID); err != nil {
		logger.Error("Failed to update session ID", err)
	}
	if err := o.contextManager.AddTokenUsage(ctx, conv, result.TotalTokens, result.InputTokens, result.OutputTokens, result.ToolUses, result.DurationMs); err != nil {
		logger.Error("Failed to add token usage", err)
	}
	if err := o.contextManager.AddToContext(ctx, conv, "review_response", result.Output); err != nil {
		logger.Error("Failed to add review response to context", err)
	}

	logger.WithField("conversation_id", conv.ID).Info("PR review handled")
}

// HandlePullRequestComment processes a comment on a pull request
func (o *Orchestrator) HandlePullRequestComment(event *types.PullRequestReviewCommentEvent) {
	ctx := context.Background()

	logger.WithFields(logrus.Fields{
		"repo":      event.Repository.FullName,
		"pr_number": event.PullRequest.Number,
		"commenter": event.Comment.User.Login,
	}).Info("Handling pull request comment")

	// Get conversation
	conv, err := o.contextManager.GetConversationByPR(ctx, event.Repository.FullName, event.PullRequest.Number)
	if err != nil {
		logger.Error("Failed to get PR conversation", err)
		return
	}

	// Build prompt
	contextStr := o.contextManager.BuildContextString(conv)
	prompt := o.promptBuilder.BuildPRReviewPrompt(event, contextStr)

	// Execute Claude with conversation-specific work directory for isolated stdout/stderr logs
	workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
	result, err := o.claudeClient.Execute(ctx, prompt, "", workDir)
	if err != nil {
		logger.Error("Failed to execute Claude", err)
		return
	}

	// Update conversation
	if err := o.contextManager.UpdateSessionID(ctx, conv, result.SessionID); err != nil {
		logger.Error("Failed to update session ID", err)
	}
	if err := o.contextManager.AddTokenUsage(ctx, conv, result.TotalTokens, result.InputTokens, result.OutputTokens, result.ToolUses, result.DurationMs); err != nil {
		logger.Error("Failed to add token usage", err)
	}
	if err := o.contextManager.AddToContext(ctx, conv, "comment_response", result.Output); err != nil {
		logger.Error("Failed to add comment response to context", err)
	}

	logger.WithField("conversation_id", conv.ID).Info("PR comment handled")
}
