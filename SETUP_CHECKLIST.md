# Kinetik Automation System - Setup Checklist

Use this checklist to track your progress through the setup process.

## Phase 1: GitHub Bot Account Setup

- [ ] Create dedicated GitHub bot account (e.g., `kinetik-bot`)
- [ ] Add bot as collaborator to all three repositories:
  - [ ] TomasBack2Future/Kinetik
  - [ ] TomasBack2Future/KinetikServer
  - [ ] TomasBack2Future/kinetik_agent
- [ ] Accept all collaboration invitations from bot account

## Phase 2: Generate Personal Access Token

- [ ] Log in to bot GitHub account
- [ ] Navigate to Settings → Developer settings → Personal access tokens → Fine-grained tokens
- [ ] Generate new token with:
  - [ ] Name: "Kinetik Automation Bot"
  - [ ] Expiration: 90 days
  - [ ] Repository access: Select all three repositories
  - [ ] Permissions:
    - [ ] Contents: Read and write
    - [ ] Issues: Read and write
    - [ ] Pull requests: Read and write
    - [ ] Workflows: Read and write
    - [ ] Metadata: Read-only (auto-selected)
- [ ] Copy token and save securely
- [ ] Export environment variable:
  ```bash
  export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxx"
  echo 'export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxx"' >> ~/.bashrc
  ```

## Phase 3: Create GitHub App

**Location**: https://github.com/settings/apps/new

**Detailed Guide**: See [docs/github-app-installation.md](../docs/github-app-installation.md) for step-by-step instructions with troubleshooting

- [ ] Go to https://github.com/settings/apps/new
- [ ] Fill in basic information:
  - [ ] App name: "Kinetik Automation Bot" (or add suffix if taken)
  - [ ] Homepage URL: https://github.com/TomasBack2Future/Kinetik
  - [ ] Webhook URL: Leave blank for now
- [ ] Generate webhook secret:
  ```bash
  openssl rand -hex 32
  ```
- [ ] Copy webhook secret and save it
- [ ] Configure permissions:
  - [ ] Contents: Read-only
  - [ ] Issues: Read & write
  - [ ] Pull requests: Read & write
  - [ ] Metadata: Read-only
- [ ] Subscribe to events:
  - [ ] Issues
  - [ ] Issue comment
  - [ ] Pull request
  - [ ] Pull request review
  - [ ] Pull request review comment
- [ ] Set "Where can this GitHub App be installed?": Only on this account
- [ ] Create GitHub App
- [ ] Install app on all three repositories:
  - After creating the app, you'll be on the app settings page
  - Click "Install App" in the left sidebar
  - Click the "Install" button next to your account name
  - On the installation page, select "Only select repositories"
  - Use the dropdown to select:
    - [ ] TomasBack2Future/Kinetik
    - [ ] TomasBack2Future/KinetikServer
    - [ ] TomasBack2Future/kinetik_agent
  - Click "Install"
  - You'll be redirected back to GitHub
- [ ] Verify installation:
  - Go to each repository's Settings → Integrations → GitHub Apps
  - You should see "Kinetik Automation Bot" (or your app name) listed
- [ ] Export webhook secret:
  ```bash
  export GITHUB_WEBHOOK_SECRET="67f9f0d9dc944c1a10d5a189af2d7d34c1cefdf182a14a90c2780330110422a1"
  echo 'export GITHUB_WEBHOOK_SECRET="67f9f0d9dc944c1a10d5a189af2d7d34c1cefdf182a14a90c2780330110422a1"' >> ~/.bashrc
  ```

## Phase 4: Database Setup

- [ ] Connect to PostgreSQL:
  ```bash
  sudo -u postgres psql
  ```
- [ ] Create database:
  ```sql
  CREATE DATABASE kinetik_automation;
  ```
- [ ] Create user:
  ```sql
  CREATE USER kinetik_automation WITH PASSWORD 'your_secure_password';
  ```
- [ ] Grant privileges:
  ```sql
  GRANT ALL PRIVILEGES ON DATABASE kinetik_automation TO kinetik_automation;
  ```
- [ ] Exit psql: `\q`
- [ ] Export database URL:
  ```bash
  export DATABASE_URL="postgresql://kinetik_automation:your_password@localhost:5432/kinetik_automation"
  echo 'export DATABASE_URL="postgresql://..."' >> ~/.bashrc
  ```

## Phase 5: Install Dependencies

- [ ] Verify Go is installed:
  ```bash
  go version  # Should be 1.22+
  ```
- [ ] Verify PostgreSQL is running:
  ```bash
  sudo systemctl status postgresql
  ```
- [ ] Verify Claude CLI is installed:
  ```bash
  /root/.nvm/versions/node/v22.19.0/bin/claude --version
  ```
- [ ] Install ngrok (for local testing):
  ```bash
  # Follow instructions at https://ngrok.com/download
  ```

## Phase 6: Configure Application

- [ ] Navigate to automation directory:
  ```bash
  cd /root/CodeBase/code/Kinetik/automation
  ```
- [ ] Install Go dependencies:
  ```bash
  go mod download
  ```
- [ ] Copy example config:
  ```bash
  cp config/config.example.yaml config/config.yaml
  ```
- [ ] Edit config if needed (environment variables will override)

## Phase 7: Run Database Migrations

