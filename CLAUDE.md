# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Purpose

The primary goal of this project is **learning** — not just shipping features. The user is building their first AI-integrated application while learning Go from scratch. Every implementation decision should be treated as a teaching opportunity. Prefer clarity and explainability over cleverness or brevity.

## Reference Examples

The `examples/` directory contains reference implementations that represent best practices for this project. **Always consult these before writing new code.**

- `examples/prjstart/` — Clean architecture baseline: the canonical Go project structure (cmd, config, models, db, handlers, services). Use this as the structural reference for all new code.
- `examples/flashcards/` — Full flashcard domain implementation with AI features (MCP server, agent handler, quiz, notes, memory). Reference this for domain logic and AI integration patterns.
- `examples/examples/basic-usage/` — Minimal Claude API usage
- `examples/examples/function-calling/` — Tool use / function calling with Claude
- `examples/examples/llm-streaming/` — Streaming responses from Claude
- `examples/examples/codeagent/` — Building an agent with tools
- `examples/examples/mcp/` — MCP server setup (local and HTTP)
- `examples/examples/vector-search/` — Vector search / semantic retrieval

When implementing any new feature, check whether a relevant example exists in `examples/` first and follow its patterns. If you deviate from an example's pattern, explain why.

## Common Development Commands

**Build and Run:**
- `go build ./cmd/main.go` - Build the application (after making changes, always run this to verify compilation)
- `make build` - Build with proper binary name (creates `todo-api`)
- `make run` - Run the application directly
- `make clean` - Clean build artifacts

**Database Operations:**
- `make db-start` - Start Supabase local PostgreSQL environment
- `make db-up` - Run database migrations
- `make db-stop` - Stop Supabase environment
- `make db-reset` - Complete database reset (stop, start, migrate)

## Architecture Overview

This is a **dual-purpose REST API** serving both **todos and flashcards** with identical architectural patterns.

**Clean Architecture Pattern:**
- `cmd/main.go` - Application entry point with dependency injection
- `models/` - Data structures and DTOs (todo.go, flashcard.go)
- `handlers/` - HTTP request/response handling (todoHandler.go, flashcardHandler.go) 
- `services/` - Business logic and validation (todoService.go, flashcardService.go)
- `db/` - Repository pattern with PostgreSQL implementation (todoDb.go, flashcardDb.go)
- `config/` - Environment-based configuration management

**Key Patterns:**
- Both todos and flashcards follow **identical architectural patterns**
- Repository interfaces for database abstraction
- Service layer handles validation and business logic
- Handlers manage HTTP concerns (JSON, status codes, error responses)
- Gorilla Mux for routing with pattern-based routes (`/todos/{id:[0-9]+}`, `/flashcards/{id:[0-9]+}`)

**Database:**
- PostgreSQL with schema `gocourse.todos` and `gocourse.flashcards`
- Migrations in `supabase/migrations/` with timestamp prefixes
- Local development via Supabase CLI (accessible at localhost:54322)

**Data Models:**
- **Todos**: ID, Title, Description, Completed, CreatedAt, UpdatedAt
- **Flashcards**: ID, Content, CreatedAt, UpdatedAt (minimal content-only design)

**When adding new entities**, follow the exact same layered pattern: model → migration → repository → service → handler → route registration in main.go.

## Environment Setup

Requires `.env` file with:
- `DB_URL` - PostgreSQL connection string
- `PORT` - Application port (defaults to 8080)

Both Docker and Supabase CLI must be installed and running for database operations.

## Development Workflow Guidance

- After each task is complete, go build the project to verify your work and then fix any issues which come up.

## Teaching & Explanation Rules

This is the user's **first AI project** and they are **new to Go**. Always follow these rules:

**Explain every decision:**
- Before writing or changing code, explain *why* this approach is chosen over alternatives.
- After writing code, explain what each significant part does in plain language — not just what, but *why* it's structured that way.
- When using Go-specific syntax or idioms (interfaces, goroutines, defer, error handling, structs, etc.), stop and explain them as if the user has never seen them before.

**Go language explanations:**
- Never assume familiarity with Go syntax. Treat every Go concept as new.
- Compare Go patterns to general programming concepts when helpful (e.g., "a Go interface is like a contract — any type that has these methods satisfies it").
- Explain Go error handling explicitly — it is very different from languages with exceptions.
- Explain why Go uses certain patterns (e.g., why we use interfaces for the repository layer, what dependency injection means here).

**Architecture explanations:**
- When adding a new file or layer, explain its role in the overall architecture and why clean architecture separates concerns this way.
- Explain what would go wrong if we skipped a layer (e.g., "if we put database code directly in the handler, we'd have trouble testing it and changing databases later").

**Decision transparency:**
- If there are multiple valid approaches, briefly mention the tradeoffs and explain which one is chosen and why.
- If something is a Go convention or best practice, say so explicitly.