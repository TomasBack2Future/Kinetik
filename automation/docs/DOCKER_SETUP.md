# Docker Setup Guide

This guide helps you build and run the Kinetik Automation service locally using Docker Compose with ngrok tunneling.

## Prerequisites

- Docker and Docker Compose installed
- Ngrok account (free tier works) - [Sign up here](https://ngrok.com/)
- GitHub Personal Access Token
- Anthropic API Key

## Quick Start

### 1. Build the Docker Image

```bash
cd automation
./build-local.sh
```

Or manually:
```bash
cd automation
docker build -t kinetik-automation:latest .
```

### 2. Configure Environment Variables

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit `.env` and fill in your credentials:

```env
# GitHub Configuration
GITHUB_WEBHOOK_SECRET=your-webhook-secret-here
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_xxxxxxxxxxxxx

# Anthropic Configuration
ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxx

# Ngrok Configuration
NGROK_AUTHTOKEN=your-ngrok-authtoken-here
```

**Getting your Ngrok Auth Token:**
1. Sign up at [ngrok.com](https://ngrok.com/)
2. Go to [Dashboard > Your Authtoken](https://dashboard.ngrok.com/get-started/your-authtoken)
3. Copy your authtoken

### 3. Start the Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database on port 5433
- Webhook server on port 8080
- Ngrok tunnel (web interface on port 4040)

### 4. Get Your Public Webhook URL

Visit the ngrok web interface:
```bash
open http://localhost:4040
```

Or get the URL via API:
```bash
curl http://localhost:4040/api/tunnels | jq '.tunnels[0].public_url'
```

Copy this URL (e.g., `https://xxxx-xxx-xxx-xxx.ngrok-free.app`) and use it to configure your GitHub webhook.

### 5. Configure GitHub Webhook

1. Go to your GitHub repository settings
2. Navigate to **Settings > Webhooks > Add webhook**
3. Set Payload URL: `https://your-ngrok-url.ngrok-free.app/webhook/github`
4. Set Content type: `application/json`
5. Set Secret: Same value as `GITHUB_WEBHOOK_SECRET` in your `.env`
6. Select events: Choose `Issues`, `Issue comments`, or `Let me select individual events`
7. Click **Add webhook**

## Managing the Services

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f webhook-server
docker-compose logs -f ngrok
```

### Stop Services

```bash
docker-compose down
```

### Stop and Remove Volumes

```bash
docker-compose down -v
```

### Rebuild and Restart

```bash
./build-local.sh
docker-compose up -d --force-recreate
```

## Troubleshooting

### Issue: "Error: image not found"

Make sure you've built the image first:
```bash
./build-local.sh
```

### Issue: Ngrok tunnel not working

Check if your authtoken is correct:
```bash
docker-compose logs ngrok
```

### Issue: Webhook server can't connect to database

Wait a few seconds for PostgreSQL to be ready, then check:
```bash
docker-compose logs postgres
docker-compose logs webhook-server
```

### Issue: Claude CLI authentication

Make sure your `~/.claude` directory exists and contains valid credentials:
```bash
ls -la ~/.claude
```

## Network Architecture

```
GitHub Webhook
    ↓
Ngrok (public URL)
    ↓
Webhook Server (port 8080)
    ↓
PostgreSQL (port 5432)
```

## Useful Commands

```bash
# Check running containers
docker-compose ps

# Access webhook server shell
docker-compose exec webhook-server sh

# Access PostgreSQL
docker-compose exec postgres psql -U kinetik_automation -d kinetik_automation

# View ngrok web interface
open http://localhost:4040

# Restart a single service
docker-compose restart webhook-server
```

## Production Deployment

For production, consider:
1. Using a proper domain with SSL instead of ngrok
2. Setting up proper secret management (AWS Secrets Manager, HashiCorp Vault)
3. Using a managed PostgreSQL instance
4. Implementing proper logging and monitoring
5. Setting up health checks and auto-restart policies
