# Automation Configuration Guide

## Overview

The Kinetik GitHub automation system uses **environment variables** and **configuration files** to avoid hardcoding secrets and paths.

## Required Environment Variables

Set these in your shell environment or `.env` file:

```bash
# GitHub Personal Access Token (for API authentication and MCP server)
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# GitHub Webhook Secret (for webhook signature verification)
export GITHUB_WEBHOOK_SECRET="your-webhook-secret-here"

# Optional: Database credentials (if not using config.yaml defaults)
export DB_HOST="your-db-host"
export DB_USER="your-db-user"
export DB_PASSWORD="your-db-password"
export DB_NAME="your-db-name"
```

## Configuration File

Edit `automation/config/config.yaml`:

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DB_HOST:-35.194.125.198}
  port: 5432
  name: ${DB_NAME:-kinetik_automation}
  user: ${DB_USER:-kinetik_automation}
  password: ${DB_PASSWORD:-your-password}
  ssl_mode: require

github:
  # Loaded from environment variable
  webhook_secret: ${GITHUB_WEBHOOK_SECRET}
  personal_access_token: ${GITHUB_PERSONAL_ACCESS_TOKEN}
  bot_username: KinetikBot
  allowed_repos:
    - TomasBack2Future/Kinetik
    - TomasBack2Future/KinetikServer
    - TomasBack2Future/kinetik_agent

claude:
  cli_path: /root/.nvm/versions/node/v22.19.0/bin/claude
  work_dir: /tmp/claude-sessions
  # Repository root where project-scoped MCP servers are configured
  # Change this if your repo is in a different location
  repo_root: /root/CodeBase/code/Kinetik
  timeout: 300s
  # Maximum number of retries if Claude CLI fails or is killed (default: 2)
  max_retries: 2

workflow:
  approval_keywords:
    - approved
    - lgtm
    - proceed
    - go ahead

logging:
  level: info
  enabled: true
  file: logs/webhook-server.log
```

## Claude CLI Retry Behavior

The automation system includes retry logic for Claude CLI execution failures:

- **Automatic Retries**: If the Claude CLI process is killed, exits with an error, or encounters transient failures, it will automatically retry
- **Configurable Retries**: Set `max_retries` in `config.yaml` (default: 2)
- **Exponential Backoff**: Retries use exponential backoff delays (2s, 4s, 8s...)
- **Smart Retry Logic**: Only retries on retryable errors (killed, timeout, connection issues)
- **Non-retryable Errors**: Syntax errors, authentication failures, and other permanent errors won't trigger retries

**Example scenarios that trigger retries:**
- Process killed by OOM killer
- Network timeout or connection refused
- Context deadline exceeded
- Signal interruptions (SIGTERM, SIGKILL)

**Example scenarios that don't trigger retries:**
- Invalid syntax or command errors
- Authentication failures
- Permission denied errors

## GitHub CLI Integration

The system uses the GitHub CLI (`gh`) for all GitHub API interactions. The `gh` CLI is:
- Installed automatically in the Docker container via the `github-cli` Alpine package
- Authenticated automatically using the `GITHUB_TOKEN` environment variable
- More reliable than MCP tools as it makes direct GitHub API calls
- Easier to debug with visible command output

**Authentication:** The `gh` CLI automatically uses `GITHUB_TOKEN` or `GH_TOKEN` environment variables. The automation system passes `GITHUB_PERSONAL_ACCESS_TOKEN` as `GITHUB_TOKEN` to Claude's subprocess, which then uses it for all `gh` commands.

No additional setup is required - authentication is handled automatically via the environment variable configured in `config.yaml`.

## Verify Configuration

### 1. Check environment variables
```bash
env | grep GITHUB
```

Should show:
```
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_...
GITHUB_WEBHOOK_SECRET=...
```

### 2. Test gh CLI authentication
```bash
export GITHUB_TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN"
gh auth status
```

Should show:
```
github.com
  ✓ Logged in to github.com as YourUsername (keyring)
  ✓ Git operations for github.com configured to use https protocol.
  ✓ Token: ghp_************************************
```

### 3. Test gh CLI commands
```bash
export GITHUB_TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN"
gh repo view TomasBack2Future/Kinetik
```

Should display repository information without errors.

## Deployment

### Using Environment File

Create `.env` in automation directory (git-ignored):
```bash
# automation/.env
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
GITHUB_WEBHOOK_SECRET=your-webhook-secret-here
```

Load before running:
```bash
cd automation
source .env
./bin/webhook-server
```

### Using Systemd

Create service file `/etc/systemd/system/kinetik-automation.service`:
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

The `.env` file is loaded automatically and not committed to version control.

### Using Docker

Pass as environment variables:
```bash
docker run -d \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_PERSONAL_ACCESS_TOKEN \
  -e GITHUB_WEBHOOK_SECRET=$GITHUB_WEBHOOK_SECRET \
  -v /root/CodeBase/code/Kinetik:/app \
  kinetik-automation
```

## Security Best Practices

1. **Never commit tokens** - Add `.env` to `.gitignore`
2. **Use environment variables** - Don't hardcode in config files
3. **Restrict permissions** - `.env` file should be `chmod 600`
4. **Rotate tokens regularly** - Update GitHub PAT every 90 days
5. **Use fine-grained PATs** - Limit scope to only required permissions

## Changing Configuration

### To change repository path:
1. Edit `automation/config/config.yaml` → `claude.repo_root`
2. Rebuild: `make build`
3. Restart server

### To change GitHub token:
1. Update environment variable: `export GITHUB_PERSONAL_ACCESS_TOKEN="new-token"`
2. No rebuild needed - loaded at runtime

### To change webhook secret:
1. Update environment variable: `export GITHUB_WEBHOOK_SECRET="new-secret"`
2. Restart server to reload config

### To change retry behavior:
1. Edit `automation/config/config.yaml` → `claude.max_retries`
2. Set to 0 to disable retries, or any positive number for retry count
3. Rebuild: `make build`
4. Restart server
