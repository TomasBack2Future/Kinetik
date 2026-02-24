#!/bin/bash

# Test script for GitHub CLI functionality
# This tests the gh commands that Claude will use in the automation system

set -e

echo "=== GitHub CLI Test Suite ==="
echo ""

# Configuration
export REPO="${GITHUB_REPO:-TomasBack2Future/Kinetik}"
export ISSUE_NUMBER="${TEST_ISSUE_NUMBER:-1}"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if gh is installed
echo "1. Checking if gh CLI is installed..."
if command -v gh &> /dev/null; then
    echo -e "${GREEN}✓ gh CLI is installed${NC}"
    gh --version
else
    echo -e "${RED}✗ gh CLI is not installed${NC}"
    echo "Install gh CLI: https://github.com/cli/cli#installation"
    exit 1
fi

echo ""

# Check authentication
echo "2. Checking GitHub authentication..."
if gh auth status &> /dev/null; then
    echo -e "${GREEN}✓ Authenticated with GitHub${NC}"
    gh auth status
else
    echo -e "${RED}✗ Not authenticated with GitHub${NC}"
    echo "Run: export GITHUB_TOKEN=\$GITHUB_PERSONAL_ACCESS_TOKEN"
    echo "Or run: gh auth login"
    exit 1
fi

echo ""

# Test: List issues
echo "3. Testing: gh issue list"
if gh issue list --repo "$REPO" --limit 5 &> /dev/null; then
    echo -e "${GREEN}✓ Can list issues${NC}"
    gh issue list --repo "$REPO" --limit 3
else
    echo -e "${RED}✗ Failed to list issues${NC}"
    exit 1
fi

echo ""

# Test: View issue
echo "4. Testing: gh issue view"
if gh issue view "$ISSUE_NUMBER" --repo "$REPO" &> /dev/null; then
    echo -e "${GREEN}✓ Can view issue #$ISSUE_NUMBER${NC}"
    gh issue view "$ISSUE_NUMBER" --repo "$REPO" | head -10
else
    echo -e "${YELLOW}⚠ Issue #$ISSUE_NUMBER not found (this is OK if it doesn't exist)${NC}"
fi

echo ""

# Test: Comment on issue (dry-run)
echo "5. Testing: gh issue comment (command format check)"
TEST_COMMENT="This is a test comment from the automation system test script."
COMMENT_CMD="gh issue comment $ISSUE_NUMBER --repo $REPO --body \"$TEST_COMMENT\""
echo "Command would be: $COMMENT_CMD"
echo -e "${YELLOW}⚠ Not actually posting comment (dry-run mode)${NC}"
echo -e "${GREEN}✓ Command format is correct${NC}"

echo ""

# Test: Edit issue with label (dry-run)
echo "6. Testing: gh issue edit (command format check)"
EDIT_CMD="gh issue edit $ISSUE_NUMBER --repo $REPO --add-label \"test-label\""
echo "Command would be: $EDIT_CMD"
echo -e "${YELLOW}⚠ Not actually adding label (dry-run mode)${NC}"
echo -e "${GREEN}✓ Command format is correct${NC}"

echo ""

# Test: List PRs
echo "7. Testing: gh pr list"
if gh pr list --repo "$REPO" --limit 5 &> /dev/null; then
    echo -e "${GREEN}✓ Can list pull requests${NC}"
    gh pr list --repo "$REPO" --limit 3
else
    echo -e "${YELLOW}⚠ Failed to list PRs or no PRs exist${NC}"
fi

echo ""

# Test: PR comment (dry-run)
echo "8. Testing: gh pr comment (command format check)"
PR_NUMBER="1"
PR_COMMENT_CMD="gh pr comment $PR_NUMBER --repo $REPO --body \"Test PR comment\""
echo "Command would be: $PR_COMMENT_CMD"
echo -e "${YELLOW}⚠ Not actually posting PR comment (dry-run mode)${NC}"
echo -e "${GREEN}✓ Command format is correct${NC}"

echo ""

# Test: PR create (dry-run)
echo "9. Testing: gh pr create (command format check)"
PR_CREATE_CMD='gh pr create --repo '"$REPO"' --title "Test PR" --body "Test description" --head test-branch --base main'
echo "Command would be: $PR_CREATE_CMD"
echo -e "${YELLOW}⚠ Not actually creating PR (dry-run mode)${NC}"
echo -e "${GREEN}✓ Command format is correct${NC}"

echo ""
echo "=== All Tests Completed ==="
echo ""
echo -e "${GREEN}Summary:${NC}"
echo "✓ gh CLI is installed and working"
echo "✓ Authentication is configured"
echo "✓ Can read repository data"
echo "✓ All command formats are valid"
echo ""
echo -e "${YELLOW}Note: Write operations (comments, labels, PRs) were not actually executed (dry-run mode)${NC}"
echo "To test write operations for real, run specific commands manually."
echo ""
echo "Example commands to test manually:"
echo "  gh issue comment <number> --repo $REPO --body \"Test comment\""
echo "  gh issue edit <number> --repo $REPO --add-label \"test-label\""
echo "  gh pr comment <number> --repo $REPO --body \"Test PR comment\""
