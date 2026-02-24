# Kinetik Automation System Architecture

## Overview

The Kinetik GitHub Automation System is a self-hosted webhook service that monitors GitHub activity across the Kinetik project repositories and uses Claude Code with GitHub MCP to automatically analyze issues, propose solutions, and create pull requests.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          GitHub                                 │
│  ┌─────────────┐  ┌──────────────────┐  ┌─────────────────┐   │
│  │   Kinetik   │  │  KinetikServer   │  │  kinetik_agent  │   │
│  └──────┬──────┘  └────────┬─────────┘  └────────┬────────┘   │
│         │                  │                      │             │
│         └──────────────────┴──────────────────────┘             │
│                            │ Webhooks                            │
└────────────────────────────┼─────────────────────────────────────┘
                             │
                             ▼
          ┌──────────────────────────────────────┐
          │     Webhook Server (Go + Chi)        │
          │  ┌────────────────────────────────┐  │
          │  │  Middleware Layer              │  │
          │  │  • Signature Validation        │  │
          │  │  • Request Logging             │  │
          │  │  • Recovery                    │  │
          │  └────────────────────────────────┘  │
          │  ┌────────────────────────────────┐  │
          │  │  Event Router                  │  │
          │  │  • Issues                      │  │
          │  │  • Issue Comments              │  │
          │  │  • Pull Requests               │  │
          │  │  • PR Reviews                  │  │
          │  └────────────────────────────────┘  │
          └──────────────┬───────────────────────┘
                         │
                         ▼
          ┌──────────────────────────────────────┐
          │     Workflow Orchestrator            │
          │  ┌────────────────────────────────┐  │
          │  │  • Issue Analysis Workflow     │  │
          │  │  • Approval Workflow           │  │
          │  │  • Implementation Workflow     │  │
          │  │  • PR Review Workflow          │  │
          │  └────────────────────────────────┘  │
          └──────┬───────────────────────┬────────┘
                 │                       │
                 ▼                       ▼
    ┌────────────────────┐   ┌────────────────────┐
    │  Context Manager   │   │   Claude Client    │
    │  ┌──────────────┐  │   │  ┌──────────────┐  │
    │  │  PostgreSQL  │  │   │  │  CLI Wrapper │  │
    │  │  Storage     │  │   │  └──────────────┘  │
    │  └──────────────┘  │   │  ┌──────────────┐  │
    │  • Conversations   │   │  │   Prompt     │  │
    │  • State Machine   │   │  │   Builder    │  │
    │  • History         │   │  └──────────────┘  │
    └────────────────────┘   └──────────┬─────────┘
                                        │
                                        ▼
                             ┌────────────────────┐
                             │   Claude Code CLI  │
                             │  with GitHub MCP   │
                             │  ┌──────────────┐  │
                             │  │  Code Tools  │  │
                             │  │  • Read      │  │
                             │  │  • Write     │  │
                             │  │  • Edit      │  │
                             │  │  • Glob/Grep │  │
                             │  └──────────────┘  │
                             │  ┌──────────────┐  │
                             │  │ GitHub MCP   │  │
                             │  │  • Issues    │  │
                             │  │  • Comments  │  │
                             │  │  • PRs       │  │
                             │  │  • Labels    │  │
                             │  └──────────────┘  │
                             └────────────────────┘
