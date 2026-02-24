package workflow

import (
	"strings"
	"testing"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
)

func TestValidateIssue(t *testing.T) {
	validator := NewIssueValidator()

	tests := []struct {
		name            string
		issueBody       string
		expectedValid   bool
		expectedCommit  bool
		expectedVersion bool
	}{
		{
			name:            "Has commit hash",
			issueBody:       "Bug found in commit abc123def456",
			expectedValid:   true,
			expectedCommit:  true,
			expectedVersion: false,
		},
		{
			name:            "Has version",
			issueBody:       "Issue in version v1.2.3",
			expectedValid:   true,
			expectedCommit:  false,
			expectedVersion: true,
		},
		{
			name:            "Has both",
			issueBody:       "Bug in v2.0.0 at commit abc123",
			expectedValid:   true,
			expectedCommit:  true,
			expectedVersion: true,
		},
		{
			name:            "Missing both",
			issueBody:       "Something is broken",
			expectedValid:   false,
			expectedCommit:  false,
			expectedVersion: false,
		},
		{
			name:            "Has commit keyword",
			issueBody:       "Found at commit: main branch",
			expectedValid:   true,
			expectedCommit:  true,
			expectedVersion: false,
		},
		{
			name:            "Has version keyword",
			issueBody:       "Latest version has this bug",
			expectedValid:   true,
			expectedCommit:  false,
			expectedVersion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &types.Issue{
				Body: tt.issueBody,
			}

			result := validator.ValidateIssue(issue)

			if result.Valid != tt.expectedValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectedValid, result.Valid)
			}

			if result.HasCommitInfo != tt.expectedCommit {
				t.Errorf("Expected HasCommitInfo=%v, got %v", tt.expectedCommit, result.HasCommitInfo)
			}

			if result.HasVersionInfo != tt.expectedVersion {
				t.Errorf("Expected HasVersionInfo=%v, got %v", tt.expectedVersion, result.HasVersionInfo)
			}
		})
	}
}

func TestBuildRequestInfoComment(t *testing.T) {
	validator := NewIssueValidator()

	validation := &ValidationResult{
		Valid:          false,
		MissingFields:  []string{"commit/version information"},
		HasCommitInfo:  false,
		HasVersionInfo: false,
	}

	comment := validator.BuildRequestInfoComment(validation)

	if comment == "" {
		t.Error("Expected non-empty comment")
	}

	// Check for key phrases (case-insensitive)
	commentLower := strings.ToLower(comment)
	if !contains(commentLower, "commit hash") {
		t.Error("Comment should mention commit hash")
	}

	if !contains(commentLower, "version number") {
		t.Error("Comment should mention version number")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
