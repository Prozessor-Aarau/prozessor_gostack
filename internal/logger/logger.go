package logger

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// DBHandler is a custom slog.Handler that writes log entries to both stdout and database.
// It wraps another handler for stdout logging and asynchronously writes to a database table.
type DBHandler struct {
	db   *sql.DB      // Database connection for logging
	next slog.Handler // Next handler in the chain (typically for stdout)
}

// NewDBHandler creates a new DBHandler that logs to both the provided handler and database.
// The db parameter should be a valid database connection.
// The next parameter is typically a TextHandler or JSONHandler for stdout logging.
func NewDBHandler(db *sql.DB, next slog.Handler) *DBHandler {
	return &DBHandler{
		db:   db,
		next: next,
	}
}

// Handle processes a log record by first passing it to the next handler (stdout),
// then asynchronously writing it to the database.
// Database logging is done in a goroutine to avoid blocking the main execution.
func (h *DBHandler) Handle(ctx context.Context, r slog.Record) error {
	// Log to stdout/file first
	if err := h.next.Handle(ctx, r); err != nil {
		return err
	}

	// Log to database asynchronously to avoid blocking
	go h.logToDB(r)
	return nil
}

// logToDB writes a log record to the database.
// It extracts the visitor_id attribute if present and determines the calling package.
// Uses PostgreSQL parameter placeholders ($1, $2, etc.).
func (h *DBHandler) logToDB(r slog.Record) {
	// Extract visitor_id from log attributes
	visitorID := ""
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "visitor_id" {
			visitorID = a.Value.String()
			return false // stop iteration
		}
		return true // continue iteration
	})

	// Get the package name where the log call originated
	packageName := h.getCallerPackage()

	// Handle NULL visitor_id for UUID database column
	var visitorIDParam any
	if visitorID == "" {
		visitorIDParam = nil
	} else {
		visitorIDParam = visitorID
	}

	// Insert log entry into database
	query := `INSERT INTO log (type, module, message, visitor_id, created_on) VALUES ($1, $2, $3, $4, $5)`
	_, err := h.db.Exec(query, r.Level.String(), packageName, r.Message, visitorIDParam, time.Now())
	if err != nil {
		// Use standard log to avoid recursion with slog
		log.Printf("failed to log to database: %v", err)
	}
}

// getCallerPackage extracts the package name from the call stack.
// It skips several frames to get to the actual caller that initiated the log call.
// Returns "unknown" if the package name cannot be determined.
func (h *DBHandler) getCallerPackage() string {
	// Skip frames: runtime.Callers, getCallerPackage, logToDB, Handle, slog internals
	const skip = 6
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return "unknown"
	}

	frame, _ := runtime.CallersFrames(pc).Next()
	if frame.Function == "" {
		return "unknown"
	}

	// Extract package name from function name
	// e.g., "github.com/user/project/internal/handlers.(*Repository).SomeHandler" -> "handlers"
	parts := strings.Split(frame.Function, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if dotIndex := strings.Index(lastPart, "."); dotIndex != -1 {
			return lastPart[:dotIndex]
		}
	}

	return "unknown"
}

// Enabled reports whether the handler handles records at the given level.
// Delegates to the next handler in the chain.
func (h *DBHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// WithAttrs returns a new handler with the given attributes added.
// The attributes are passed to the next handler in the chain.
func (h *DBHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DBHandler{
		db:   h.db,
		next: h.next.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the given group name.
// The group is passed to the next handler in the chain.
func (h *DBHandler) WithGroup(name string) slog.Handler {
	return &DBHandler{
		db:   h.db,
		next: h.next.WithGroup(name),
	}
}

// LogError is a convenience function for logging errors with additional attributes.
// It automatically adds the error as an attribute and logs at ERROR level.
func LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	allAttrs := append([]slog.Attr{slog.String("error", err.Error())}, attrs...)
	slog.LogAttrs(ctx, slog.LevelError, msg, allAttrs...)
}