```

## Components

### 1. Webhook Server

**Technology**: Go 1.22 + Chi v5

**Responsibilities**:
- Receive and validate GitHub webhooks
- Route events to appropriate handlers
- Manage HTTP middleware (auth, logging, recovery)

**Key Features**:
- HMAC-SHA256 signature validation
- Repository whitelist filtering
- Structured JSON logging
- Graceful shutdown

**Endpoints**:
- `GET /health` - Health check
- `POST /github/webhook` - GitHub webhook receiver

### 2. Event Handlers

**Location**: `internal/handlers/`

**Components**:
- `webhook_handler.go` - Main router
- `types.go` - GitHub event structures

**Supported Events**:
- `issues` (action: opened)
- `issue_comment` (action: created)
- `pull_request` (action: opened, synchronize)
- `pull_request_review` (action: submitted)
- `pull_request_review_comment` (action: created)

### 3. Workflow Orchestrator

**Location**: `internal/workflow/orchestrator.go`

**Responsibilities**:
- Coordinate event processing
- Manage workflow state transitions
- Integrate context manager and Claude client

**Workflows**:

#### Issue Analysis Workflow
1. Receive new issue event
2. Load or create conversation context
3. Build analysis prompt
4. Execute Claude with GitHub MCP
5. Claude analyzes code and posts comment
6. Update state to "pending_approval"

#### Approval Workflow
1. Detect approval keyword in comment
2. Verify conversation state
3. Build implementation prompt
4. Execute Claude with previous session ID
5. Claude implements changes and creates PR
6. Update state to "completed"

#### Mention Workflow
1. Detect @bot mention
2. Load conversation context
3. Build contextual prompt
4. Execute Claude for response
5. Claude posts reply

#### PR Review Workflow
1. Receive PR review or comment
2. Load PR conversation context
3. Build review response prompt
4. Execute Claude to address feedback
5. Claude updates code and responds

### 4. Context Manager

**Location**: `internal/context/`

**Responsibilities**:
- Manage conversation state
- Store and retrieve context
- Build context strings for prompts

**State Machine**:
```
analyzing → pending_approval → executing → completed
    ↓                                          ↑
  failed ←──────────────────────────────────────┘
```

**Storage**:
- Primary: PostgreSQL
- Context stored as JSONB
- Indexed by repository and issue/PR number

### 5. Claude Integration

**Location**: `internal/claude/`

**Components**:
- `cli_client.go` - CLI wrapper and executor
- `prompt_builder.go` - Context-aware prompt generation

**Execution Flow**:
1. Build environment with GitHub PAT
2. Create work directory for session
3. Execute Claude CLI with prompt
4. Capture output and session ID
5. Parse results

**GitHub MCP Integration**:
- PAT passed via environment variable
- Claude automatically uses GitHub MCP tools
- No custom GitHub client needed
- Simplified architecture

### 6. Database Schema

**Tables**:

#### conversations
- `id` (UUID) - Primary key
- `repo_full_name` (VARCHAR) - Repository identifier
- `issue_number` (INTEGER) - Issue number
- `pr_number` (INTEGER) - PR number
- `state` (VARCHAR) - Workflow state
- `claude_session_id` (VARCHAR) - Session continuity
- `context` (JSONB) - Conversation data
- `created_at`, `updated_at` (TIMESTAMP)

#### events (audit log)
- `id` (SERIAL) - Primary key
- `conversation_id` (UUID) - Foreign key
- `event_type` (VARCHAR) - Event classification
- `event_data` (JSONB) - Event payload
- `created_at` (TIMESTAMP)

**Indexes**:
- `(repo_full_name, issue_number)`
- `(repo_full_name, pr_number)`
- `state`
- `created_at DESC`

## Data Flow

### New Issue Flow

```
1. GitHub → Webhook: POST /github/webhook
   {
     "action": "opened",
     "issue": { "number": 123, "title": "..." }
   }

2. Webhook → Handler: Validate & Route
   • Verify HMAC signature
   • Check repository whitelist
   • Route to issue handler

3. Handler → Orchestrator: HandleNewIssue()
   • Launch async goroutine
   • Return 202 Accepted immediately

4. Orchestrator → Context Manager: GetOrCreateConversation()
   • Query PostgreSQL
   • Return existing or create new

5. Orchestrator → Prompt Builder: BuildIssueAnalysisPrompt()
   • Construct prompt with context
   • Include issue details

6. Orchestrator → Claude Client: Execute()
   • Set GitHub PAT in environment
   • Create work directory
   • Run Claude CLI

7. Claude: Analyze & Act
   • Clone repository
   • Read relevant files
   • Analyze issue
   • Use github_create_issue_comment MCP tool
   • Use github_update_issue MCP tool (add label)

8. Claude Client → Orchestrator: ExecutionResult
   • Session ID
   • Output logs
   • Success status

