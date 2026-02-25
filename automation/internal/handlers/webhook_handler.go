package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
	"github.com/TomasBack2Future/Kinetik/automation/internal/workflow"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/config"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/sirupsen/logrus"
)

type WebhookHandler struct {
	config       *config.Config
	orchestrator *workflow.Orchestrator
}

func NewWebhookHandler(cfg *config.Config, orchestrator *workflow.Orchestrator) *WebhookHandler {
	return &WebhookHandler{
		config:       cfg,
		orchestrator: orchestrator,
	}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")
	deliveryID := r.Header.Get("X-GitHub-Delivery")

	logger.WithFields(logrus.Fields{
		"event_type":  eventType,
		"delivery_id": deliveryID,
	}).Info("Received GitHub webhook")

	// Read payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read webhook payload", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Route to appropriate handler
	switch eventType {
	case "ping":
		logger.Info("Received ping event")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "pong")
		return
	case "installation", "installation_repositories":
		logger.Info("Received installation event")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "ok")
		return
	}

	// Parse common fields to get repository (for events that have repositories)
	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Error("Failed to parse webhook payload", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Verify repository is allowed
	if !h.isAllowedRepo(payload.Repository.FullName) {
		logger.WithField("repo", payload.Repository.FullName).Warn("Repository not allowed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Route to repository-specific handlers
	switch eventType {
	case "issues":
		h.handleIssuesEvent(body, deliveryID)
	case "issue_comment":
		h.handleIssueCommentEvent(body, deliveryID)
	case "pull_request":
		h.handlePullRequestEvent(body, deliveryID)
	case "pull_request_review":
		h.handlePullRequestReviewEvent(body, deliveryID)
	case "pull_request_review_comment":
		h.handlePullRequestReviewCommentEvent(body, deliveryID)
	default:
		logger.WithField("event_type", eventType).Debug("Ignoring event type")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *WebhookHandler) isAllowedRepo(fullName string) bool {
	for _, allowed := range h.config.GitHub.AllowedRepos {
		if allowed == fullName {
			return true
		}
	}
	return false
}

func (h *WebhookHandler) handleIssuesEvent(body []byte, deliveryID string) {
	var event types.IssueEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse issues event", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"action":       event.Action,
		"issue_number": event.Issue.Number,
		"repo":         event.Repository.FullName,
	}).Info("Processing issues event")

	// Only process opened issues
	if event.Action == "opened" {
		go h.orchestrator.HandleNewIssue(&event)
	}
}

func (h *WebhookHandler) handleIssueCommentEvent(body []byte, deliveryID string) {
	var event types.IssueCommentEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse issue_comment event", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"action":       event.Action,
		"issue_number": event.Issue.Number,
		"repo":         event.Repository.FullName,
	}).Info("Processing issue_comment event")

	// Only process created comments
	if event.Action == "created" {
		// Ignore comments made by the bot itself to prevent feedback loops
		if strings.EqualFold(event.Comment.User.Login, h.getBotUsername()) {
			logger.WithField("issue_number", event.Issue.Number).Debug("Ignoring comment from bot itself")
			return
		}

		// Check for approval keywords or bot mention
		commentBody := strings.ToLower(event.Comment.Body)

		// Check for bot mention (case-insensitive)
		if strings.Contains(commentBody, "@"+strings.ToLower(h.getBotUsername())) {
			go h.orchestrator.HandleIssueMention(&event)
			return
		}

		// Check for approval keywords
		for _, keyword := range h.config.Workflow.ApprovalKeywords {
			if strings.Contains(commentBody, strings.ToLower(keyword)) {
				go h.orchestrator.HandleIssueApproval(&event)
				return
			}
		}
	}
}

func (h *WebhookHandler) handlePullRequestEvent(body []byte, deliveryID string) {
	var event types.PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse pull_request event", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"action":    event.Action,
		"pr_number": event.PullRequest.Number,
		"repo":      event.Repository.FullName,
	}).Info("Processing pull_request event")

	// Process opened or synchronized PRs
	if event.Action == "opened" || event.Action == "synchronize" {
		go h.orchestrator.HandlePullRequest(&event)
	}
}

func (h *WebhookHandler) handlePullRequestReviewEvent(body []byte, deliveryID string) {
	var event types.PullRequestReviewEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse pull_request_review event", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"action":    event.Action,
		"pr_number": event.PullRequest.Number,
		"repo":      event.Repository.FullName,
	}).Info("Processing pull_request_review event")

	if event.Action == "submitted" {
		go h.orchestrator.HandlePullRequestReview(&event)
	}
}

func (h *WebhookHandler) handlePullRequestReviewCommentEvent(body []byte, deliveryID string) {
	var event types.PullRequestReviewCommentEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse pull_request_review_comment event", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"action":    event.Action,
		"pr_number": event.PullRequest.Number,
		"repo":      event.Repository.FullName,
	}).Info("Processing pull_request_review_comment event")

	if event.Action == "created" {
		commentBody := strings.ToLower(event.Comment.Body)
		if strings.Contains(commentBody, "@"+strings.ToLower(h.getBotUsername())) {
			go h.orchestrator.HandlePullRequestComment(&event)
		}
	}
}

func (h *WebhookHandler) getBotUsername() string {
	// Return bot username from configuration
	if h.config.GitHub.BotUsername != "" {
		return h.config.GitHub.BotUsername
	}
	// Fallback to default if not configured
	return "kinetik-bot"
}
