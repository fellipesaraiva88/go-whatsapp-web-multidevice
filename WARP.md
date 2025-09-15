# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Essential Development Commands

### Building and Running

```bash
# Development - run directly
cd src && go run . rest            # Start REST API server
cd src && go run . mcp             # Start MCP server for AI agents
cd src && go run . --help          # Show all available commands and flags

# Production builds
cd src && go build -o whatsapp     # Build binary (Linux/macOS)
cd src && go build -o whatsapp.exe # Build binary (Windows)
cd src && ./whatsapp rest          # Run REST mode
cd src && ./whatsapp mcp           # Run MCP mode

# Docker development
docker-compose up -d --build       # Build and run with Docker
docker-compose logs whatsapp       # View application logs
```

### Testing and Quality

```bash
cd src && go test ./...                    # Run all tests
cd src && go test ./validations           # Run specific package tests
cd src && go test -cover ./...            # Run tests with coverage
cd src && go test -race ./...             # Run with race detector
cd src && go test -v ./validations        # Run specific package with verbose output

cd src && go fmt ./...                    # Format all code
cd src && go vet ./...                    # Static analysis
cd src && go mod tidy                     # Clean dependencies
cd src && go mod download                 # Download dependencies
```

### Live Development

```bash
# Using Air for live reload (if installed: go install github.com/air-verse/air@latest)
cd src && air                            # Uses .air.toml config for hot reload
```

## Architecture Overview

This Go application implements a WhatsApp Web API server with dual operational modes: REST API and Model Context Protocol (MCP) server. It's built using **Domain-Driven Design** and **Clean Architecture** principles.

### Core Architecture Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                      Command Layer                          │
├─────────────────────────────────────────────────────────────┤
│  cmd/root.go (Cobra CLI) │ cmd/rest.go │ cmd/mcp.go        │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    UI/Transport Layer                       │
├─────────────────────────────────────────────────────────────┤
│ ui/rest/     │ ui/websocket/ │ ui/mcp/                     │
│ (Fiber HTTP) │ (Real-time)   │ (AI Agent Protocol)         │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Use Case Layer                           │
├─────────────────────────────────────────────────────────────┤
│ usecase/ - Application services bridging domains and UI     │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
├─────────────────────────────────────────────────────────────┤
│ domains/app/  │ domains/user/  │ domains/message/           │
│ domains/chat/ │ domains/group/ │ domains/send/              │
│ domains/newsletter/ - Business logic and entities          │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                        │
├─────────────────────────────────────────────────────────────┤
│ infrastructure/whatsapp/  │ infrastructure/chatstorage/    │
│ (WhatsApp Protocol)       │ (Data Persistence)             │
└─────────────────────────────────────────────────────────────┘
```

### Key Directory Structure

- `src/cmd/` - CLI commands using Cobra (root, rest, mcp)
- `src/domains/` - Domain models organized by business context (app, chat, group, message, user, newsletter, send)
- `src/infrastructure/` - External integrations (WhatsApp client via whatsmeow, database operations)
- `src/ui/` - Interface layers (REST API with Fiber, WebSocket hub, MCP server)
- `src/usecase/` - Application services implementing business workflows
- `src/validations/` - Input validation using ozzo-validation
- `src/views/` - HTML templates and frontend assets (embedded in binary)

### Critical Dependencies

- `go.mau.fi/whatsmeow` - WhatsApp Web multi-device protocol implementation
- `github.com/gofiber/fiber/v2` - HTTP framework for REST API mode
- `github.com/mark3labs/mcp-go` - Model Context Protocol server for AI agents
- `github.com/spf13/cobra` + `github.com/spf13/viper` - CLI and configuration management

## Configuration Management

### Priority Order
1. Command-line flags (highest)
2. Environment variables
3. `.env` file in `src/` directory (lowest)

### Environment Configuration

Copy `src/.env.example` to `src/.env` and customize:

```bash
# Essential Variables
APP_PORT=3000                           # Server port
APP_DEBUG=true                          # Enable debug logging  
APP_OS=Chrome                           # WhatsApp device name
APP_BASIC_AUTH=admin:password           # HTTP basic auth (comma-separated for multiple users)
APP_BASE_PATH=/api/v1                   # Subpath deployment support

# Database Configuration
DB_URI="file:storages/whatsapp.db?_foreign_keys=on"     # Main WhatsApp connection database
DB_KEYS_URI="file::memory:?cache=shared&_foreign_keys=on" # Crypto keys database (separate for security)

