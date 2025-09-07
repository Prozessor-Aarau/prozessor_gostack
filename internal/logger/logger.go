package logger

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
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
func (h *DBHandler) logToDB(r slog.Record) {
	// Extract visitor_id and path from log attributes
	visitorID := ""
	requestPath := ""
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "visitor_id":
			visitorID = a.Value.String()
		case "path":
			requestPath = a.Value.String()
		}
		return true // continue iteration
	})

	// Handle NULL visitor_id for UUID database column
	var visitorIDParam any
	if visitorID == "" {
		visitorIDParam = nil
	} else {
		visitorIDParam = visitorID
	}

	// Use "system" if no path
	if requestPath == "" {
		requestPath = "system"
	}

	// Insert log entry into database
	query := `INSERT INTO log (level, location, message, visitor_id, created_on) VALUES ($1, $2, $3, $4, $5)`
	_, err := h.db.Exec(query, r.Level.String(), requestPath, r.Message, visitorIDParam, time.Now())
	if err != nil {
		log.Printf("failed to log to database: %v", err)
	}
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

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[37m"
)

type ColoredHandler struct {
	writer io.Writer
	level  slog.Level
}

func NewColoredHandler(w io.Writer, level slog.Level) *ColoredHandler {
	return &ColoredHandler{
		writer: w,
		level:  level,
	}
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006-01-02 15:04:05")

	var levelColor, levelText string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = ColorGray
		levelText = "DEBUG"
	case slog.LevelInfo:
		levelColor = ColorBlue
		levelText = "INFO"
	case slog.LevelWarn:
		levelColor = ColorYellow
		levelText = "WARN"
	case slog.LevelError:
		levelColor = ColorRed
		levelText = "ERROR"
	default:
		levelColor = ColorReset
		levelText = r.Level.String()
	}

	fmt.Fprintf(h.writer, "%s [%s%s%s] %s\n",
		timestamp, levelColor, levelText, ColorReset, r.Message)

	return nil
}

func (h *ColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

// LogError is a convenience function for logging errors with additional attributes.
// It automatically adds the error as an attribute and logs at ERROR level.
func LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	allAttrs := append([]slog.Attr{slog.String("error", err.Error())}, attrs...)
	slog.LogAttrs(ctx, slog.LevelError, msg, allAttrs...)
}
