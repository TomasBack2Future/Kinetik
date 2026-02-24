# GitHub CLI Testing Guide

## Quick Installation (Ubuntu/Debian)

### Option 1: Using apt (recommended)

```bash
# Add GitHub CLI repository
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null

# Install
sudo apt update
sudo apt install gh -y
```

### Option 2: Download .deb manually

```bash
# Download latest release
cd /tmp
wget https://github.com/cli/cli/releases/download/v2.62.0/gh_2.62.0_linux_amd64.deb

# Install
sudo dpkg -i gh_2.62.0_linux_amd64.deb
```

### Option 3: Use the compiled binary

```bash
# Download and extract
cd /tmp
wget https://github.com/cli/cli/releases/download/v2.62.0/gh_2.62.0_linux_amd64.tar.gz
tar xzf gh_2.62.0_linux_amd64.tar.gz

# Move to PATH
sudo mv gh_2.62.0_linux_amd64/bin/gh /usr/local/bin/
```

## Authentication

### Method 1: Using environment variable (recommended for automation)

```bash
export GITHUB_TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN"
# Or
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# Test authentication
gh auth status
```

### Method 2: Interactive login

```bash
gh auth login
```

## Run Tests

```bash
cd /root/CodeBase/code/Kinetik/automation

# Set your test parameters
export GITHUB_TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN"
export GITHUB_REPO="TomasBack2Future/Kinetik"
export TEST_ISSUE_NUMBER="1"  # Use an existing issue number

# Run the test script
./test-gh-cli.sh
```

## Manual Testing

### Test Issue Operations

```bash
# List issues
gh issue list --repo TomasBack2Future/Kinetik

# View specific issue
gh issue view 1 --repo TomasBack2Future/Kinetik

# Post a comment (WILL ACTUALLY POST!)
gh issue comment 1 --repo TomasBack2Future/Kinetik --body "Test comment from gh CLI"

# Add a label (WILL ACTUALLY ADD!)
gh issue edit 1 --repo TomasBack2Future/Kinetik --add-label "test-label"

# Multi-line comment with heredoc
gh issue comment 1 --repo TomasBack2Future/Kinetik --body "$(cat <<'EOF'
This is a multi-line comment.

## Analysis
- Point 1
- Point 2

This is how Claude will post comments.
EOF
)"
```

### Test PR Operations

```bash
# List PRs
gh pr list --repo TomasBack2Future/Kinetik

# View PR
gh pr view 1 --repo TomasBack2Future/Kinetik

# Comment on PR (WILL ACTUALLY POST!)
gh pr comment 1 --repo TomasBack2Future/Kinetik --body "Test PR comment"

# Create PR (WILL ACTUALLY CREATE!)
gh pr create --repo TomasBack2Future/Kinetik \
  --title "Test PR from gh CLI" \
  --body "This is a test pull request" \
  --head test-branch \
  --base main
```

## Verify Commands Claude Will Use

These are the exact commands the automation system will use:

### Issue Analysis Workflow
```bash
# 1. Post analysis comment
gh issue comment <number> --repo TomasBack2Future/Kinetik --body "$(cat <<'EOF'
<analysis content here>
EOF
)"

# 2. Add awaiting-approval label
gh issue edit <number> --repo TomasBack2Future/Kinetik --add-label "awaiting-approval"
```

### Issue Implementation Workflow
```bash
# 1. Create pull request
gh pr create --repo TomasBack2Future/Kinetik \
  --title "<title>" \
  --body "<description>" \
  --head <branch-name> \
  --base main

# 2. Link PR to issue
gh issue comment <issue-number> --repo TomasBack2Future/Kinetik --body "Implemented in PR #<pr-number>"
```

### Issue Mention Workflow
```bash
# Respond to mention
gh issue comment <number> --repo TomasBack2Future/Kinetik --body "<response>"
```

### PR Review Workflow
```bash
# Comment on PR
gh pr comment <number> --repo TomasBack2Future/Kinetik --body "<response>"
```

## Troubleshooting

### Check gh version
```bash
gh --version
```

### Check authentication status
```bash
gh auth status
```

Should show:
```
github.com
  ✓ Logged in to github.com as YourUsername
  ✓ Git operations for github.com configured to use https protocol.
  ✓ Token: ghp_************************************
```

### Test API access
```bash
# Get your user info
gh api user

# Check repo access
gh repo view TomasBack2Future/Kinetik
```

### Common Issues

**Error: "HTTP 401: Bad credentials"**
- Check that `GITHUB_TOKEN` is set correctly
- Verify token has required permissions: `repo`, `workflow`
- Try: `gh auth refresh`

**Error: "resource not accessible by personal access token"**
- Token needs additional scopes
- Regenerate token with `repo`, `workflow`, `write:packages` permissions

**Error: "Not Found"**
- Check repository name is correct
- Verify you have access to the repository
- Make sure issue/PR number exists

## Testing in Docker

Once Docker build succeeds, you can test gh CLI in the container:

```bash
# Build image (when network is working)
cd /root/CodeBase/code/Kinetik/automation
docker build -t kinetik-automation:latest .

# Run interactive shell in container
docker run -it --rm \
  -e GITHUB_TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN" \
  kinetik-automation:latest \
  /bin/sh

# Inside container, test gh CLI
gh --version
gh auth status
gh repo view TomasBack2Future/Kinetik
```

## Expected Test Results

When you run `./test-gh-cli.sh`, you should see:

```
=== GitHub CLI Test Suite ===

1. Checking if gh CLI is installed...
✓ gh CLI is installed
gh version 2.62.0 (2024-xx-xx)

2. Checking GitHub authentication...
✓ Authenticated with GitHub
github.com
  ✓ Logged in to github.com as YourUsername
  ...

3. Testing: gh issue list
✓ Can list issues
#1  Issue Title  (label1, label2)  about 1 day ago
...

All Tests Completed

Summary:
✓ gh CLI is installed and working
✓ Authentication is configured
✓ Can read repository data
✓ All command formats are valid
```

## Next Steps After Testing

1. Verify all test commands pass
2. Manually test write operations on a test issue
3. Confirm heredoc syntax works for multi-line comments
4. Build Docker image with gh CLI installed
5. Deploy updated automation system
6. Monitor first real webhook event

## Safety Notes

- The test script runs in dry-run mode for write operations
- To test writes, run commands manually
- Use a test issue/repository if possible
- Always verify the `--repo` flag is correct to avoid posting to wrong repo
