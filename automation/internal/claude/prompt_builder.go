package claude

import (
	"fmt"
	"strings"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
)

type PromptBuilder struct {
	botUsername string
}

func NewPromptBuilder(botUsername string) *PromptBuilder {
	return &PromptBuilder{
		botUsername: botUsername,
	}
}

// getGitHubInstructions returns standardized gh CLI usage instructions
func (pb *PromptBuilder) getGitHubInstructions() string {
	return `
**IMPORTANT: GitHub Interaction Instructions**

Use the GitHub CLI (gh) for all GitHub operations via the Bash tool. The gh CLI is authenticated with your GITHUB_TOKEN.

**CRITICAL: For ALL comments (issues or PRs), you MUST:**
1. First write the comment content to a file (e.g., /tmp/comment.md)
2. Then use --body-file flag to post the comment

**Common Commands:**
- Post issue comment:
  1. Write content: Write tool to create /tmp/comment-<issue_number>.md
  2. Post: gh issue comment <issue_number> --repo <owner/repo> --body-file /tmp/comment-<issue_number>.md
- Post PR comment:
  1. Write content: Write tool to create /tmp/comment-pr-<pr_number>.md
  2. Post: gh pr comment <pr_number> --repo <owner/repo> --body-file /tmp/comment-pr-<pr_number>.md
- Add label: gh issue edit <issue_number> --repo <owner/repo> --add-label "label-name"
- Create PR: gh pr create --repo <owner/repo> --title "title" --body "description" --head branch --base main

**Guidelines:**
- ALWAYS use --body-file instead of --body for comments (to avoid size limitations)
- Always use --repo <owner/repo> flag to specify repository
- Check command exit codes to verify success
- Repository format: "owner/repo" (e.g., "TomasBack2Future/Kinetik")

`
}

// BuildIssueAnalysisPrompt creates a prompt for analyzing a new issue
func (pb *PromptBuilder) BuildIssueAnalysisPrompt(event *types.IssueEvent, conversationContext string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a GitHub automation bot for the Kinetik project. A new issue was created.\n\n")
	fmt.Fprintf(&prompt, "**Repository:** %s\n", event.Repository.FullName)
	fmt.Fprintf(&prompt, "**Issue #%d:** %s\n\n", event.Issue.Number, event.Issue.Title)
	fmt.Fprintf(&prompt, "**Description:**\n%s\n\n", event.Issue.Body)

	if conversationContext != "" {
		fmt.Fprintf(&prompt, "**Previous Context:**\n%s\n\n", conversationContext)
	}

	prompt.WriteString("**Task:**\n")
	prompt.WriteString("1. Clone the repository and explore the codebase\n")
	prompt.WriteString("2. Analyze the issue and understand the requirements\n")
	prompt.WriteString("3. Identify the root cause or determine what needs to be implemented\n")
	prompt.WriteString("4. Propose a solution with a detailed implementation plan\n")
	prompt.WriteString(pb.getGitHubInstructions())
	fmt.Fprintf(&prompt, "5. **REQUIRED:** Post your analysis to issue #%d using gh CLI:\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "   a. Write analysis to file: /tmp/comment-%d.md\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "   b. Post comment: gh issue comment %d --repo %s --body-file /tmp/comment-%d.md\n",
		event.Issue.Number, event.Repository.FullName, event.Issue.Number)
	fmt.Fprintf(&prompt, "6. **REQUIRED:** Add label \"awaiting-approval\" to issue #%d using gh CLI:\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "   gh issue edit %d --repo %s --add-label \"awaiting-approval\"\n\n",
		event.Issue.Number, event.Repository.FullName)

	prompt.WriteString("**CRITICAL - You MUST complete these steps:**\n")
	fmt.Fprintf(&prompt, "1. Write comment to file: Use Write tool to create /tmp/comment-%d.md with your analysis\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "2. Post comment: gh issue comment %d --repo %s --body-file /tmp/comment-%d.md\n",
		event.Issue.Number, event.Repository.FullName, event.Issue.Number)
	fmt.Fprintf(&prompt, "3. Add label: gh issue edit %d --repo %s --add-label \"awaiting-approval\"\n\n",
		event.Issue.Number, event.Repository.FullName)

	prompt.WriteString("**Your comment should include:**\n")
	prompt.WriteString("- Summary of the issue\n")
	prompt.WriteString("- Root cause analysis (for bugs) or requirements analysis (for features)\n")
	prompt.WriteString("- Proposed solution approach\n")
	prompt.WriteString("- Implementation plan with affected files and changes\n")
	prompt.WriteString("- Request: 'Please comment \"approved\" to proceed with implementation'\n\n")

	prompt.WriteString("**Important:**\n")
	prompt.WriteString("- Do NOT create a PR yet - wait for approval first\n")
	prompt.WriteString("- Keep your analysis concise but thorough\n")
	prompt.WriteString("- Your task is NOT complete until both gh commands succeed\n")

	return prompt.String()
}

