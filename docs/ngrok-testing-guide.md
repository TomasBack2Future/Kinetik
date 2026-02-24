# Testing Locally with ngrok

This guide shows you how to test the webhook server locally using ngrok.

## What is ngrok?

ngrok creates a secure tunnel from a public URL to your local server. This allows GitHub to send webhooks to your development machine.

```
GitHub → ngrok URL (https://abc123.ngrok.io) → ngrok tunnel → Your Local Server (http://localhost:8080)
```

## Prerequisites

- [ ] ngrok account created at https://ngrok.com
- [ ] ngrok installed on your server
- [ ] GitHub App created
- [ ] Database set up and migrations run
- [ ] Environment variables configured

## Step-by-Step Guide

### Step 1: Set Up Environment Variables

Make sure these are exported in your current terminal:

```bash
export GITHUB_WEBHOOK_SECRET="67f9f0d9dc944c1a10d5a189af2d7d34c1cefdf182a14a90c2780330110422a1"
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_your_token_here"
export DATABASE_URL="postgresql://kinetik_automation:Kinetik8estV0ip%21@35.194.125.198:5432/kinetik_automation?sslmode=require"

# Verify they're set
echo "Webhook Secret: ${GITHUB_WEBHOOK_SECRET:0:10}..."
echo "GitHub PAT: ${GITHUB_PERSONAL_ACCESS_TOKEN:0:10}..."
echo "Database: Set? $([ -n "$DATABASE_URL" ] && echo "Yes" || echo "No")"
```

### Step 2: Start Your Webhook Server

Open **Terminal 1** and start the webhook server:

```bash
cd /root/CodeBase/code/Kinetik/automation

# Make sure environment variables are loaded
source ~/.bashrc

# Start the server
go run ./cmd/webhook-server/main.go

# Or if you built it:
./bin/webhook-server
```

You should see output like:
```json
{"level":"info","message":"Starting Kinetik GitHub Automation Webhook Server","timestamp":"2026-02-10T..."}
{"level":"info","message":"Connected to database","timestamp":"2026-02-10T..."}
{"level":"info","message":"Server listening on :8080","timestamp":"2026-02-10T..."}
```

**✅ Keep this terminal running!**

### Step 3: Start ngrok

Open **Terminal 2** (new terminal window/tab):

```bash
# Start ngrok on port 8080 (same port as your webhook server)
ngrok http 8080
```

You'll see output like:
```
ngrok

Session Status                online
Account                       your@email.com (Plan: Free)
Version                       3.x.x
Region                        United States (us)
Latency                       -
Web Interface                 http://127.0.0.1:4040
Forwarding                    https://abc123def456.ngrok-free.app -> http://localhost:8080

Connections                   ttl     opn     rt1     rt5     p50     p90
                              0       0       0.00    0.00    0.00    0.00
```

**Important**: Copy the HTTPS URL from the "Forwarding" line:
```
https://abc123def456.ngrok-free.app
```

**✅ Keep this terminal running too!**

### Step 4: Update GitHub App Webhook URL

1. Go to your GitHub App settings:
   - Navigate to: https://github.com/settings/apps
   - Click on your app name (e.g., "Kinetik Automation Bot")

2. Scroll down to the **"Webhook"** section

3. Update the **"Webhook URL"** field:
   ```
   https://abc123def456.ngrok-free.app/github/webhook
   ```
   ⚠️ **Important**: Add `/github/webhook` to the end!

4. Make sure **"Webhook secret"** is still set (it should be)

5. Check **"Active"** checkbox is enabled

6. Click **"Save changes"** at the bottom

### Step 5: Test the Connection

#### Option A: Test via GitHub

1. Go to your GitHub App settings
2. Scroll to bottom → Click **"Advanced"** tab
3. Scroll to **"Recent Deliveries"** section
4. Click **"Redeliver"** on any existing delivery (if any)
5. Check if it succeeds

#### Option B: Test with curl

In **Terminal 3**:

```bash
# Test health endpoint
curl https://abc123def456.ngrok-free.app/health

# Should return: OK
```

### Step 6: Create a Test Issue

1. Go to one of your repositories:
   - https://github.com/TomasBack2Future/Kinetik/issues

2. Click **"New issue"**

3. Create a simple test issue:
   - **Title**: "Test webhook integration"
   - **Body**: "This is a test issue to verify the automation bot is working."

4. Click **"Submit new issue"**

### Step 7: Monitor the Logs

Switch back to **Terminal 1** (where your webhook server is running)

You should see logs appearing:

```json
{"level":"info","message":"HTTP request","method":"POST","path":"/github/webhook","status":202,"duration_ms":45,...}
{"event_type":"issues","delivery_id":"...","level":"info","message":"Received GitHub webhook",...}
{"action":"opened","issue_number":1,"repo":"TomasBack2Future/Kinetik","level":"info","message":"Processing issues event",...}
{"conversation_id":"...","repo":"TomasBack2Future/Kinetik","issue_number":1,"level":"info","message":"Handling new issue",...}
{"conversation_id":"...","session_id":"...","level":"info","message":"Executing Claude for issue analysis",...}
```

