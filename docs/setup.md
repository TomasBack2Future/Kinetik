# Kinetik Automation System Setup Guide

This guide walks you through setting up the GitHub automation system for the Kinetik project.

## Prerequisites

Before starting, ensure you have:

- [ ] Bot GitHub account created
- [ ] Access to target repositories (Kinetik, KinetikServer, kinetik_agent)
- [ ] PostgreSQL 14+ installed
- [ ] Go 1.22+ installed
- [ ] Node.js and npm installed (for Claude CLI)
- [ ] Claude CLI installed globally
- [ ] ngrok installed (for local testing)

## Step 1: Generate Personal Access Token for Bot Account

1. Log in to your bot GitHub account (e.g., `kinetik-bot`)
2. Navigate to: **Settings** → **Developer settings** → **Personal access tokens** → **Fine-grained tokens**
3. Click **"Generate new token"**
4. Configure token:
   - **Token name**: `Kinetik Automation Bot`
   - **Expiration**: 90 days
   - **Repository access**: Only select repositories
     - Select: `TomasBack2Future/Kinetik`
     - Select: `TomasBack2Future/KinetikServer`
     - Select: `TomasBack2Future/kinetik_agent`
   - **Repository permissions**:
     - Contents: **Read and write**
     - Issues: **Read and write**
     - Pull requests: **Read and write**
     - Workflows: **Read and write**
     - Metadata: **Read-only** (auto-selected)
5. Click **"Generate token"** and **SAVE IT SECURELY**
6. Export on your server:
   ```bash
   export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxx"
   echo 'export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxx"' >> ~/.bashrc
   ```

## Step 2: Create GitHub App for Webhooks

1. Go to: https://github.com/settings/apps/new
2. Fill in the form:
   - **GitHub App name**: `Kinetik Automation Bot` (must be globally unique)
   - **Homepage URL**: `https://github.com/TomasBack2Future/Kinetik`
   - **Webhook URL**: Leave blank for now (will configure later)
   - **Webhook secret**: Generate a strong random string:
     ```bash
     openssl rand -hex 32
     ```
     Save this as `GITHUB_WEBHOOK_SECRET`
   - **Repository permissions**:
     - Contents: Read-only
     - Issues: Read & write
     - Pull requests: Read & write
     - Metadata: Read-only
   - **Subscribe to events**:
     - ✅ Issues
     - ✅ Issue comment
     - ✅ Pull request
     - ✅ Pull request review
     - ✅ Pull request review comment
   - **Where can this GitHub App be installed?**: Only on this account
3. Click **"Create GitHub App"**
4. After creation, click **"Install App"** and select:
   - ✅ Kinetik
   - ✅ KinetikServer
   - ✅ kinetik_agent
5. Select **"Only select repositories"** and choose all three

**Save environment variable:**
```bash
export GITHUB_WEBHOOK_SECRET="your-secret-from-step-2"
echo 'export GITHUB_WEBHOOK_SECRET="your-secret"' >> ~/.bashrc
```

## Step 3: Set Up Database

Create a separate database for the automation service:

```bash
# Connect to PostgreSQL as superuser
sudo -u postgres psql

# Create database and user
CREATE DATABASE kinetik_automation;
CREATE USER kinetik_automation WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE kinetik_automation TO kinetik_automation;

# Exit psql
\q
```

**Save database credentials:**
```bash
export DATABASE_URL="postgresql://kinetik_automation:your_secure_password@localhost:5432/kinetik_automation"
echo 'export DATABASE_URL="postgresql://..."' >> ~/.bashrc
```

## Step 4: Add Bot as Collaborator

1. Go to each repository settings:
   - https://github.com/TomasBack2Future/Kinetik/settings/access
   - https://github.com/TomasBack2Future/KinetikServer/settings/access
   - https://github.com/TomasBack2Future/kinetik_agent/settings/access
2. Click **"Invite a collaborator"**
3. Enter bot account username
4. Select **"Write"** permission
5. Log in as bot account and accept all invitations

