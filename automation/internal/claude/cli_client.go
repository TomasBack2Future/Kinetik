package claude

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/TomasBack2Future/Kinetik/automation/pkg/config"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/sirupsen/logrus"
)

type CLIClient struct {
	config *config.ClaudeConfig
}

type ExecutionResult struct {
	SessionID    string `json:"session_id"`
	Output       string `json:"output"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	TotalTokens  int    `json:"total_tokens"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	ToolUses     int    `json:"tool_uses,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
}

func NewCLIClient(cfg *config.ClaudeConfig) *CLIClient {
	return &CLIClient{
		config: cfg,
	}
}

// tailFile reads the last n lines from a file
func tailFile(filePath string, n int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read all lines into a slice
	var lines []string
	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line length

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning file: %w", err)
	}

	// Return last n lines
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	result := ""
	for i := start; i < len(lines); i++ {
		result += lines[i] + "\n"
	}

	return result, nil
}

// ParseTokenUsage extracts token usage information from Claude Code output
// Expected format: <usage>total_tokens: 7884\ninput_tokens: 1000\noutput_tokens: 6884\ntool_uses: 3\nduration_ms: 15144</usage>
func ParseTokenUsage(output string) (totalTokens, inputTokens, outputTokens, toolUses int, durationMs int64) {
	// Use regex to find usage block
	usagePattern := `<usage>([\s\S]*?)</usage>`
	re := regexp.MustCompile(usagePattern)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 2 {
		return 0, 0, 0, 0, 0
	}

	usageBlock := matches[1]

	// Parse individual fields
	extractInt := func(pattern string) int {
		re := regexp.MustCompile(pattern + `:\s*(\d+)`)
		matches := re.FindStringSubmatch(usageBlock)
		if len(matches) >= 2 {
			val, _ := strconv.Atoi(matches[1])
			return val
		}
		return 0
	}

	extractInt64 := func(pattern string) int64 {
		re := regexp.MustCompile(pattern + `:\s*(\d+)`)
		matches := re.FindStringSubmatch(usageBlock)
		if len(matches) >= 2 {
			val, _ := strconv.ParseInt(matches[1], 10, 64)
			return val
		}
		return 0
	}

	totalTokens = extractInt("total_tokens")
	inputTokens = extractInt("input_tokens")
	outputTokens = extractInt("output_tokens")
	toolUses = extractInt("tool_uses")
	durationMs = extractInt64("duration_ms")

	return
}

// executeOnce runs Claude CLI once without retry logic
func (c *CLIClient) executeOnce(ctx context.Context, prompt string, sessionID string, workDir string) (*ExecutionResult, error) {
	// Use repository root as work directory to access project-scoped MCP servers
	// The GitHub MCP server was added with "local" scope, so it's only available in the project directory
	if workDir == "" {
		workDir = c.config.RepoRoot
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	promptPreview := prompt
	if len(prompt) > 100 {
		promptPreview = prompt[:100] + "..."
	}

	logger.WithFields(logrus.Fields{
		"work_dir":       workDir,
		"prompt_length":  len(prompt),
		"prompt_preview": promptPreview,
	}).Info("Executing Claude CLI")

	// Write prompt to file for reproducibility and piping
	promptPath := filepath.Join(workDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		return nil, fmt.Errorf("failed to write prompt file: %w", err)
	}
	logger.WithField("prompt_file", promptPath).Info("Saved prompt to file")

	// Build command - pipe prompt from file instead of passing as argument
	args := []string{
		"--print",
		// Note: No --output-format flag - plain text is more readable and faster
		// Note: No --session-id flag - let Claude auto-generate to avoid conflicts
		// Note: No --mcp-config flag needed - MCP servers are loaded from project scope
		// since we're running in the Kinetik repo directory
	}

	// Add verbose flag if debug logging is enabled
	if logger.Log.Level >= logrus.DebugLevel {
		args = append(args, "--verbose")
		logger.Debug("Adding --verbose flag to Claude CLI command")
	}

	logger.WithFields(logrus.Fields{
		"num_args":     len(args),
		"cli_path":     c.config.CLIPath,
		"command_line": fmt.Sprintf("%s %v < prompt.txt", c.config.CLIPath, args),
		"work_dir":     workDir,
		"prompt_file":  promptPath,
	}).Info("Built command arguments with prompt from file")

	// Set timeout
	execCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.config.CLIPath, args...)
	cmd.Dir = workDir
	cmd.Env = c.buildEnv()

	// Open prompt file for piping to stdin
	promptFile, err := os.Open(promptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open prompt file: %w", err)
	}
	defer promptFile.Close()

	// Pipe prompt file to stdin
	cmd.Stdin = promptFile

	// Create log files for stdout and stderr to avoid memory issues with large outputs
	stdoutPath := filepath.Join(workDir, "stdout.log")
	stderrPath := filepath.Join(workDir, "stderr.log")

	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log: %w", err)
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr log: %w", err)
	}
	defer stderrFile.Close()

	// Redirect output to files
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Execute
	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	// Close files to flush buffers
	stdoutFile.Close()
	stderrFile.Close()

	// Read the last 200 lines from stdout (to avoid memory issues)
	stdoutTail, readErr := tailFile(stdoutPath, 200)
	if readErr != nil {
		logger.WithField("error", readErr).Warn("Failed to tail stdout file")
		stdoutTail = fmt.Sprintf("[Error reading stdout: %v]", readErr)
	}

	// Parse token usage from the tail (usage block is typically at the end)
	totalTokens, inputTokens, outputTokens, toolUses, durationMs := ParseTokenUsage(stdoutTail)

	// Read full stderr for error reporting (usually small)
	stderrBytes, _ := os.ReadFile(stderrPath)
	stderrStr := string(stderrBytes)

	// Get file info for logging
	stdoutInfo, _ := os.Stat(stdoutPath)
	stdoutSize := int64(0)
	if stdoutInfo != nil {
		stdoutSize = stdoutInfo.Size()
	}

	// Session ID is no longer extracted since we're not using JSON format
	// It's tracked for logging but not used to resume sessions
	extractedSessionID := ""

	logger.WithFields(logrus.Fields{
		"session_id":       extractedSessionID,
		"duration_ms":      duration.Milliseconds(),
		"success":          err == nil,
		"stdout_file":      stdoutPath,
		"stdout_size_kb":   stdoutSize / 1024,
		"stderr_file":      stderrPath,
		"output_truncated": "last 200 lines",
		"total_tokens":     totalTokens,
		"tool_uses":        toolUses,
	}).Info("Claude CLI execution completed")

	result := &ExecutionResult{
		SessionID:    extractedSessionID,
		Output:       stdoutTail, // Return only last 200 lines
		Success:      err == nil,
		TotalTokens:  totalTokens,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		ToolUses:     toolUses,
		DurationMs:   durationMs,
	}

	if err != nil {
		result.Error = fmt.Sprintf("execution failed: %v\nstderr: %s", err, stderrStr)
		// Log error with file locations
		logger.WithFields(logrus.Fields{
			"error":        err.Error(),
			"stderr":       stderrStr,
			"stdout_file":  stdoutPath,
			"stderr_file":  stderrPath,
			"stdout_size":  stdoutSize,
		}).Error("Claude CLI execution failed")
		return result, err
	}

	// Log success with file location for full output
	logger.WithFields(logrus.Fields{
		"session_id":         extractedSessionID,
		"stdout_file":        stdoutPath,
		"stdout_size_bytes":  stdoutSize,
		"output_tail_length": len(stdoutTail),
	}).Info("Claude execution result (full output in file)")

	return result, nil
}

// Execute runs Claude CLI with the given prompt and options, with retry logic
func (c *CLIClient) Execute(ctx context.Context, prompt string, sessionID string, workDir string) (*ExecutionResult, error) {
	maxRetries := c.config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 2 // Default to 2 retries
	}

	var lastErr error
	var result *ExecutionResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Log retry attempt
			logger.WithFields(logrus.Fields{
				"attempt":     attempt,
				"max_retries": maxRetries,
				"last_error":  lastErr.Error(),
			}).Warn("Retrying Claude CLI execution after failure")

			// Add a small delay before retry (exponential backoff: 2s, 4s, 8s...)
			backoffDuration := time.Duration(1<<uint(attempt-1)) * 2 * time.Second
			logger.WithField("backoff_duration", backoffDuration).Debug("Waiting before retry")

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			case <-time.After(backoffDuration):
				// Continue with retry
			}
		}

		// Attempt execution
		result, lastErr = c.executeOnce(ctx, prompt, sessionID, workDir)

		if lastErr == nil {
			// Success!
			if attempt > 0 {
				logger.WithFields(logrus.Fields{
					"attempt":        attempt + 1,
					"total_attempts": attempt + 1,
				}).Info("Claude CLI execution succeeded after retry")
			}
			return result, nil
		}

		// Check if error is retryable
		if !isRetryableError(lastErr) {
			logger.WithFields(logrus.Fields{
				"error":   lastErr.Error(),
				"attempt": attempt + 1,
			}).Warn("Non-retryable error encountered, not retrying")
			return result, lastErr
		}
	}

	// All retries exhausted
	logger.WithFields(logrus.Fields{
		"total_attempts": maxRetries + 1,
		"final_error":    lastErr.Error(),
	}).Error("Claude CLI execution failed after all retries")

	return result, fmt.Errorf("execution failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Retryable error patterns
	retryablePatterns := []string{
		"killed",                    // Process was killed
		"signal:",                   // Process received signal
		"connection refused",        // Network issues
		"timeout",                   // Timeout errors
		"temporary failure",         // Temporary failures
		"context deadline exceeded", // Context timeout
		"broken pipe",              // Pipe errors
		"connection reset",         // Connection reset
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			logger.WithFields(logrus.Fields{
				"error":   errMsg,
				"pattern": pattern,
			}).Debug("Error matched retryable pattern")
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ExecuteAsync runs Claude CLI asynchronously in the background
func (c *CLIClient) ExecuteAsync(ctx context.Context, prompt string, sessionID string, workDir string) (string, error) {
	// Use repository root as work directory to access project-scoped MCP servers
	if workDir == "" {
		workDir = c.config.RepoRoot
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create work directory: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"work_dir": workDir,
	}).Info("Starting Claude CLI async execution")

	// Build command - prompt MUST come first, then options
	args := []string{
		prompt, // Prompt must be first!
		"--print",
		// Note: No --output-format flag - plain text is more readable and faster
		// Note: No --session-id flag - let Claude auto-generate to avoid conflicts
		// Note: No --mcp-config flag needed - MCP servers are loaded from project scope
		// since we're running in the Kinetik repo directory
	}

	// Add verbose flag if debug logging is enabled
	if logger.Log.Level >= logrus.DebugLevel {
		args = append(args, "--verbose")
		logger.Debug("Adding --verbose flag to Claude CLI async command")
	}

	cmd := exec.CommandContext(ctx, c.config.CLIPath, args...)
	cmd.Dir = workDir
	cmd.Env = c.buildEnv()

	// Create log files for stdout and stderr
	stdoutFile, err := os.Create(filepath.Join(workDir, "stdout.log"))
	if err != nil {
		return "", fmt.Errorf("failed to create stdout log: %w", err)
	}

	stderrFile, err := os.Create(filepath.Join(workDir, "stderr.log"))
	if err != nil {
		stdoutFile.Close()
		return "", fmt.Errorf("failed to create stderr log: %w", err)
	}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start command in background
	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Close files and wait for completion in goroutine
	go func() {
		defer stdoutFile.Close()
		defer stderrFile.Close()

		if err := cmd.Wait(); err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Async Claude CLI execution failed")
		} else {
			logger.Info("Async Claude CLI execution completed")
		}
	}()

	// Return process ID since we don't have session ID until completion
	return fmt.Sprintf("pid-%d", cmd.Process.Pid), nil
}

// buildEnv builds environment variables for Claude CLI
func (c *CLIClient) buildEnv() []string {
	env := os.Environ()

	// Add environment variables from config
	if len(c.config.Env) > 0 {
		for key, value := range c.config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		logger.WithFields(logrus.Fields{
			"num_env_vars": len(c.config.Env),
		}).Debug("Injected environment variables from config")
	}

	return env
}

// GetSessionResult reads the result from a completed async session
func (c *CLIClient) GetSessionResult(sessionID string) (*ExecutionResult, error) {
	workDir := filepath.Join(c.config.WorkDir, sessionID)

	stdoutPath := filepath.Join(workDir, "stdout.log")
	stderrPath := filepath.Join(workDir, "stderr.log")

	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	stderr, _ := os.ReadFile(stderrPath)

	result := &ExecutionResult{
		SessionID: sessionID,
		Output:    string(stdout),
		Success:   len(stderr) == 0,
	}

	if len(stderr) > 0 {
		result.Error = string(stderr)
	}

	return result, nil
}
