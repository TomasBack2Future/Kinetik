package workflow

import (
	"regexp"
	"strings"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
)

// IssueValidator checks if an issue has required information
type IssueValidator struct{}

func NewIssueValidator() *IssueValidator {
	return &IssueValidator{}
}

// ValidationResult contains validation outcome
type ValidationResult struct {
	Valid          bool
	MissingFields  []string
	HasCommitInfo  bool
	HasVersionInfo bool
}

// ValidateIssue checks if issue contains commit hash or version info
func (v *IssueValidator) ValidateIssue(issue *types.Issue) *ValidationResult {
	result := &ValidationResult{
		Valid:         true,
		MissingFields: []string{},
	}

	body := strings.ToLower(issue.Body)

	// Check for commit hash (SHA-1: 40 hex chars or short form 7+ chars)
	commitPattern := regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)
	result.HasCommitInfo = commitPattern.MatchString(body) ||
		strings.Contains(body, "commit") ||
		strings.Contains(body, "sha") ||
		strings.Contains(body, "hash")

	// Check for version info
	versionPattern := regexp.MustCompile(`v?\d+\.\d+(\.\d+)?`)
	result.HasVersionInfo = versionPattern.MatchString(body) ||
		strings.Contains(body, "version") ||
		strings.Contains(body, "release")

	// Issue is invalid if missing both
	if !result.HasCommitInfo && !result.HasVersionInfo {
		result.Valid = false
		result.MissingFields = append(result.MissingFields, "commit/version information")
	}

	return result
}

// BuildRequestInfoComment generates a comment asking for missing info
func (v *IssueValidator) BuildRequestInfoComment(validation *ValidationResult) string {
	if validation.Valid {
		return ""
	}

	comment := "👋 Thanks for opening this issue!\n\n"
	comment += "To help me better understand and fix this issue, could you please provide:\n\n"

	if !validation.HasCommitInfo && !validation.HasVersionInfo {
		comment += "- **Commit hash** or **version number** where you encountered this issue\n"
		comment += "  - Example: `commit: abc123def` or `version: v1.2.3`\n\n"
	}

	comment += "This information helps me reproduce the issue in the correct codebase state.\n\n"
	comment += "Once you've added this info, mention me with `@KinetikBot` and I'll get started!"

	return comment
}
