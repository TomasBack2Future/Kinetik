# 🏋️ Kinetik Repository Overview

Hi @TomasBack2Future! I've completed a thorough analysis of the Kinetik repository. Here's a comprehensive breakdown:

## 🎯 What is Kinetik?

Kinetik is an **AI-powered fitness tracking platform** that combines:
- 🖥️ **REST API Backend** (Go) for fitness data management
- 🎙️ **Real-time Voice Coach** (Go + Python) using AI
- 🤖 **GitHub Automation Bot** for intelligent development workflows

## 📦 Repository Structure (Monorepo)

```
Kinetik/
├── KinetikServer/      # Git submodule - Backend API (Go)
├── kinetik_agent/      # Git submodule - Voice agent (Go + Python)
├── automation/         # GitHub automation system (Go)
├── docs/              # Documentation
└── .github/           # CI/CD workflows
```

## 🔧 Technology Stack

### KinetikServer (Backend)
- **Language**: Go 1.24
- **Database**: PostgreSQL 14+ with pgvector
- **Key Features**:
  - JWT authentication
  - 873+ exercises from wger database
  - Nutrition tracking (54+ items)
  - Workout session management
  - AI-powered search (full-text + vector)
  - Agent lifecycle management
  - LLM-powered workout analysis

### kinetik_agent (Voice Coach)
- **Languages**: Go (TEN Framework runtime) + Python (agent logic)
- **Real-time**: Agora RTC (voice) + RTM (messaging)
- **AI**: LangChain + OpenAI for coaching
- **Key Features**:
  - Multi-stage workout guidance (icebreaker → pre-training → mid-training → post-training)
  - Smart tools: timers, countdown, exercise search, workout plans, history
  - Memory-enabled personalized coaching (pgvector)
  - Speech-to-text and text-to-speech integration

### Automation System
- **Language**: Go 1.22
- **Tools**: Claude Code + GitHub CLI + PostgreSQL
- **Key Features**:
  - Monitors issues/PRs via webhooks
  - Analyzes issues using Claude Code + GitHub MCP
  - Proposes implementation plans
  - Creates PRs after approval
  - Responds to code review comments
  - Maintains conversation context

## 🔄 How It All Works Together

### User Fitness Journey:
```
1. User registers → KinetikServer API
2. User starts voice session → Agora RTC → kinetik_agent
3. Agent provides coaching → Searches exercises/nutrition → KinetikServer API
4. Agent creates workout plan → Guides through exercises
5. Session saved → KinetikServer → PostgreSQL
6. Memories extracted → pgvector embeddings → Long-term personalization
```

### Development Automation:
```
1. Issue created → Webhook → Automation Server
2. Claude Code analyzes codebase → Posts analysis comment
3. User approves → Claude implements → Creates PR
4. PR comments → Claude responds and updates code
5. Context preserved in PostgreSQL for continuity
```

### Agent Lifecycle:
```
1. KinetikServer allocates agent instance
2. Generates configuration → Launches TEN Framework app
3. Agent sends heartbeat every 30s
4. Orphan detection cleans up stale agents
5. Session ends → Agent deallocated
```

## 🎨 Architecture Highlights

✅ **Modular Design**: Submodules can be developed independently
✅ **Scalability**: Components scale separately
✅ **AI-Native**: LLM integration throughout the stack
✅ **Real-Time**: Low-latency voice interaction via Agora
✅ **Context-Aware**: PostgreSQL JSONB + pgvector for rich context
✅ **Cloud-Ready**: Docker Compose, health checks, graceful shutdown

## 📊 Current Development Status

- ✅ **KinetikServer**: Core features implemented
- 🚧 **kinetik_agent**: Voice Agent Phase 2 in progress
- ✅ **Automation**: Operational and handling workflows
- 🔜 **MCP Server**: Planned for KinetikServer (Phase 7)
- 🔜 **Advanced Memory**: Memory extraction in development

## 🔗 Integration Points

**KinetikServer ↔ kinetik_agent**:
- Agent configuration and lifecycle management
- Exercise/nutrition search API
- Workout session storage
- Memory persistence

**Automation ↔ GitHub**:
- Webhook event processing
- GitHub MCP for API access
- Automated code analysis and PR creation

## 📚 Key Files to Explore

- `KinetikServer/main.go` - Backend entry point
- `kinetik_agent/ten_packages/extension/fitness_agent_python/` - Voice agent logic
- `automation/internal/handler/handler.go` - Automation webhook handler
- `docs/` - Comprehensive documentation

---

**Need help with anything specific?** Feel free to ask about:
- Setting up the development environment
- Understanding a specific component
- Implementing new features
- Architecture decisions

Happy to dive deeper into any area! 🚀
