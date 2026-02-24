-- Remove token tracking columns from conversations table
DROP INDEX IF EXISTS idx_conversations_total_tokens;

ALTER TABLE conversations
DROP COLUMN IF EXISTS total_tokens,
DROP COLUMN IF EXISTS input_tokens,
DROP COLUMN IF EXISTS output_tokens,
DROP COLUMN IF EXISTS tool_uses,
DROP COLUMN IF EXISTS duration_ms;
