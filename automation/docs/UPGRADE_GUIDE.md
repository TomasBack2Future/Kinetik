# Upgrade Guide: Queue & Validation System

## Overview

This upgrade adds issue validation and queue processing to your automation system. The changes are backward compatible and require no configuration changes.

## What's New

1. **Automatic issue validation** - Checks for commit/version info
2. **Sequential issue processing** - One issue at a time, each on its own branch
3. **Smart queue management** - Automatic FIFO processing

## Upgrade Steps

### 1. Pull Latest Code

```bash
cd /root/CodeBase/code/Kinetik/automation
git pull origin main
```

### 2. Build New Version

```bash
make build
# or
go build -o bin/webhook-server ./cmd/webhook-server
```

### 3. Run Tests (Optional but Recommended)

```bash
go test ./internal/workflow/... -v
```

Expected output: All tests pass

### 4. Deploy

#### Option A: Docker Compose

```bash
make docker-build
docker-compose down
docker-compose up -d
```

#### Option B: Systemd Service

```bash
sudo systemctl stop kinetik-automation
sudo cp bin/webhook-server /opt/kinetik-automation/
sudo systemctl start kinetik-automation
sudo systemctl status kinetik-automation
```

### 5. Verify Deployment

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check logs
docker-compose logs -f webhook-server
# or
sudo journalctl -u kinetik-automation -f
```

## Testing the New Features

### Test Issue Validation

1. Create a test issue WITHOUT commit/version info:
   ```
   Title: Test validation
   Body: Something is broken
   ```

2. Expected behavior:
   - Bot posts comment requesting info
   - `needs-info` label added
   - Issue NOT processed

3. Add commit info and mention bot:
   ```
   Found in commit abc123
   @KinetikBot
   ```

4. Expected behavior:
   - Bot acknowledges and queues issue
   - Processing starts

### Test Queue System

1. Create 3 issues quickly (all with commit/version info)

2. Expected behavior:
   - All 3 queued
   - Processed one at a time
   - Each gets unique branch name

3. Check logs:
   ```bash
   grep -i "queue" logs/webhook-server.log
   ```

## Configuration Changes

**None required!** The system uses existing configuration.

Optional: Update bot username if different:

```yaml
# config/config.yaml
github:
  bot_username: KinetikBot  # Update if needed
```

## Rollback Procedure

If you need to rollback:

```bash
# Get previous commit hash
git log --oneline -5

# Rollback
git checkout <previous-commit-hash>

# Rebuild
make build

# Redeploy (same steps as upgrade)
```

## Breaking Changes

**None.** This is a backward-compatible upgrade.

Existing issues will continue to work. New validation only applies to newly created issues.

## Database Changes

**None.** No schema migrations needed.

The system uses existing conversation context storage for branch names.

## API Changes

No external API changes. Internal changes:

- `NewOrchestrator()` now requires `githubClient` parameter
- New internal methods added (not exposed externally)

## Monitoring

### New Log Messages

Watch for these in your logs:

```
Issue queued for processing
Processing queued issue
Completed processing queued issue
Issue missing required information, requesting details
```

### Metrics to Monitor

- Queue length (check logs)
- Validation failure rate
- Branch creation success rate
- Processing time per issue

## Troubleshooting

### Issue: Bot not requesting missing info

**Check:**
1. Bot has permission to comment on issues
2. `GITHUB_PERSONAL_ACCESS_TOKEN` is valid
3. Logs show validation running

**Fix:**
```bash
# Verify token
curl -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" \
  https://api.github.com/user
```

### Issue: Queue not processing

**Check:**
1. Logs show "Issue queued for processing"
2. No errors in logs
3. Queue processor started

**Fix:**
```bash
# Check logs for errors
grep -i "error" logs/webhook-server.log

# Restart service
sudo systemctl restart kinetik-automation
```

### Issue: Branch creation fails

**Check:**
1. Bot has write permission to repository
2. Base branch exists (default: "main")
3. Branch name not already taken

**Fix:**
- Verify bot is collaborator with write access
- Check if base branch is "master" instead of "main"
- Delete old branches if needed

## Support

For issues:
1. Check logs first
2. Review QUEUE_AND_VALIDATION.md
3. Run tests to verify installation
4. Check GitHub permissions

## Next Steps

After successful upgrade:

1. Monitor logs for first few issues
2. Verify validation works as expected
3. Check queue processing is sequential
4. Confirm branches are created correctly

## Changelog

### Added
- Issue validation for commit/version info
- Sequential issue queue
- Automatic branch creation per issue
- GitHub client for API operations

### Changed
- `HandleNewIssue()` now validates before processing
- `HandleIssueMention()` re-validates and queues
- Orchestrator constructor requires GitHub client

### Fixed
- N/A (new feature, no fixes)

## Version

This upgrade is compatible with:
- Go 1.22+
- PostgreSQL 12+
- Claude CLI (any version)
- GitHub API v3

## Questions?

See:
- QUEUE_AND_VALIDATION.md - Feature documentation
- IMPLEMENTATION_NOTES.md - Technical details
- README.md - General setup guide
