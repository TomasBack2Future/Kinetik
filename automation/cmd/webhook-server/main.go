package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TomasBack2Future/Kinetik/automation/internal/claude"
	contextmgr "github.com/TomasBack2Future/Kinetik/automation/internal/context"
	"github.com/TomasBack2Future/Kinetik/automation/internal/github"
	"github.com/TomasBack2Future/Kinetik/automation/internal/handlers"
	"github.com/TomasBack2Future/Kinetik/automation/internal/middleware"
	"github.com/TomasBack2Future/Kinetik/automation/internal/repository"
	"github.com/TomasBack2Future/Kinetik/automation/internal/workflow"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/config"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize logger with basic config first
	logger.Init("info")
	logger.Info("Starting Kinetik GitHub Automation Webhook Server")

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
	}

	// Reinitialize logger with file output if enabled
	if cfg.Logging.Enabled && cfg.Logging.File != "" {
		if err := logger.InitWithFile(cfg.Logging.Level, cfg.Logging.File); err != nil {
			logger.Fatal("Failed to initialize file logging", err)
		}
		logger.Info("File logging enabled")
	}

	// Initialize database
	db, err := repository.NewPostgresDB(cfg.Database.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", err)
		}
	}()

	logger.Info("Connected to database")

	// Initialize repositories
	conversationRepo := repository.NewConversationRepo(db)

	// Initialize context manager
	contextManager := contextmgr.NewManager(conversationRepo)

	// Initialize Claude client
	claudeClient := claude.NewCLIClient(&cfg.Claude)

	// Initialize prompt builder
	promptBuilder := claude.NewPromptBuilder("kinetik-bot")

	// Initialize GitHub client
	githubClient := github.NewClient(cfg)

	// Initialize workflow orchestrator
	orchestrator := workflow.NewOrchestrator(cfg, claudeClient, promptBuilder, contextManager, githubClient)

	// Initialize webhook handler
	webhookHandler := handlers.NewWebhookHandler(cfg, orchestrator)

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "OK")
	})

	// Webhook endpoint with signature validation
	r.Route("/github", func(r chi.Router) {
		r.Use(middleware.ValidateGitHubWebhook(cfg.GitHub.WebhookSecret))
		r.Post("/webhook", webhookHandler.Handle)
	})

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server listening on %s", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server stopped")
}
