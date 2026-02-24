# Session-Based stdout/stderr Routing Implementation

## Problem
Previously, all Claude CLI executions wrote their stdout and stderr to the same files:
```
/tmp/claude-sessions/stdout.log
/tmp/claude-sessions/stderr.log
```

This caused multiple concurrent sessions to overwrite each other's logs, making debugging and log analysis impossible.

## Solution
Implemented session-based folder routing using conversation IDs. Each conversation now gets its own isolated directory:

```
/tmp/claude-sessions/
├── [conversation-uuid-1]/
│   ├── stdout.log
│   └── stderr.log
├── [conversation-uuid-2]/
│   ├── stdout.log
│   └── stderr.log
└── [conversation-uuid-3]/
    ├── stdout.log
    └── stderr.log
```

## Changes Made

### File: `/root/CodeBase/code/Kinetik/automation/internal/workflow/orchestrator.go`

**1. Added filepath import** (line 5)
```go
import (
    "context"
    "path/filepath"  // Added
    ...
)
```

**2. Updated all Claude execution calls** (5 locations)

Each `Execute()` call now passes a conversation-specific work directory:

```go
// Before:
result, err := o.claudeClient.Execute(ctx, prompt, "", "")

// After:
workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
result, err := o.claudeClient.Execute(ctx, prompt, "", workDir)
```

**Updated locations:**
- Line 63: `HandleNewIssue()` - Issue analysis
- Line 124: `HandleIssueApproval()` - Issue implementation
- Line 169: `HandleIssueMention()` - Bot mention responses
- Line 236: `HandlePullRequestReview()` - PR review handling
- Line 272: `HandlePullRequestComment()` - PR comment handling

## How It Works

1. **Conversation Creation**: When a new issue/PR event is received, a conversation is created with a unique UUID
   - Generated via: `uuid.New().String()`
   - Example: `"a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d"`

2. **Work Directory Construction**: Before executing Claude CLI, the orchestrator builds a session-specific path:
   ```go
   workDir := filepath.Join(o.config.Claude.WorkDir, conv.ID)
   // Example: "/tmp/claude-sessions/a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d"
   ```

3. **Directory Creation**: The CLI client's `executeOnce()` function creates the directory if it doesn't exist:
   ```go
   if err := os.MkdirAll(workDir, 0755); err != nil {
       return nil, fmt.Errorf("failed to create work directory: %w", err)
   }
   ```

4. **Log File Creation**: stdout and stderr files are created in the session-specific directory:
   ```go
   stdoutPath := filepath.Join(workDir, "stdout.log")
   stderrPath := filepath.Join(workDir, "stderr.log")
   ```

## Benefits

1. **Isolation**: Each conversation's logs are completely isolated
2. **Concurrent Safety**: Multiple Claude executions can run simultaneously without conflicts
3. **Debugging**: Easy to trace logs for specific conversations/issues
4. **Persistence**: Logs are preserved per conversation for auditing
5. **Cleanup**: Can easily delete old conversation logs by removing the folder

## Configuration

The base work directory is configured in `config.yaml`:
```yaml
claude:
  work_dir: /tmp/claude-sessions
```

## Testing

Build verification passed:
```bash
cd /root/CodeBase/code/Kinetik/automation
go build -o /tmp/test-build ./cmd/webhook-server
# Success - no compilation errors
```

## Database Integration

The conversation ID used for folder routing is the same UUID stored in the PostgreSQL `conversations` table:
- Field: `conversations.id`
- Type: `UUID` (Primary Key)
- Also used for: Tracking conversation state, storing context, linking issues/PRs

## Backward Compatibility

This change is fully backward compatible:
- Existing conversations in the database will work correctly
- No database migration required
- Configuration remains unchanged
- CLI client already supported workDir parameter

## Example Session Flow

1. GitHub webhook receives new issue event
2. `HandleNewIssue()` creates conversation with ID: `a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d`
3. Work directory created: `/tmp/claude-sessions/a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d/`
4. Claude CLI executes, logs written to:
   - `/tmp/claude-sessions/a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d/stdout.log`
   - `/tmp/claude-sessions/a7f3c92b-4e1d-4f2a-9c8e-5d6f1b3a2c4d/stderr.log`
5. Same conversation ID used for all subsequent interactions with that issue

## Related Files

- `/root/CodeBase/code/Kinetik/automation/internal/workflow/orchestrator.go` - Orchestrator (updated)
- `/root/CodeBase/code/Kinetik/automation/internal/claude/cli_client.go` - CLI client (no changes needed)
- `/root/CodeBase/code/Kinetik/automation/internal/repository/conversation_repo.go` - Conversation model
- `/root/CodeBase/code/Kinetik/automation/config/config.example.yaml` - Configuration example
