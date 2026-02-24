package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	GitHub   GitHubConfig   `yaml:"github"`
	Claude   ClaudeConfig   `yaml:"claude"`
	Workflow WorkflowConfig `yaml:"workflow"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

type GitHubConfig struct {
	WebhookSecret        string   `yaml:"webhook_secret"`
	PersonalAccessToken  string   `yaml:"personal_access_token"`
	BotUsername          string   `yaml:"bot_username"`
	AllowedRepos         []string `yaml:"allowed_repos"`
}

type ClaudeConfig struct {
	CLIPath    string            `yaml:"cli_path"`
	WorkDir    string            `yaml:"work_dir"`
	RepoRoot   string            `yaml:"repo_root"`
	Timeout    time.Duration     `yaml:"timeout"`
	MaxRetries int               `yaml:"max_retries"` // Maximum number of retries on failure (default: 2)
	Env        map[string]string `yaml:"env"`         // Environment variables to inject into Claude CLI subprocess
}

type WorkflowConfig struct {
	ApprovalKeywords []string `yaml:"approval_keywords"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Enabled bool   `yaml:"enabled"`
}

// Load reads configuration from YAML file and environment variables
func Load(configPath string) (*Config, error) {
	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	cfg.overrideFromEnv()

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// overrideFromEnv overrides configuration values with environment variables
func (c *Config) overrideFromEnv() {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		c.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		c.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		c.Database.Name = dbName
	}
	if webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET"); webhookSecret != "" {
		c.GitHub.WebhookSecret = webhookSecret
	}
	if pat := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"); pat != "" {
		c.GitHub.PersonalAccessToken = pat
	}
}

// validate checks if required configuration values are present
func (c *Config) validate() error {
	if c.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.GitHub.WebhookSecret == "" {
		return fmt.Errorf("github webhook secret is required")
	}
	if c.GitHub.PersonalAccessToken == "" {
		return fmt.Errorf("github personal access token is required")
	}
	if len(c.GitHub.AllowedRepos) == 0 {
		return fmt.Errorf("at least one allowed repository is required")
	}
	if c.Claude.CLIPath == "" {
		return fmt.Errorf("claude CLI path is required")
	}
	if c.Claude.RepoRoot == "" {
		return fmt.Errorf("claude repo root is required")
	}
	// Set default max retries if not specified
	if c.Claude.MaxRetries == 0 {
		c.Claude.MaxRetries = 2
	}
	return nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, sslMode,
	)
}
