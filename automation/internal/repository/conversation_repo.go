package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID              string
	RepoFullName    string
	IssueNumber     int
	PRNumber        int
	State           string
	ClaudeSessionID string
	Context         map[string]interface{}
	TotalTokens     int
	InputTokens     int
	OutputTokens    int
	ToolUses        int
	DurationMs      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ConversationRepo struct {
	db *PostgresDB
}

func NewConversationRepo(db *PostgresDB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// Create creates a new conversation
func (r *ConversationRepo) Create(ctx context.Context, conv *Conversation) error {
	if conv.ID == "" {
		conv.ID = uuid.New().String()
	}

	contextJSON, err := json.Marshal(conv.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	query := `
		INSERT INTO conversations (id, repo_full_name, issue_number, pr_number, state, claude_session_id, context,
			total_tokens, input_tokens, output_tokens, tool_uses, duration_ms, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	conv.CreatedAt = now
	conv.UpdatedAt = now

	_, err = r.db.DB().ExecContext(ctx, query,
		conv.ID,
		conv.RepoFullName,
		conv.IssueNumber,
		conv.PRNumber,
		conv.State,
		conv.ClaudeSessionID,
		contextJSON,
		conv.TotalTokens,
		conv.InputTokens,
		conv.OutputTokens,
		conv.ToolUses,
		conv.DurationMs,
		conv.CreatedAt,
		conv.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	return nil
}

// Get retrieves a conversation by ID
func (r *ConversationRepo) Get(ctx context.Context, id string) (*Conversation, error) {
	query := `
		SELECT id, repo_full_name, issue_number, pr_number, state, claude_session_id, context,
			total_tokens, input_tokens, output_tokens, tool_uses, duration_ms, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	var conv Conversation
	var contextJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.RepoFullName,
		&conv.IssueNumber,
		&conv.PRNumber,
		&conv.State,
		&conv.ClaudeSessionID,
		&contextJSON,
		&conv.TotalTokens,
		&conv.InputTokens,
		&conv.OutputTokens,
		&conv.ToolUses,
		&conv.DurationMs,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &conv.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &conv, nil
}

// GetByIssue retrieves a conversation by repository and issue number
func (r *ConversationRepo) GetByIssue(ctx context.Context, repoFullName string, issueNumber int) (*Conversation, error) {
	query := `
		SELECT id, repo_full_name, issue_number, pr_number, state, claude_session_id, context,
			total_tokens, input_tokens, output_tokens, tool_uses, duration_ms, created_at, updated_at
		FROM conversations
		WHERE repo_full_name = $1 AND issue_number = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var conv Conversation
	var contextJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, repoFullName, issueNumber).Scan(
		&conv.ID,
		&conv.RepoFullName,
		&conv.IssueNumber,
		&conv.PRNumber,
		&conv.State,
		&conv.ClaudeSessionID,
		&contextJSON,
		&conv.TotalTokens,
		&conv.InputTokens,
		&conv.OutputTokens,
		&conv.ToolUses,
		&conv.DurationMs,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation by issue: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &conv.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &conv, nil
}

// GetByPR retrieves a conversation by repository and PR number
func (r *ConversationRepo) GetByPR(ctx context.Context, repoFullName string, prNumber int) (*Conversation, error) {
	query := `
		SELECT id, repo_full_name, issue_number, pr_number, state, claude_session_id, context,
			total_tokens, input_tokens, output_tokens, tool_uses, duration_ms, created_at, updated_at
		FROM conversations
		WHERE repo_full_name = $1 AND pr_number = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var conv Conversation
	var contextJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, repoFullName, prNumber).Scan(
		&conv.ID,
		&conv.RepoFullName,
		&conv.IssueNumber,
		&conv.PRNumber,
		&conv.State,
		&conv.ClaudeSessionID,
		&contextJSON,
		&conv.TotalTokens,
		&conv.InputTokens,
		&conv.OutputTokens,
		&conv.ToolUses,
		&conv.DurationMs,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation by PR: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &conv.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &conv, nil
}

// Update updates an existing conversation
func (r *ConversationRepo) Update(ctx context.Context, conv *Conversation) error {
	contextJSON, err := json.Marshal(conv.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	query := `
		UPDATE conversations
		SET state = $1, claude_session_id = $2, context = $3, pr_number = $4,
			total_tokens = $5, input_tokens = $6, output_tokens = $7, tool_uses = $8, duration_ms = $9,
			updated_at = $10
		WHERE id = $11
	`

	conv.UpdatedAt = time.Now()

	result, err := r.db.DB().ExecContext(ctx, query,
		conv.State,
		conv.ClaudeSessionID,
		contextJSON,
		conv.PRNumber,
		conv.TotalTokens,
		conv.InputTokens,
		conv.OutputTokens,
		conv.ToolUses,
		conv.DurationMs,
		conv.UpdatedAt,
		conv.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}

// AddTokenUsage increments the token usage counters for a conversation
func (r *ConversationRepo) AddTokenUsage(ctx context.Context, id string, totalTokens, inputTokens, outputTokens, toolUses int, durationMs int64) error {
	query := `
		UPDATE conversations
		SET total_tokens = total_tokens + $1,
			input_tokens = input_tokens + $2,
			output_tokens = output_tokens + $3,
			tool_uses = tool_uses + $4,
			duration_ms = duration_ms + $5,
			updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.DB().ExecContext(ctx, query,
		totalTokens,
		inputTokens,
		outputTokens,
		toolUses,
		durationMs,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to add token usage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}

// Delete deletes a conversation
func (r *ConversationRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = $1`

	result, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}
