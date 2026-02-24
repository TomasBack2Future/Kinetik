# Webhook Not Received - Debugging Checklist

Follow these steps in order to find the issue.

## Step 1: Is the Webhook Server Running?

Check your terminal where you ran `make run`:

```
✅ Should see: {"level":"info","message":"Server listening on :8080",...}
❌ If not running: Start it with `make run`
```

## Step 2: Is ngrok Running?

Check your ngrok terminal:

```
✅ Should see: Forwarding https://xxxx.ngrok-free.app -> http://localhost:8080
❌ If not running: Start it with `ngrok http 8080`
```

## Step 3: Check GitHub App Webhook Configuration

1. Go to: https://github.com/settings/apps
2. Click your app name
3. Scroll to **"Webhook"** section
4. Check:
   - ✅ Webhook URL is set to: `https://YOUR-NGROK-URL.ngrok-free.app/github/webhook`
   - ✅ **Active** checkbox is checked
   - ✅ Webhook secret matches your environment variable
   - ✅ Content type is `application/json`

## Step 4: Check Recent Webhook Deliveries

This is the most important step!

1. In your GitHub App settings, click **"Advanced"** tab (at the top)
2. Scroll down to **"Recent Deliveries"**
3. You should see delivery attempts for your issue

### What to Look For:

#### Scenario A: No deliveries at all
**Problem**: GitHub App not installed or not subscribed to events

**Solutions**:
- Go to app settings → "Install App" → Verify it's installed on `TomasBack2Future/Kinetik`
- Go to app settings → "Permissions & events" → Verify "Issues" is subscribed
- Reinstall the app if needed

#### Scenario B: Deliveries with red X (failed)
**Problem**: Webhook URL is wrong or server not reachable

**Click on the failed delivery to see details**:
- Look at Response tab
- Common errors:
  - `Connection refused` → Server not running or wrong port
  - `404 Not Found` → Wrong URL path (missing `/github/webhook`)
  - `401 Unauthorized` → Webhook secret mismatch
  - `Could not resolve host` → ngrok URL is wrong

#### Scenario C: Deliveries with green checkmark
**Problem**: Webhook delivered but server not processing it

**Check**: Your server logs should show the webhook

## Step 5: Test ngrok URL Directly

Open a new terminal and test:

```bash
# Replace with YOUR actual ngrok URL
curl https://YOUR-NGROK-URL.ngrok-free.app/health

# Should return: OK
```

If this fails, ngrok is not working correctly.

## Step 6: Check ngrok Web Interface

1. Open browser: http://localhost:4040
2. Look for POST requests to `/github/webhook`
3. If you see requests here, ngrok is receiving them
4. Check the response code:
   - 200/202 = Success
   - 401 = Unauthorized (secret mismatch)
   - 404 = Wrong path
   - 500 = Server error

## Step 7: Check Server Logs

In your webhook server terminal, look for:

```json
{"level":"info","message":"HTTP request","method":"POST","path":"/github/webhook",...}
```

If you don't see this, the request isn't reaching your server.

## Step 8: Verify GitHub App Installation

1. Go to: https://github.com/TomasBack2Future/Kinetik/settings/installations
2. You should see your app listed
3. Click "Configure"
4. Verify:
   - ✅ Repository access includes this repo
   - ✅ Permissions are granted

## Quick Diagnosis Script

Run this in a terminal:

```bash
echo "=== Webhook Debugging ==="
echo ""
echo "1. Webhook Server Status:"
ps aux | grep webhook-server | grep -v grep && echo "✅ Running" || echo "❌ Not running"
echo ""
echo "2. Port 8080 Status:"
lsof -i :8080 && echo "✅ Port in use" || echo "❌ Port not in use"
echo ""
echo "3. ngrok Status:"
ps aux | grep ngrok | grep -v grep && echo "✅ Running" || echo "❌ Not running"
echo ""
echo "4. Environment Variables:"
[ -n "$GITHUB_WEBHOOK_SECRET" ] && echo "✅ GITHUB_WEBHOOK_SECRET set" || echo "❌ GITHUB_WEBHOOK_SECRET not set"
[ -n "$GITHUB_PERSONAL_ACCESS_TOKEN" ] && echo "✅ GITHUB_PERSONAL_ACCESS_TOKEN set" || echo "❌ GITHUB_PERSONAL_ACCESS_TOKEN not set"
[ -n "$DB_PASSWORD" ] && echo "✅ DB_PASSWORD set" || echo "❌ DB_PASSWORD not set"
echo ""
echo "5. Test Health Endpoint:"
echo "Run: curl http://localhost:8080/health"
echo "Expected: OK"
```

## Common Issues & Fixes

### Issue 1: "Recent Deliveries" is Empty

**Cause**: GitHub isn't sending webhooks

**Fix**:
1. Verify app is installed: Settings → Integrations → GitHub Apps
2. Verify event subscriptions: App Settings → Permissions & events
3. Try creating another issue to trigger a new webhook

### Issue 2: Connection Refused

**Cause**: Webhook server not running

**Fix**:
```bash
cd /root/CodeBase/code/Kinetik/automation
make run
```

### Issue 3: 404 Not Found

**Cause**: Wrong webhook URL

**Fix**: Update GitHub App webhook URL to include `/github/webhook`:
```
https://YOUR-NGROK-URL.ngrok-free.app/github/webhook
                                       ^^^^^^^^^^^^^^^^^
                                       Don't forget this part!
```

### Issue 4: 401 Unauthorized

**Cause**: Webhook secret mismatch

**Fix**:
```bash
# Check what's set
echo $GITHUB_WEBHOOK_SECRET

# Compare with GitHub App settings
# They must match exactly!

# If different, export the correct one:
export GITHUB_WEBHOOK_SECRET="67f9f0d9dc944c1a10d5a189af2d7d34c1cefdf182a14a90c2780330110422a1"

# Restart webhook server
```

### Issue 5: ngrok URL Changed

**Cause**: ngrok was restarted (free tier gives new URL each time)

**Fix**:
1. Check ngrok terminal for current URL
2. Update GitHub App webhook URL with new ngrok URL
3. Remember to add `/github/webhook` at the end

## Manual Webhook Test

You can manually trigger a webhook from GitHub:

1. Go to: GitHub App Settings → Advanced → Recent Deliveries
2. Find a previous delivery (if any)
3. Click on it
4. Click "Redeliver" button
5. Check server logs

## Still Not Working?

Run through this checklist:

- [ ] Webhook server shows "Server listening on :8080"
- [ ] ngrok shows "Forwarding https://xxx.ngrok-free.app -> http://localhost:8080"
- [ ] GitHub App webhook URL = `https://xxx.ngrok-free.app/github/webhook`
- [ ] GitHub App webhook Active checkbox is checked
- [ ] GitHub App is installed on the repository
- [ ] `curl http://localhost:8080/health` returns "OK"
- [ ] GitHub App → Advanced → Recent Deliveries shows delivery attempts
- [ ] http://localhost:4040 shows incoming requests

If all checked and still not working, share:
1. Screenshot of GitHub App webhook settings
2. Screenshot of Recent Deliveries
3. Output from ngrok terminal
4. Output from webhook server logs