9. Orchestrator → Context Manager: Update()
   • Store session ID
   • Update state to "pending_approval"
   • Add analysis to context
```

### Approval Flow

```
1. User comments: "approved"

2. GitHub → Webhook: POST /github/webhook
   {
     "action": "created",
     "comment": { "body": "approved" }
   }

3. Webhook → Handler: Detect approval keyword

4. Handler → Orchestrator: HandleIssueApproval()

5. Orchestrator → Context Manager: GetConversation()
   • Load existing conversation
   • Verify state is "pending_approval"

6. Orchestrator → Prompt Builder: BuildImplementationPrompt()
   • Include previous analysis from context
   • Use same session ID for continuity

7. Orchestrator → Claude Client: Execute()
   • Resume previous session
   • Maintain context

8. Claude: Implement & Create PR
   • Apply code changes
   • Use github_create_pull_request MCP tool
   • Use github_create_issue_comment to link PR

9. Orchestrator → Context Manager: Update()
   • Update state to "completed"
   • Store PR number
```

## Security Architecture

### Webhook Authentication
- HMAC-SHA256 signature verification
- Constant-time comparison prevents timing attacks
- Webhook secret stored in environment variable

### GitHub Authentication
- Bot account with dedicated PAT
- Fine-grained permissions (repo, workflow only)
- PAT expires in 90 days
- Passed to Claude via environment (never logged)

### Repository Access Control
- Whitelist of allowed repositories
- Events from other repos rejected (403)

### Database Security
- Separate database user with minimal privileges
- Connection credentials in environment
- SSL mode configurable

### Application Security
- Runs as non-root user
- Read-only filesystem (except work directory)
- Graceful shutdown handles SIGTERM/SIGINT
- Panic recovery middleware

## Scalability Considerations

### Current Architecture
- Single instance deployment
- Synchronous webhook handling with async processing
- PostgreSQL for persistence
- File-based Claude sessions

### Future Enhancements

**Horizontal Scaling**:
- Add Redis for distributed session storage
- Implement job queue (RabbitMQ/Redis)
- Load balancer for webhook endpoints

**Performance Optimization**:
- Cache frequently accessed conversations
- Batch database operations
- Implement webhook event queuing

**High Availability**:
- Multi-region deployment
- Database replication
- Health checks and auto-recovery

## Monitoring & Observability

### Logging
- Structured JSON logs (logrus)
- Request/response logging
- Error tracking with stack traces

### Metrics (Future)
- Webhook processing time
- Claude execution duration
- Success/failure rates
- Active conversations

### Alerting (Future)
- Failed webhook validations
- Claude execution failures
- Database connection issues
- High error rates

## Technology Stack Summary

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.22 |
| Web Framework | Chi | v5 |
| Database | PostgreSQL | 14+ |
| AI | Claude Code CLI | Latest |
| Logging | Logrus | v1.9 |
| Containerization | Docker | Latest |

## Key Design Decisions

### Why Go?
- Consistent with KinetikServer
- Excellent HTTP server performance
- Simple concurrency model (goroutines)
- Strong standard library

### Why Chi Router?
- Proven in KinetikServer
- Lightweight and fast
- Middleware composition
- Context-based routing

### Why GitHub MCP?
- Eliminates custom GitHub client (~500-1000 LOC)
- Claude handles API complexity
- Better error handling and retries
- Unified context management

### Why PostgreSQL?
- JSONB for flexible context storage
- ACID compliance for state management
- Mature and reliable
- Already used in ecosystem

### Why Async Processing?
- Immediate webhook acknowledgment
- Claude operations can be slow (minutes)
- Prevents GitHub webhook timeouts
- Better user experience

## Development Workflow

1. **Local Development**: Use ngrok + docker-compose
2. **Testing**: Unit tests + integration tests with test database
3. **CI/CD**: GitHub Actions for automated testing and building
4. **Deployment**: Docker Compose or systemd service
5. **Monitoring**: Logs via docker-compose logs or journalctl

## References

- [GitHub Webhooks Documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [Claude Code Documentation](https://docs.anthropic.com/claude/docs)
- [Go Chi Router](https://github.com/go-chi/chi)
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
