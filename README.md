# Kinetik Project

A comprehensive fitness tracking platform with voice agent integration and automated GitHub workflow management.

## Project Structure

This repository is a monorepo containing:

- **[KinetikServer](./KinetikServer)** - Go REST API for fitness tracking (submodule)
- **[kinetik_agent](./kinetik_agent)** - TEN Framework-based voice agent with Agora RTC (submodule)
- **[automation](./automation)** - GitHub automation system with Claude Code integration

## Overview

### KinetikServer
Go-based REST API that provides:
- User authentication and management
- Workout tracking and analytics
- RESTful endpoints for fitness data
- PostgreSQL database integration

### kinetik_agent
Voice agent powered by TEN Framework:
- Real-time voice interaction
- Agora RTC integration
- AI-powered fitness coaching
- Multi-modal communication

### Automation System
Intelligent GitHub automation that:
- Monitors issues and PRs across all repositories
- Automatically analyzes and proposes solutions using Claude Code
- Creates pull requests with approval-based workflow
- Responds to comments and reviews
- Maintains conversation context across interactions

## Quick Start

### Clone Repository with Submodules

```bash
git clone --recurse-submodules https://github.com/TomasBack2Future/Kinetik.git
cd Kinetik
```

If you already cloned without submodules:

```bash
git submodule update --init --recursive
```

### Set Up Automation System

The automation system requires setup of GitHub credentials and database. See detailed instructions:

- **[Setup Guide](./docs/setup.md)** - Step-by-step setup instructions
- **[Architecture](./docs/architecture.md)** - System architecture and design

Quick setup:

```bash
cd automation

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with your credentials

# Start with Docker Compose
docker-compose up -d
```

## Documentation

- [Setup Guide](./docs/setup.md) - Complete setup instructions
- [Architecture](./docs/architecture.md) - System architecture overview
- [Automation README](./automation/README.md) - Automation service documentation

## Features

### Automated Issue Analysis
When a new issue is created, the bot:
1. Analyzes the codebase
2. Identifies root cause or requirements
3. Proposes implementation plan
4. Waits for approval

### Approval-Based PR Creation
After user approval:
1. Bot implements the proposed solution
2. Creates a pull request
3. Links PR to original issue
4. Responds to review feedback

### Context-Aware Responses
The system maintains:
- Conversation history across multiple comments
- Session continuity with Claude
- State machine for workflow tracking
- Full audit trail in PostgreSQL

### Multi-Repository Support
Monitors and responds to events in:
- Main Kinetik repository
- KinetikServer submodule
- kinetik_agent submodule

## Development

### Prerequisites

- Go 1.22+
- Node.js (for Claude CLI)
- PostgreSQL 14+
- Docker & Docker Compose
- GitHub Personal Access Token

### Local Development

```bash
# Start automation service
cd automation
make dev

# Run tests
make test

# View logs
make docker-logs
```

### Testing Webhooks Locally

Use ngrok to expose local server:

```bash
# Start ngrok
ngrok http 8080

# Update GitHub App webhook URL to ngrok URL
# Create test issue to trigger automation
```

## CI/CD

GitHub Actions workflows:
- **Automation CI** - Tests and builds automation service
- Runs on push/PR to main and develop branches
- Automated Docker image builds

## Security

- Webhook HMAC-SHA256 signature validation
- GitHub PAT with fine-grained permissions
- Environment-based secrets management
- Non-root container execution
- Database credential isolation

## Architecture

```
GitHub Repositories
        ↓
   Webhooks
        ↓
Webhook Server (Go)
        ↓
Workflow Orchestrator
        ↓
Claude Code with GitHub MCP
        ↓
Automated Analysis & PRs
```

See [Architecture Documentation](./docs/architecture.md) for details.

## Contributing

1. Create feature branch from `develop`
2. Make changes and test locally
3. Run tests: `make test`
4. Submit pull request
5. Automated checks will run via CI

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: Create an issue in this repository
- **Documentation**: See [docs/](./docs/) directory
- **Questions**: Use GitHub Discussions

## Status

- ✅ KinetikServer - Active development
- ✅ kinetik_agent - Active development
- ✅ Automation System - Implemented and operational

## Roadmap

### Short Term
- [ ] Complete automation system testing
- [ ] Deploy to production
- [ ] Monitor and iterate based on usage

### Medium Term
- [ ] Add support for more GitHub events
- [ ] Implement PR review quality checks
- [ ] Add metrics and monitoring dashboard
- [ ] Support for custom workflow rules

### Long Term
- [ ] Machine learning for issue prioritization
- [ ] Automated dependency updates
- [ ] Integration testing automation
- [ ] Multi-language support for analysis

## Authors

- **TomasBack2Future** - Project maintainer

## Acknowledgments

- Claude Code for AI-powered automation
- Anthropic for Claude API
- GitHub for webhook infrastructure
- TEN Framework for voice agent capabilities
