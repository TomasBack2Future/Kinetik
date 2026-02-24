-- Add token tracking columns to conversations table
ALTER TABLE conversations
ADD COLUMN total_tokens INTEGER DEFAULT 0,
ADD COLUMN input_tokens INTEGER DEFAULT 0,
ADD COLUMN output_tokens INTEGER DEFAULT 0,
ADD COLUMN tool_uses INTEGER DEFAULT 0,
ADD COLUMN duration_ms BIGINT DEFAULT 0;

-- Create index for token usage queries
CREATE INDEX idx_conversations_total_tokens ON conversations(total_tokens DESC);

-- Add comment for documentation
COMMENT ON COLUMN conversations.total_tokens IS 'Cumulative total tokens used across all Claude Code executions in this conversation';
COMMENT ON COLUMN conversations.input_tokens IS 'Cumulative input tokens used';
COMMENT ON COLUMN conversations.output_tokens IS 'Cumulative output tokens used';
COMMENT ON COLUMN conversations.tool_uses IS 'Cumulative number of tool uses';
COMMENT ON COLUMN conversations.duration_ms IS 'Cumulative execution duration in milliseconds';
