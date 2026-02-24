-- Create conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id VARCHAR(36) PRIMARY KEY,
    repo_full_name VARCHAR(255) NOT NULL,
    issue_number INTEGER DEFAULT 0,
    pr_number INTEGER DEFAULT 0,
    state VARCHAR(50) NOT NULL,
    claude_session_id VARCHAR(100),
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient lookups
CREATE INDEX idx_conversations_repo_issue ON conversations(repo_full_name, issue_number);
CREATE INDEX idx_conversations_repo_pr ON conversations(repo_full_name, pr_number);
CREATE INDEX idx_conversations_state ON conversations(state);
CREATE INDEX idx_conversations_created_at ON conversations(created_at DESC);

-- Create events table for audit log
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    conversation_id VARCHAR(36) REFERENCES conversations(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_conversation_id ON events(conversation_id);
CREATE INDEX idx_events_created_at ON events(created_at DESC);