### Step 8: Verify Bot Response

Go back to your test issue on GitHub. Within 1-3 minutes, you should see:

1. ✅ A comment from your bot account with analysis
2. ✅ A label "awaiting-approval" added to the issue

### Step 9: Test Approval Workflow

1. Comment on the issue: **"approved"** or **"lgtm"**

2. Check **Terminal 1** logs again - you should see:
   ```json
   {"level":"info","message":"Handling issue approval",...}
   {"level":"info","message":"Executing Claude for implementation",...}
   ```

3. Wait 2-5 minutes for Claude to create a PR

4. Check your repository for a new pull request linked to the issue

## ngrok Web Interface

While ngrok is running, you can view detailed request/response data:

1. Open a browser
2. Go to: http://localhost:4040
3. You'll see all HTTP requests going through ngrok
4. Click on any request to see:
   - Full request headers and body
   - Response status and body
   - Timing information

This is very helpful for debugging webhook issues!

## Common Issues & Solutions

### Issue 1: "connection refused" in logs

**Problem**: Webhook server isn't running

**Solution**:
```bash
# Terminal 1
cd /root/CodeBase/code/Kinetik/automation
go run ./cmd/webhook-server/main.go
```

### Issue 2: "401 Unauthorized" in GitHub webhook deliveries

**Problem**: Webhook secret doesn't match

**Solution**:
1. Check environment variable: `echo $GITHUB_WEBHOOK_SECRET`
2. Compare with GitHub App settings
3. Restart webhook server after fixing

### Issue 3: ngrok URL changes every restart

**Problem**: Free ngrok gives random URLs

**Solutions**:
- **Option A**: Get ngrok paid plan for static domain
- **Option B**: Update GitHub webhook URL each time you restart ngrok
- **Option C**: Use ngrok's reserved domain feature

### Issue 4: No logs appearing after creating issue

**Checklist**:
1. ✅ Webhook server is running (Terminal 1 shows "Server listening")
2. ✅ ngrok is running (Terminal 2 shows "Forwarding")
3. ✅ GitHub App webhook URL includes `/github/webhook`
4. ✅ GitHub App is installed on the repository
5. ✅ Check GitHub App → Advanced → Recent Deliveries for errors

### Issue 5: "Failed to connect to database"

**Solution**:
```bash
# Test database connection
psql "$DATABASE_URL" -c "SELECT version();"

# If it fails, check:
# 1. DATABASE_URL is correct
# 2. Password is URL-encoded (! = %21)
# 3. Database exists
# 4. User has permissions
```

### Issue 6: Claude execution fails

**Solution**:
```bash
# Test Claude CLI
/root/.nvm/versions/node/v22.19.0/bin/claude --version

# Test GitHub PAT
curl -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" https://api.github.com/user

# Check logs for specific error
```

## Stopping Everything

When you're done testing:

1. **Terminal 1**: Press `Ctrl+C` to stop webhook server
2. **Terminal 2**: Press `Ctrl+C` to stop ngrok
3. Optionally update GitHub App webhook URL to empty (to avoid failed deliveries)

## Production Deployment

Once local testing works, you can deploy to production:

1. Deploy webhook server to your server with public IP/domain
2. Update GitHub App webhook URL to production URL
3. No need for ngrok anymore!

## Quick Command Reference

```bash
# Start webhook server
cd /root/CodeBase/code/Kinetik/automation
go run ./cmd/webhook-server/main.go

# Start ngrok
ngrok http 8080

# Test health endpoint
curl https://YOUR-NGROK-URL.ngrok-free.app/health

# Test database connection
psql "$DATABASE_URL" -c "SELECT 1;"

# View logs with grep
go run ./cmd/webhook-server/main.go 2>&1 | grep -E "webhook|issue"
```

## Next Steps After Successful Testing

Once you verify everything works locally:

1. ✅ Local testing complete
2. ⏭️ Deploy to production server
3. ⏭️ Set up systemd service or Docker
4. ⏭️ Configure production domain
5. ⏭️ Update GitHub App webhook URL to production
6. ⏭️ Set up monitoring and logging
7. ⏭️ Schedule PAT rotation reminder

## Need Help?

Common debugging commands:

```bash
# Check if webhook server is running
ps aux | grep webhook-server

# Check if port 8080 is in use
lsof -i :8080

# Test webhook endpoint
curl -X POST https://YOUR-NGROK-URL.ngrok-free.app/github/webhook

# View ngrok web interface
open http://localhost:4040  # or visit in browser
```

If you're still stuck, check:
1. Server logs in Terminal 1
2. ngrok web interface at http://localhost:4040
3. GitHub App → Advanced → Recent Deliveries
