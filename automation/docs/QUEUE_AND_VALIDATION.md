# Issue Queue and Validation System

This document describes the new issue queue and validation features added to the Kinetik automation system.

## Features

### 1. Issue Validation

When a new issue is created, the bot automatically validates that it contains required information:

- **Commit hash** (e.g., `abc123def`, `commit: main`, SHA patterns)
- **Version number** (e.g., `v1.2.3`, `version 2.0`, `release 1.5`)

#### Behavior

**If validation fails:**
1. Bot posts a comment requesting the missing information
2. Adds `needs-info` label to the issue
3. Does NOT start processing the issue

**If validation passes:**
1. Issue is added to the processing queue
2. Bot proceeds with analysis and implementation

#### Example Comment

```markdown
👋 Thanks for opening this issue!

To help me better understand and fix this issue, could you please provide:

- **Commit hash** or **version number** where you encountered this issue
  - Example: `commit: abc123def` or `version: v1.2.3`

This information helps me reproduce the issue in the correct codebase state.

Once you've added this info, mention me with `@KinetikBot` and I'll get started!
```

### 2. Issue Queue System

Multiple issues are now processed sequentially, one at a time, each on its own dedicated branch.

#### How It Works

1. **Enqueue**: Valid issues are added to a FIFO queue
2. **Sequential Processing**: Issues are processed one by one (no parallel processing)
3. **Branch Creation**: Each issue gets its own branch named `issue-{number}-{timestamp}`
4. **Automatic Processing**: Queue processes automatically in the background

#### Queue Status

Users can see queue status when they mention the bot:

```markdown
Thanks for the info! I've added this to my queue. 
Currently processing 2 issue(s). 
I'll work on this on branch `issue-123-1709123456`.
```

#### Branch Naming

Branches are automatically created with the format:
```
issue-{issue_number}-{unix_timestamp}
```

Example: `issue-42-1709123456`

This ensures:
- Unique branch names even for reopened issues
- Easy identification of which issue a branch addresses
- Chronological ordering

### 3. Updated Workflow

#### New Issue Flow

```
1. Issue Created
   ↓
2. Validate (has commit/version?)
   ↓
   ├─ NO → Post comment requesting info
   │        Add "needs-info" label
   │        Wait for user response
   │
   └─ YES → Add to queue
            ↓
            Create branch
            ↓
            Run analysis
            ↓
            Wait for approval
            ↓
            Implement fix
```

#### User Provides Missing Info

```
1. User adds commit/version to issue
   ↓
2. User mentions @KinetikBot
   ↓
3. Bot re-validates issue
   ↓
   ├─ Still invalid → Remind user
   │
   └─ Valid → Add to queue
              Post acknowledgment with queue status
```

## Configuration

No additional configuration needed. The system uses existing settings:

```yaml
github:
  bot_username: KinetikBot
  personal_access_token: ${GITHUB_PERSONAL_ACCESS_TOKEN}
```

## API Changes

### Orchestrator

**Updated Constructor:**
```go
func NewOrchestrator(
    cfg *config.Config,
    claudeClient *claude.CLIClient,
    promptBuilder *claude.PromptBuilder,
    contextManager *contextmgr.Manager,
    githubClient *github.Client,  // NEW
) *Orchestrator
```

**New Methods:**
- `processQueuedIssue(ctx, queued)` - Internal queue processor

### New Components

**IssueValidator:**
```go
validator := NewIssueValidator()
result := validator.ValidateIssue(issue)
comment := validator.BuildRequestInfoComment(result)
```

**IssueQueue:**
```go
queue := NewIssueQueue(processorFunc)
branchName := queue.Enqueue(event)
length := queue.GetQueueLength()
isProcessing := queue.IsProcessing()
```

**GitHub Client:**
```go
client := github.NewClient(cfg)
client.CreateIssueComment(owner, repo, issueNum, body)
client.CreateBranch(owner, repo, branchName, baseBranch)
client.AddIssueLabel(owner, repo, issueNum, label)
```

## Testing

Run tests:
```bash
cd automation
go test ./internal/workflow/... -v
```

Test coverage includes:
- Issue validation with various input patterns
- Queue enqueue/dequeue operations
- Sequential processing order
- Queue status tracking

## Monitoring

Check logs for queue activity:

```bash
# Docker
docker-compose logs -f webhook-server | grep -i queue

# Systemd
sudo journalctl -u kinetik-automation -f | grep -i queue
```

Log messages include:
- `Issue queued for processing` - Issue added to queue
- `Processing queued issue` - Issue being processed
- `Completed processing queued issue` - Issue finished
- `Queue empty, stopping processor` - All issues processed

## Future Enhancements

Potential improvements:
- Configurable validation rules
- Priority queue (urgent issues first)
- Queue size limits
- Timeout handling for stuck issues
- Queue status API endpoint
- Dashboard showing queue state
