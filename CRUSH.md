# Prozessor GoStack Development Guide

## Build & Development Commands
- **Build CSS**: `npm run build` (builds Tailwind CSS)
- **Watch CSS**: `npm run watch` (watches Tailwind CSS changes)
- **Generate Templates**: `templ generate` (generates Go code from .templ files)
- **Build App**: `go build -C ./cmd -o ../tmp/main .`
- **Full Build**: `npm run build && templ generate && go build -C ./cmd -o ../tmp/main .`
- **Dev Server**: `air` (hot reload with Air)
- **Run Tests**: `go test ./...` (no tests currently exist)
- **Single Test**: `go test -run TestName ./path/to/package`

## Code Style Guidelines
- **Package Structure**: Use internal/ for private packages, cmd/ for main applications
- **Imports**: Group standard library, third-party, then local imports with blank lines
- **Naming**: Use PascalCase for exported types/functions, camelCase for unexported
- **Error Handling**: Always check errors, use structured logging with slog
- **Types**: Define structs in models/ package, use time.Time for dates
- **Templates**: Use .templ files with templ generate, not Go templates
- **Database**: Use repository pattern with interfaces in repository/ package
- **Handlers**: Use Make() wrapper for error handling, rendering() for template rendering
- **Context**: Pass context through request chain, store session data in context
- **Configuration**: Use config.AppConfig struct, load env vars with godotenv
- **Logging**: Use slog for structured logging, separate info/error loggers
- **Sessions**: Use scs/v2 for session management
- **Middleware**: Chain middleware in routes.go, use chi router
- **Static Files**: Serve from static/ directory, generate assets to static/css/

## Project Structure
- `cmd/`: Main application entry point
- `internal/`: Private application code (handlers, models, repository, etc.)
- `view/`: Templ template files (.templ)
- `static/`: Static assets (CSS, images, generated files)
- `migrations/`: Database migration files