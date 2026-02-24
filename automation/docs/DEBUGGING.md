# Debugging Claude CLI Bot Execution

## Problem
The bot was failing with `exit status 1` but showing empty stderr, making it impossible to debug.

## Solution Applied

### 1. Enhanced Error Logging
Modified `internal/claude/cli_client.go` to log **stdout** in addition to stderr when errors occur. Claude CLI outputs to stdout even on failure, so this captures the actual error messages.

### 2. Command Line Logging
Added logging of the exact command line being executed, including all arguments, working directory, and environment.

### 3. Verbose Mode Support
Added `--verbose` flag to Claude CLI when debug logging is enabled, providing more detailed output from Claude CLI itself.

## How to Debug

### Step 1: Enable Debug Logging

Edit `config/config.yaml` and change the logging level:

```yaml
logging:
  level: debug  # Changed from "info"
  enabled: true
  file: logs/webhook-server.log
```

### Step 2: Restart the Service

```bash
cd /root/CodeBase/code/Kinetik/automation
make restart
```

Or if running manually:
```bash
./bin/webhook-server
```

### Step 3: View Logs in Real-Time

```bash
# Follow the log file
tail -f /root/CodeBase/code/Kinetik/automation/logs/webhook-server.log

# Or view stdout directly if running in foreground
docker-compose logs -f webhook-server
```

### Step 4: Trigger the Bot

Comment on a GitHub issue with `@KinetikBot <your request>`

## What You'll See in Debug Mode

With debug logging enabled, you'll now see:

1. **Full command line being executed**:
   ```json
   {
     "command_line": "/root/.nvm/versions/node/v22.19.0/bin/claude [prompt, --print, --output-format, json, --session-id, <id>, --verbose]",
     "work_dir": "/root/CodeBase/code/Kinetik"
   }
   ```

2. **Stdout on error** (the actual error message):
   ```json
   {
     "error": "exit status 1",
     "stderr": "",
     "stdout": "<actual error message from Claude CLI>",
     "stdout_length": 1234
   }
   ```

3. **Verbose Claude CLI output** (with `--verbose` flag):
   - MCP server loading status
   - Tool execution details
   - Internal Claude CLI diagnostics

## Common Issues to Look For

### 1. MCP Server Not Loading
Look for messages about GitHub MCP server connection:
```
"Failed to connect to MCP server: github"
```

**Solution**: Check that GitHub token is set and MCP config exists in the project.

### 2. Permission Issues
```
"EACCES: permission denied"
```

**Solution**: Check file permissions on work directory and Claude CLI executable.

### 3. Claude CLI Not Found
```
"fork/exec /root/.nvm/versions/node/v22.19.0/bin/claude: no such file or directory"
```

**Solution**: Verify Claude CLI is installed and path in config.yaml is correct:
```bash
which claude
# Update config.yaml with the correct path
```

### 4. Timeout Issues
```
"context deadline exceeded"
```

**Solution**: Increase timeout in config.yaml:
```yaml
claude:
  timeout: 600s  # Increased from 300s
```

## Reverting to Normal Logging

Once you've identified the issue, revert to info level to reduce log noise:

```yaml
logging:
  level: info
  enabled: true
  file: logs/webhook-server.log
```

## Viewing Historical Logs

Logs are appended to `logs/webhook-server.log`. Use standard log viewing tools:

```bash
# View last 100 lines
tail -n 100 logs/webhook-server.log

# Search for specific session
grep "session_id.*<session-id>" logs/webhook-server.log

# View all errors
grep '"level":"error"' logs/webhook-server.log | jq .

# View Claude CLI executions
grep "Executing Claude CLI" logs/webhook-server.log | jq .
```

## Testing Locally

You can test Claude CLI execution directly:

```bash
cd /root/CodeBase/code/Kinetik

# Test basic execution
claude "help me understand the repo" --print --output-format json

# Test with verbose mode
claude "help me understand the repo" --print --output-format json --verbose

# Check MCP server status
claude --list-mcp-servers
```

## Next Steps

1. Enable debug logging
2. Reproduce the error
3. Check the stdout in the error logs
4. Look for specific error messages mentioned above
5. Fix the underlying issue
6. Disable debug logging

## Getting Help

If you're still stuck after reviewing debug logs, include:
- Full command line from logs
- Complete stdout from error logs
- Session ID for tracing
- Output of `claude --version`
