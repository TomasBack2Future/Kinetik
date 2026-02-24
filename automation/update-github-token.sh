#!/bin/bash

# Script to update GitHub token in all locations

echo "=== GitHub Token Update Script ==="
echo ""
echo "This script will help you update your GitHub token in all necessary locations."
echo ""

# Prompt for new token
read -p "Enter your new GitHub Personal Access Token (ghp_...): " NEW_TOKEN

if [[ ! $NEW_TOKEN =~ ^ghp_ ]]; then
    echo "ERROR: Token should start with 'ghp_'"
    exit 1
fi

if [ ${#NEW_TOKEN} -ne 40 ]; then
    echo "ERROR: Token should be 40 characters long"
    exit 1
fi

echo ""
echo "Token format looks correct!"
echo ""

# Update environment variable
export GITHUB_PERSONAL_ACCESS_TOKEN="$NEW_TOKEN"
export GITHUB_TOKEN="$NEW_TOKEN"

echo "✓ Updated environment variables for current session"
echo ""

# Test the token with gh CLI
echo "Testing token with gh CLI..."
if gh auth login --with-token <<< "$NEW_TOKEN" 2>&1 | grep -q "Logged in"; then
    echo "✓ gh CLI authentication successful!"
else
    echo "Testing alternative method..."
    gh auth status 2>&1 || true
fi

echo ""
echo "Now manually update the token in these locations:"
echo ""
echo "1. Update in your shell profile:"
echo "   echo 'export GITHUB_PERSONAL_ACCESS_TOKEN=\"$NEW_TOKEN\"' >> ~/.bashrc"
echo ""
echo "2. Update in automation config:"
echo "   File: /root/CodeBase/code/Kinetik/automation/config/config.yaml"
echo "   (Token is read from GITHUB_PERSONAL_ACCESS_TOKEN environment variable)"
echo ""
echo "3. Reload environment:"
echo "   source ~/.bashrc"
echo ""
echo "4. Test gh CLI:"
echo "   export GITHUB_TOKEN=\"\$GITHUB_PERSONAL_ACCESS_TOKEN\""
echo "   gh auth status"
echo "   gh repo view TomasBack2Future/Kinetik"
echo ""
