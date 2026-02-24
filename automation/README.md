# Kinetik GitHub Automation System

An intelligent GitHub automation system that monitors issues and pull requests across the Kinetik project repositories and uses Claude Code to automatically analyze, propose solutions, and create pull requests.

## Features

- **Automated Issue Analysis**: Analyzes new issues and provides detailed implementation plans
- **Approval-Based Workflow**: Waits for user approval before creating PRs
- **Context-Aware**: Maintains conversation history across multiple interactions
- **Multi-Repository Support**: Monitors main repo and all submodules
- **PR Review Assistance**: Responds to PR comments and reviews automatically
- **GitHub CLI Integration**: Uses gh CLI for reliable GitHub API interactions
- **Automatic Retry Logic**: Retries failed Claude CLI executions with exponential backoff (configurable, default: 2 retries)

## Architecture

```
GitHub Webhooks → Webhook Server (Go) → Context Manager (PostgreSQL)
                       ↓                         ↓
                  Event Router              Load History
                       ↓                         ↓
                Claude Code CLI ←──────── Build Prompt
                       ↓
                  gh CLI (Bash)
                       ↓
                GitHub API (comment/PR)
```

## Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Node.js (for Claude CLI)
- Docker & Docker Compose (for local development)
- GitHub Personal Access Token (bot account)
- GitHub App (for webhooks)

## Setup

### 1. Clone Repository

```bash
git clone https://github.com/TomasBack2Future/Kinetik.git
cd Kinetik/automation
```

### 2. Configure Environment

Create `.env` file:

```bash
export GITHUB_WEBHOOK_SECRET="your-webhook-secret"
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxx"
export DATABASE_URL="postgresql://kinetik_automation:password@localhost:5432/kinetik_automation"
```

### 3. Install Dependencies

```bash
make deps
```

### 4. Setup Database

```bash
# Create database
createdb kinetik_automation

# Run migrations
make migrate-up
```

### 5. Run Locally

```bash
# Start with Docker Compose
make dev

# Or run directly
make run
```

## Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

github:
  webhook_secret: ${GITHUB_WEBHOOK_SECRET}
  personal_access_token: ${GITHUB_PERSONAL_ACCESS_TOKEN}
  allowed_repos:
    - TomasBack2Future/Kinetik
    - TomasBack2Future/KinetikServer
    - TomasBack2Future/kinetik_agent

claude:
  cli_path: /root/.nvm/versions/node/v22.19.0/bin/claude
  work_dir: /tmp/claude-sessions
  timeout: 300s
```

## Usage

### Webhook Setup

1. Go to repository settings → Webhooks → Add webhook
2. Payload URL: `https://your-domain.com/github/webhook`
3. Content type: `application/json`
4. Secret: Your webhook secret
5. Events: Issues, Issue comments, Pull requests, Pull request reviews

### Triggering Automation

**New Issue**: Bot automatically analyzes and posts implementation plan

**Approve Implementation**: Comment "approved" or "lgtm" on the issue

**Ask Questions**: Mention `@kinetik-bot` in any comment

**PR Reviews**: Bot responds to comments and requested changes

## Development

### Run Tests

```bash
make test
```

### Build Docker Image

```bash
make docker-build
```

### View Logs

```bash
make docker-logs
```

### Clean Build

```bash
make clean
```

## Deployment

### Using Docker Compose

```bash
docker-compose up -d
```

### Using Systemd

Create `/etc/systemd/system/kinetik-automation.service`:

```ini
[Unit]
Description=Kinetik GitHub Automation Webhook Server
After=network.target postgresql.service

[Service]
Type=simple
User=kinetik
WorkingDirectory=/opt/kinetik-automation
Environment="CONFIG_PATH=/opt/kinetik-automation/config/config.yaml"
EnvironmentFile=/opt/kinetik-automation/.env
ExecStart=/opt/kinetik-automation/bin/webhook-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable kinetik-automation
sudo systemctl start kinetik-automation
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### View Logs

```bash
# Docker
docker-compose logs -f webhook-server

# Systemd
sudo journalctl -u kinetik-automation -f
```

## Troubleshooting

### Webhook Not Received

- Check webhook secret is correct
- Verify webhook URL is accessible
- Check firewall rules
- Review GitHub webhook delivery logs

### Claude Execution Fails

- Verify Claude CLI is installed: `which claude`
- Check GitHub PAT is valid
- Ensure work directory is writable
- Review Claude session logs in `/tmp/claude-sessions`

### Database Connection Issues

- Verify PostgreSQL is running
- Check database credentials
- Ensure migrations have been run
- Test connection: `psql $DATABASE_URL`

## Contributing

1. Create feature branch
2. Make changes
3. Run tests: `make test`
4. Submit pull request

## License

MIT License - see LICENSE file for details