# WhatsApp Behavior
WHATSAPP_AUTO_REPLY="I'm currently away"  # Auto-reply message
WHATSAPP_AUTO_MARK_READ=true              # Auto-mark messages as read
WHATSAPP_WEBHOOK=https://your-webhook.com # Webhook URL for events (supports multiple URLs)
WHATSAPP_WEBHOOK_SECRET=your-secret-key   # HMAC validation secret
WHATSAPP_ACCOUNT_VALIDATION=true          # Enable account validation
WHATSAPP_CHAT_STORAGE=true               # Enable chat history storage
```

### Runtime Mode Differences

**REST Mode (`./whatsapp rest`)**:
- Fiber web server with HTML UI at `http://localhost:3000`
- WebSocket support for real-time updates
- Complete API surface (60+ endpoints)
- Built-in QR code display and device management UI

**MCP Mode (`./whatsapp mcp`)**:
- Model Context Protocol server for AI agent integration
- Server-Sent Events (SSE) transport at `http://localhost:8080/sse`
- Limited to core messaging tools (`whatsapp_send_text`, `whatsapp_send_contact`, etc.)
- Designed for programmatic AI agent interaction

## WhatsApp Integration Details

### Multi-Device Protocol
- Uses `whatsmeow` library implementing WhatsApp's multi-device protocol
- Session data stored in SQLite database (`storages/whatsapp.db`)
- Supports both QR code and pairing code authentication
- Auto-reconnection with exponential backoff

### Media Handling
- **FFmpeg Required**: Install via `brew install ffmpeg` (macOS), `sudo apt install ffmpeg` (Linux)
- Auto-compression for images and videos before sending
- Media files stored in `src/statics/media/`
- QR codes generated in `src/statics/qrcode/`

### Storage Architecture
- **Main Database**: Connection state, device info (`storages/whatsapp.db`)
- **Chat Storage**: Message history, optional (`storages/chatstorage.db`)
- **Keys Database**: Cryptographic keys, can be in-memory for security

### Important Limitations
- **Mutually Exclusive Modes**: Cannot run REST and MCP simultaneously due to whatsmeow library constraints
- **Single WhatsApp Account**: Each instance connects to one WhatsApp account
- **No Concurrent Sessions**: Only one active session per WhatsApp account

## Development Workflow Tips

### Local Development Setup
```bash
# 1. Set up environment
cd src && cp .env.example .env
# Edit .env with your preferences

# 2. Install FFmpeg (required for media processing)
brew install ffmpeg                     # macOS
sudo apt update && sudo apt install ffmpeg  # Linux

# 3. Start development server
go run . rest                          # or `go run . mcp` for MCP mode
```

### WhatsApp Account Setup
1. Visit `http://localhost:3000` (REST mode)
2. Click "Login with QR Code" 
3. Scan with WhatsApp mobile app (Settings > Linked Devices)
4. Monitor logs for connection status

### Testing Strategy
- **Unit Tests**: Focus on `validations/` and `pkg/utils/` packages
- **Integration Tests**: Use test WhatsApp account (recommended)
- **Manual Testing**: Use built-in HTML interface for REST API
- **MCP Testing**: Use MCP-compatible clients like Cursor IDE

### Debugging Common Issues

**Connection Problems**:
```bash
# Enable debug logging
go run . rest --debug=true

# Check database state
sqlite3 src/storages/whatsapp.db ".schema"
```

**Media Upload Failures**:
```bash
# Verify FFmpeg installation
ffmpeg -version

# Check media directory permissions
ls -la src/statics/media/
```

**Webhook Issues**:
```bash
# Test webhook with curl (check HMAC signature)
curl -X POST https://your-webhook.com \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{"event": "test"}'
```

### Configuration for Different Environments

**Development**:
```bash
go run . rest --debug=true --port=3000 --autoreply="Dev mode"
```

**Production Docker**:
```yaml
# Use provided docker-compose.yml as template
environment:
  - APP_DEBUG=false
  - APP_BASIC_AUTH=admin:secure_password
  - WHATSAPP_WEBHOOK=https://your-production-webhook.com
```

**MCP Integration (e.g., Cursor IDE)**:
```json
{
  "mcpServers": {
    "whatsapp": {
      "url": "http://localhost:8080/sse"
    }
  }
}
```