## Step 5: Install and Configure Application

### Install Dependencies

```bash
cd /root/CodeBase/code/Kinetik/automation
go mod download
```

### Run Database Migrations

```bash
make migrate-up
```

### Configure Application

Copy example config:

```bash
cp config/config.example.yaml config/config.yaml
```

Edit `config/config.yaml` as needed (environment variables will override).

## Step 6: Test Locally with ngrok

### Start ngrok

```bash
ngrok http 8080
```

Note the HTTPS URL (e.g., `https://abc123.ngrok.io`)

### Update GitHub App Webhook URL

1. Go to your GitHub App settings
2. Update **Webhook URL** to: `https://abc123.ngrok.io/github/webhook`
3. Save changes

### Start the Webhook Server

```bash
cd /root/CodeBase/code/Kinetik/automation
make run
```

### Test the Integration

1. Create a test issue in one of the repositories
2. Check webhook server logs:
   ```bash
   # Should see: "Received GitHub webhook"
   ```
3. Wait for Claude to analyze and post a comment
4. Verify the comment appears on the issue

## Step 7: Deploy to Production

### Option A: Using Docker Compose

```bash
cd /root/CodeBase/code/Kinetik/automation
docker-compose up -d
```

### Option B: Using Systemd

1. Build the binary:
   ```bash
   make build
   ```

2. Copy files to production location:
   ```bash
   sudo mkdir -p /opt/kinetik-automation
   sudo cp bin/webhook-server /opt/kinetik-automation/
   sudo cp -r config /opt/kinetik-automation/
   sudo cp -r migrations /opt/kinetik-automation/
   ```

3. Create systemd service (see README for service file)

4. Start service:
   ```bash
   sudo systemctl enable kinetik-automation
   sudo systemctl start kinetik-automation
   ```

### Update GitHub App Webhook URL

Update the webhook URL to your production domain:
`https://your-domain.com/github/webhook`

## Step 8: Verify Everything Works

### Check Health Endpoint

```bash
curl https://your-domain.com/health
# Should return: OK
```

### Create Test Issue

1. Create a new issue in the Kinetik repository
2. Monitor logs:
   ```bash
   # Docker
   docker-compose logs -f webhook-server

   # Systemd
   sudo journalctl -u kinetik-automation -f
   ```
3. Verify bot posts analysis comment within 2 minutes
4. Comment "approved" on the issue
5. Verify bot creates a pull request

## Troubleshooting

### Webhook Not Received

- Check GitHub App webhook delivery logs
- Verify webhook secret matches
- Ensure firewall allows incoming connections
- Test with curl:
  ```bash
  curl -X POST https://your-domain.com/health
  ```

### Bot Not Responding

- Check Claude CLI is accessible:
  ```bash
  /root/.nvm/versions/node/v22.19.0/bin/claude --version
  ```
- Verify GitHub PAT is valid:
  ```bash
  curl -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" https://api.github.com/user
  ```
- Check logs for errors

### Database Connection Issues

- Test connection:
  ```bash
  psql $DATABASE_URL -c "SELECT 1;"
  ```
- Verify migrations ran:
  ```bash
  psql $DATABASE_URL -c "\dt"
  # Should show: conversations, events
  ```

## Security Checklist

- [ ] GitHub webhook secret is strong (32+ characters)
- [ ] Bot PAT has minimal required scopes
- [ ] PAT expires in 90 days (set reminder to rotate)
- [ ] Database password is strong
- [ ] Firewall rules restrict access to webhook endpoint
- [ ] HTTPS is enforced (no HTTP)
- [ ] Service runs as non-root user
- [ ] Environment variables not logged or exposed

## Next Steps

- Set up monitoring and alerting
- Configure log aggregation
- Set up automated backups for PostgreSQL
- Create runbook for common issues
- Schedule PAT rotation (before 90 days)

## Support

For issues or questions:
- Check logs: `make docker-logs` or `journalctl -u kinetik-automation`
- Review [Troubleshooting section](README.md#troubleshooting)
- Create issue in repository