- [ ] Run migrations:
  ```bash
  make migrate-up
  ```
- [ ] Verify tables created:
  ```bash
  psql $DATABASE_URL -c "\dt"
  # Should show: conversations, events
  ```

## Phase 8: Test Locally with ngrok

- [ ] Start ngrok:
  ```bash
  ngrok http 8080
  ```
- [ ] Note ngrok URL (e.g., `https://abc123.ngrok.io`)
- [ ] Update GitHub App webhook URL:
  - [ ] Go to GitHub App settings
  - [ ] Set Webhook URL to: `https://abc123.ngrok.io/github/webhook`
  - [ ] Save changes
- [ ] Start webhook server:
  ```bash
  cd /root/CodeBase/code/Kinetik/automation
  make run
  ```
- [ ] Test health endpoint:
  ```bash
  curl http://localhost:8080/health
  # Should return: OK
  ```

## Phase 9: End-to-End Testing

- [ ] Create test issue in Kinetik repository
- [ ] Monitor webhook server logs for:
  - [ ] Webhook received
  - [ ] Issue analysis started
  - [ ] Claude execution completed
- [ ] Verify bot posted analysis comment on issue
- [ ] Verify "awaiting-approval" label added
- [ ] Comment "approved" on the issue
- [ ] Monitor logs for:
  - [ ] Approval detected
  - [ ] Implementation started
  - [ ] PR creation
- [ ] Verify PR was created and linked to issue
- [ ] Comment on PR to test review handling
- [ ] Verify bot responds to PR comments

## Phase 10: Production Deployment

### Option A: Docker Compose

- [ ] Update docker-compose.yml with production settings
- [ ] Build image:
  ```bash
  make docker-build
  ```
- [ ] Start services:
  ```bash
  docker-compose up -d
  ```
- [ ] Check logs:
  ```bash
  make docker-logs
  ```

### Option B: Systemd Service

- [ ] Build binary:
  ```bash
  make build
  ```
- [ ] Copy files to production location:
  ```bash
  sudo mkdir -p /opt/kinetik-automation
  sudo cp bin/webhook-server /opt/kinetik-automation/
  sudo cp -r config /opt/kinetik-automation/
  sudo cp -r migrations /opt/kinetik-automation/
  ```
- [ ] Create systemd service file (see automation/README.md)
- [ ] Enable and start service:
  ```bash
  sudo systemctl enable kinetik-automation
  sudo systemctl start kinetik-automation
  ```
- [ ] Check status:
  ```bash
  sudo systemctl status kinetik-automation
  ```

## Phase 11: Update Production Webhook URL

- [ ] Set up domain and SSL certificate
- [ ] Update GitHub App webhook URL to production domain:
  - [ ] `https://your-domain.com/github/webhook`
- [ ] Test webhook delivery from GitHub App settings

## Phase 12: Monitoring and Verification

- [ ] Set up log monitoring:
  ```bash
  # Docker
  docker-compose logs -f webhook-server

  # Systemd
  sudo journalctl -u kinetik-automation -f
  ```
- [ ] Create test issue in production
- [ ] Verify bot responds correctly
- [ ] Test full workflow: issue → analysis → approval → PR
- [ ] Monitor for errors or issues

## Phase 13: Security Hardening

- [ ] Verify webhook secret is strong (32+ characters)
- [ ] Verify PAT has minimal required scopes
- [ ] Set calendar reminder to rotate PAT before 90 days
- [ ] Verify database password is strong
- [ ] Configure firewall rules if needed
- [ ] Verify HTTPS is enforced
- [ ] Verify service runs as non-root user
- [ ] Audit environment variables are not logged

## Phase 14: Documentation and Handoff

- [ ] Document any custom configuration
- [ ] Create runbook for common issues
- [ ] Set up alerting (optional)
- [ ] Configure backup for PostgreSQL
- [ ] Document PAT rotation procedure
- [ ] Share access with team members if needed

## Troubleshooting Quick Reference

### Webhook Not Received
1. Check GitHub App webhook delivery logs
2. Verify webhook secret matches
3. Test with curl: `curl https://your-domain.com/health`
4. Check firewall rules

### Bot Not Responding
1. Check Claude CLI: `/root/.nvm/versions/node/v22.19.0/bin/claude --version`
2. Verify PAT: `curl -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" https://api.github.com/user`
3. Check service logs for errors
4. Verify database connection

### Database Issues
1. Test connection: `psql $DATABASE_URL -c "SELECT 1;"`
2. Verify migrations: `psql $DATABASE_URL -c "\dt"`
3. Check PostgreSQL status: `sudo systemctl status postgresql`

## Success Criteria

All items below should be ✅:

- [ ] Bot account created and configured
- [ ] GitHub PAT generated and working
- [ ] GitHub App created and installed
- [ ] Database created and migrated
- [ ] Application builds successfully
- [ ] Local testing with ngrok passed
- [ ] Production deployment completed
- [ ] End-to-end workflow tested
- [ ] Bot responds to issues within 2 minutes
- [ ] PRs created after approval
- [ ] Context maintained across interactions
- [ ] Logs show no errors
- [ ] Health check returns OK

## Completion Date

**Setup Started**: _______________
**Setup Completed**: _______________
**Deployed By**: _______________

## Notes

Use this section to document any issues encountered or custom configuration applied:

---
