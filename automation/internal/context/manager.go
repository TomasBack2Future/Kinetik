package context

import (
	"context"
	"fmt"

	"github.com/TomasBack2Future/Kinetik/automation/internal/repository"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/sirupsen/logrus"
)

const (
	StateAnalyzing       = "analyzing"
	StatePendingApproval = "pending_approval"
	StateExecuting       = "executing"
	StateCompleted       = "completed"
	StateFailed          = "failed"
)

type Manager struct {
	repo *repository.ConversationRepo
}

func NewManager(repo *repository.ConversationRepo) *Manager {
	return &Manager{
		repo: repo,
	}
}

// GetOrCreateConversation retrieves existing conversation or creates a new one
func (m *Manager) GetOrCreateConversation(ctx context.Context, repoFullName string, issueNumber int) (*repository.Conversation, error) {
	// Try to get existing conversation
	conv, err := m.repo.GetByIssue(ctx, repoFullName, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if conv != nil {
		logger.WithFields(logrus.Fields{
			"conversation_id": conv.ID,
			"repo":            repoFullName,
			"issue_number":    issueNumber,
			"state":           conv.State,
		}).Info("Found existing conversation")
		return conv, nil
	}

	// Create new conversation
	conv = &repository.Conversation{
		RepoFullName: repoFullName,
		IssueNumber:  issueNumber,
		State:        StateAnalyzing,
		Context:      make(map[string]interface{}),
	}

	if err := m.repo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"repo":            repoFullName,
		"issue_number":    issueNumber,
	}).Info("Created new conversation")

	return conv, nil
}

// GetConversationByPR retrieves conversation by PR number
func (m *Manager) GetConversationByPR(ctx context.Context, repoFullName string, prNumber int) (*repository.Conversation, error) {
	conv, err := m.repo.GetByPR(ctx, repoFullName, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation by PR: %w", err)
	}

	if conv == nil {
		// Create new conversation for PR
		conv = &repository.Conversation{
			RepoFullName: repoFullName,
			PRNumber:     prNumber,
			State:        StateAnalyzing,
			Context:      make(map[string]interface{}),
		}

		if err := m.repo.Create(ctx, conv); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}

		logger.WithFields(logrus.Fields{
			"conversation_id": conv.ID,
			"repo":            repoFullName,
			"pr_number":       prNumber,
		}).Info("Created new conversation for PR")
	}

	return conv, nil
}

// UpdateState updates the conversation state
func (m *Manager) UpdateState(ctx context.Context, conv *repository.Conversation, state string) error {
	conv.State = state

	if err := m.repo.Update(ctx, conv); err != nil {
		return fmt.Errorf("failed to update conversation state: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"new_state":       state,
	}).Info("Updated conversation state")

	return nil
}

// UpdateSessionID updates the Claude session ID
func (m *Manager) UpdateSessionID(ctx context.Context, conv *repository.Conversation, sessionID string) error {
	conv.ClaudeSessionID = sessionID

	if err := m.repo.Update(ctx, conv); err != nil {
		return fmt.Errorf("failed to update session ID: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"session_id":      sessionID,
	}).Info("Updated Claude session ID")

	return nil
}

// AddToContext adds data to conversation context
func (m *Manager) AddToContext(ctx context.Context, conv *repository.Conversation, key string, value interface{}) error {
	if conv.Context == nil {
		conv.Context = make(map[string]interface{})
	}

	conv.Context[key] = value

	if err := m.repo.Update(ctx, conv); err != nil {
		return fmt.Errorf("failed to update context: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"context_key":     key,
	}).Debug("Added to conversation context")

	return nil
}

// GetFromContext retrieves data from conversation context
func (m *Manager) GetFromContext(conv *repository.Conversation, key string) (interface{}, bool) {
	if conv.Context == nil {
		return nil, false
	}

	value, ok := conv.Context[key]
	return value, ok
}

// BuildContextString builds a human-readable context string for prompts
func (m *Manager) BuildContextString(conv *repository.Conversation) string {
	if len(conv.Context) == 0 {
		return ""
	}

	// Extract key information from context
	var contextStr string

	if analysis, ok := conv.Context["analysis"]; ok {
		contextStr += fmt.Sprintf("Previous Analysis:\n%v\n\n", analysis)
	}

	if plan, ok := conv.Context["implementation_plan"]; ok {
		contextStr += fmt.Sprintf("Implementation Plan:\n%v\n\n", plan)
	}

	if comments, ok := conv.Context["comments"]; ok {
		contextStr += fmt.Sprintf("Comments:\n%v\n\n", comments)
	}

	return contextStr
}

// AddTokenUsage increments token usage counters for a conversation
func (m *Manager) AddTokenUsage(ctx context.Context, conv *repository.Conversation, totalTokens, inputTokens, outputTokens, toolUses int, durationMs int64) error {
	if err := m.repo.AddTokenUsage(ctx, conv.ID, totalTokens, inputTokens, outputTokens, toolUses, durationMs); err != nil {
		logger.Error("Failed to add token usage", err)
		return err
	}

	// Update in-memory conversation object
	conv.TotalTokens += totalTokens
	conv.InputTokens += inputTokens
	conv.OutputTokens += outputTokens
	conv.ToolUses += toolUses
	conv.DurationMs += durationMs

	logger.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"total_tokens":    conv.TotalTokens,
		"tool_uses":       conv.ToolUses,
	}).Info("Updated token usage")

	return nil
}
