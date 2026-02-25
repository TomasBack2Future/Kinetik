#!/usr/bin/env bash
# checkout-server-commit.sh
#
# Given a KinetikServer git commit or tag:
#   1. Cleans KinetikServer and kinetik_agent (resets uncommitted + removes untracked)
#   2. Checks out that commit in KinetikServer
#   3. Reads the agent version from .github/workflows/docker-build.yml at that commit
#   4. Checks out the matching tag in kinetik_agent and updates its submodules
#   5. If a branch name is provided, creates that branch in both repos
#
# Usage:
#   ./scripts/checkout-server-commit.sh <server-commit-or-tag> [branch-name]
#
# Examples:
#   ./scripts/checkout-server-commit.sh 9bc8088
#   ./scripts/checkout-server-commit.sh v1.2.3 fix/my-bug

set -euo pipefail

SERVER_COMMIT="${1:-}"
BRANCH_NAME="${2:-}"

if [[ -z "$SERVER_COMMIT" ]]; then
  echo "ERROR: You must provide a server commit hash or tag." >&2
  echo "Usage: $0 <server-commit-or-tag> [branch-name]" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# KINETIK_ROOT can be set explicitly (e.g. when running inside Docker with a volume mount).
# Falls back to two levels up from this script (automation/scripts/ -> Kinetik root).
ROOT="${KINETIK_ROOT:-$(cd "$SCRIPT_DIR/../.." && pwd)}"
SERVER_DIR="$ROOT/KinetikServer"
AGENT_DIR="$ROOT/kinetik_agent"

# --- Helper: clean a repo ---
clean_repo() {
  local dir="$1"
  local name="$2"
  echo "==> Cleaning $name (resetting uncommitted changes and removing untracked files)..."
  git -C "$dir" reset --hard HEAD
  git -C "$dir" clean -fd
}

# --- Clean both repos ---
clean_repo "$SERVER_DIR" "KinetikServer"
clean_repo "$AGENT_DIR"  "kinetik_agent"

# --- Checkout server commit ---
echo "==> Fetching latest refs for KinetikServer..."
git -C "$SERVER_DIR" fetch --tags --quiet

echo "==> Checking out KinetikServer at: $SERVER_COMMIT"
git -C "$SERVER_DIR" checkout "$SERVER_COMMIT"

# --- Derive agent version from workflow ---
WORKFLOW_FILE="$SERVER_DIR/.github/workflows/docker-build.yml"
if [[ ! -f "$WORKFLOW_FILE" ]]; then
  echo "ERROR: Workflow file not found at $WORKFLOW_FILE" >&2
  exit 1
fi

AGENT_VERSION=$(grep -oP '(?<=download_agent\.sh\s)\S+' "$WORKFLOW_FILE" | head -1)
if [[ -z "$AGENT_VERSION" ]]; then
  echo "ERROR: Could not extract agent version from $WORKFLOW_FILE" >&2
  exit 1
fi

# --- Checkout agent at matching tag ---
echo "==> Fetching latest refs for kinetik_agent..."
git -C "$AGENT_DIR" fetch --tags --quiet

AGENT_TAG="v${AGENT_VERSION}"
if git -C "$AGENT_DIR" rev-parse "$AGENT_TAG" &>/dev/null; then
  AGENT_REF="$AGENT_TAG"
elif git -C "$AGENT_DIR" rev-parse "$AGENT_VERSION" &>/dev/null; then
  AGENT_REF="$AGENT_VERSION"
else
  echo "WARNING: Tag '$AGENT_TAG' not found in kinetik_agent. Using HEAD." >&2
  AGENT_REF="HEAD"
fi

echo "==> Checking out kinetik_agent at: $AGENT_REF"
git -C "$AGENT_DIR" checkout "$AGENT_REF"

# --- Init and update submodules inside kinetik_agent ---
echo "==> Initialising and updating submodules inside kinetik_agent..."
git -C "$AGENT_DIR" submodule update --init --recursive

# --- Optionally create branches ---
if [[ -n "$BRANCH_NAME" ]]; then
  echo "==> Creating branch '$BRANCH_NAME' in KinetikServer..."
  git -C "$SERVER_DIR" checkout -b "$BRANCH_NAME"

  echo "==> Creating branch '$BRANCH_NAME' in kinetik_agent..."
  git -C "$AGENT_DIR" checkout -b "$BRANCH_NAME"
fi

echo ""
echo "==> Results:"
echo "    Server commit : $SERVER_COMMIT"
echo "    Agent version : $AGENT_VERSION  (ref: $AGENT_REF)"
[[ -n "$BRANCH_NAME" ]] && echo "    Branch created: $BRANCH_NAME (in KinetikServer + kinetik_agent)"
echo ""
echo "AGENT_VERSION=$AGENT_VERSION"
echo "SERVER_COMMIT=$SERVER_COMMIT"
[[ -n "$BRANCH_NAME" ]] && echo "BRANCH_NAME=$BRANCH_NAME"