// BuildIssueImplementationPrompt creates a prompt for implementing an approved issue
func (pb *PromptBuilder) BuildIssueImplementationPrompt(event *types.IssueCommentEvent, conversationContext string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a GitHub automation bot for the Kinetik project. An issue has been approved for implementation.\n\n")
	fmt.Fprintf(&prompt, "**Repository:** %s\n", event.Repository.FullName)
	fmt.Fprintf(&prompt, "**Issue #%d:** %s\n\n", event.Issue.Number, event.Issue.Title)

	if conversationContext != "" {
		fmt.Fprintf(&prompt, "**Approved Plan:**\n%s\n\n", conversationContext)
	}

	prompt.WriteString("**Task:**\n")
	prompt.WriteString("1. Implement the approved plan from the issue analysis\n")
	prompt.WriteString("2. Make the necessary code changes using Edit/Write tools\n")
	prompt.WriteString("3. Create a pull request using gh CLI:\n")
	fmt.Fprintf(&prompt, "   gh pr create --repo %s --title \"<title>\" --body \"<description>\" --head <branch> --base main\n",
		event.Repository.FullName)
	prompt.WriteString("4. Link the PR to the issue by commenting:\n")
	fmt.Fprintf(&prompt, "   a. Write comment to file: /tmp/comment-%d.md\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "   b. Post: gh issue comment %d --repo %s --body-file /tmp/comment-%d.md\n\n",
		event.Issue.Number, event.Repository.FullName, event.Issue.Number)

	prompt.WriteString("**Pull Request Guidelines:**\n")
	prompt.WriteString("- Title should be clear and concise (under 70 characters)\n")
	prompt.WriteString("- Description should summarize changes and reference the issue\n")
	prompt.WriteString("- Include test plan if applicable\n")
	prompt.WriteString("- Use conventional commit format (fix:, feat:, refactor:, etc.)\n\n")

	prompt.WriteString("**Important:**\n")
	prompt.WriteString("- Follow the existing code style and patterns in the repository\n")
	prompt.WriteString("- Avoid over-engineering - keep solutions simple and focused\n")
	prompt.WriteString("- Do NOT add unnecessary features or refactoring\n")
	prompt.WriteString("- Use gh CLI via Bash tool for all GitHub interactions\n")

	return prompt.String()
}

// BuildIssueMentionPrompt creates a prompt for handling @bot mentions
func (pb *PromptBuilder) BuildIssueMentionPrompt(event *types.IssueCommentEvent, conversationContext string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a GitHub automation bot for the Kinetik project. You were mentioned in an issue comment.\n\n")
	fmt.Fprintf(&prompt, "**Repository:** %s\n", event.Repository.FullName)
	fmt.Fprintf(&prompt, "**Issue #%d:** %s\n\n", event.Issue.Number, event.Issue.Title)
	fmt.Fprintf(&prompt, "**Issue Description:**\n%s\n\n", event.Issue.Body)
	fmt.Fprintf(&prompt, "**Comment by @%s:**\n%s\n\n", event.Comment.User.Login, event.Comment.Body)

	if conversationContext != "" {
		fmt.Fprintf(&prompt, "**Conversation History:**\n%s\n\n", conversationContext)
	}

	prompt.WriteString("**Task:**\n")
	prompt.WriteString("1. Understand what the user is asking or requesting\n")
	prompt.WriteString("2. Explore the codebase if needed to provide context\n")
	prompt.WriteString("3. Formulate a helpful response addressing their request\n")
	prompt.WriteString(pb.getGitHubInstructions())
	fmt.Fprintf(&prompt, "4. **REQUIRED:** Post your response using gh CLI to issue #%d in repository %s\n\n",
		event.Issue.Number, event.Repository.FullName)

	prompt.WriteString("**CRITICAL - You MUST post a comment using these steps:**\n")
	fmt.Fprintf(&prompt, "1. Write response to file: Use Write tool to create /tmp/comment-%d.md\n", event.Issue.Number)
	fmt.Fprintf(&prompt, "2. Post comment: gh issue comment %d --repo %s --body-file /tmp/comment-%d.md\n",
		event.Issue.Number, event.Repository.FullName, event.Issue.Number)
	prompt.WriteString("Your response is NOT complete until you have successfully posted a comment using gh CLI.\n\n")

	prompt.WriteString("**Guidelines:**\n")
	prompt.WriteString("- Be concise and helpful\n")
	prompt.WriteString("- If asked to implement something, provide an analysis first and request approval\n")
	prompt.WriteString("- Always acknowledge the user by posting a comment\n")

	return prompt.String()
}

// BuildPRReviewPrompt creates a prompt for reviewing PR comments
func (pb *PromptBuilder) BuildPRReviewPrompt(event *types.PullRequestReviewCommentEvent, conversationContext string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a GitHub automation bot for the Kinetik project. A comment was made on a pull request.\n\n")
	fmt.Fprintf(&prompt, "**Repository:** %s\n", event.Repository.FullName)
	fmt.Fprintf(&prompt, "**PR #%d:** %s\n\n", event.PullRequest.Number, event.PullRequest.Title)
	fmt.Fprintf(&prompt, "**Comment by @%s:**\n%s\n\n", event.Comment.User.Login, event.Comment.Body)

	if conversationContext != "" {
		fmt.Fprintf(&prompt, "**PR Context:**\n%s\n\n", conversationContext)
	}

	prompt.WriteString("**Task:**\n")
	prompt.WriteString("1. Understand the feedback or request in the comment\n")
	prompt.WriteString("2. Make the requested changes to the code\n")
	prompt.WriteString("3. Post response using gh CLI (PRs use pr comment command):\n")
	fmt.Fprintf(&prompt, "   a. Write response to file: /tmp/comment-pr-%d.md\n", event.PullRequest.Number)
	fmt.Fprintf(&prompt, "   b. Post comment: gh pr comment %d --repo %s --body-file /tmp/comment-pr-%d.md\n",
		event.PullRequest.Number, event.Repository.FullName, event.PullRequest.Number)
	prompt.WriteString("4. If changes were made, push them to the PR branch\n\n")

	prompt.WriteString("**Important:**\n")
	prompt.WriteString("- Address all feedback in the comment\n")
	prompt.WriteString("- Explain what you changed and why\n")
	prompt.WriteString("- Use gh CLI via Bash tool for all GitHub interactions\n")

	return prompt.String()
}

// BuildPRGeneralReviewPrompt creates a prompt for general PR review
func (pb *PromptBuilder) BuildPRGeneralReviewPrompt(event *types.PullRequestReviewEvent, conversationContext string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a GitHub automation bot for the Kinetik project. A review was submitted on a pull request.\n\n")
	fmt.Fprintf(&prompt, "**Repository:** %s\n", event.Repository.FullName)
	fmt.Fprintf(&prompt, "**PR #%d:** %s\n\n", event.PullRequest.Number, event.PullRequest.Title)
	fmt.Fprintf(&prompt, "**Review by @%s (%s):**\n%s\n\n",
		event.Review.User.Login, event.Review.State, event.Review.Body)

	if conversationContext != "" {
		fmt.Fprintf(&prompt, "**PR Context:**\n%s\n\n", conversationContext)
	}

	prompt.WriteString("**Task:**\n")

	if event.Review.State == "changes_requested" {
		prompt.WriteString("1. Analyze the requested changes\n")
		prompt.WriteString("2. Make the necessary code modifications\n")
		prompt.WriteString("3. Post response using gh CLI:\n")
		fmt.Fprintf(&prompt, "   a. Write response to file: /tmp/comment-pr-%d.md\n", event.PullRequest.Number)
		fmt.Fprintf(&prompt, "   b. Post comment: gh pr comment %d --repo %s --body-file /tmp/comment-pr-%d.md\n",
			event.PullRequest.Number, event.Repository.FullName, event.PullRequest.Number)
		prompt.WriteString("4. Push the changes to the PR branch\n\n")
	} else {
		prompt.WriteString("1. Acknowledge the review\n")
		prompt.WriteString("2. Thank the reviewer using gh CLI:\n")
		fmt.Fprintf(&prompt, "   a. Write message to file: /tmp/comment-pr-%d.md\n", event.PullRequest.Number)
		fmt.Fprintf(&prompt, "   b. Post comment: gh pr comment %d --repo %s --body-file /tmp/comment-pr-%d.md\n\n",
			event.PullRequest.Number, event.Repository.FullName, event.PullRequest.Number)
	}

	prompt.WriteString("**Important:**\n")
	prompt.WriteString("- Be professional and courteous\n")
	prompt.WriteString("- Use gh CLI via Bash tool for all GitHub interactions\n")

	return prompt.String()
}
