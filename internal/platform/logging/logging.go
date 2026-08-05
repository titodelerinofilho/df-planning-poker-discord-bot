package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

const unknownEnvironment = "unknown"

type Config struct {
	Environment string
	Level       string
	Version     string
	Stdout      io.Writer
	Stderr      io.Writer
}

type Handlers struct {
	exceptions      *slog.Logger
	requests        *slog.Logger
	responses       *slog.Logger
	databaseQueries *slog.Logger
}

func NewHandlers(cfg Config) (Handlers, error) {
	level, err := parseLevel(cfg.Level)

	if err != nil {
		return Handlers{}, err
	}

	environment := strings.TrimSpace(cfg.Environment)

	if environment == "" {
		environment = unknownEnvironment
	}

	baseAttrs := []slog.Attr{
		slog.String("version", cfg.Version),
		slog.String("environment", environment),
	}

	handlers := Handlers{
		exceptions:      newLogger(cfg.Stderr, level, "exceptions", baseAttrs),
		requests:        newLogger(cfg.Stdout, level, "requests", baseAttrs),
		responses:       newLogger(cfg.Stdout, level, "responses", baseAttrs),
		databaseQueries: newLogger(cfg.Stdout, level, "database_queries", baseAttrs),
	}

	return handlers, nil
}

func NewBootstrapHandlers(stderr io.Writer, version string) Handlers {
	baseAttrs := []slog.Attr{
		slog.String("version", version),
		slog.String("environment", unknownEnvironment),
	}

	return Handlers{
		exceptions: newLogger(stderr, slog.LevelInfo, "exceptions", baseAttrs),
	}
}

func (handlers Handlers) CriticalException(ctx context.Context, correlationIdentifier string, content any) {
	handlers.exceptions.ErrorContext(
		ctx,
		"exception",
		slog.String("correlation_identifier", correlationIdentifier),
		slog.Any("content", content),
	)
}

func (handlers Handlers) InfoRequest(ctx context.Context, correlationIdentifier string, content any) {
	handlers.requests.InfoContext(
		ctx,
		"request",
		slog.String("correlation_identifier", correlationIdentifier),
		slog.Any("content", content),
	)
}

func (handlers Handlers) InfoResponse(ctx context.Context, correlationIdentifier string, content any) {
	handlers.responses.InfoContext(
		ctx,
		"response",
		slog.String("correlation_identifier", correlationIdentifier),
		slog.Any("content", content),
	)
}

func (handlers Handlers) InfoDatabaseQuery(ctx context.Context, correlationIdentifier string, content any) {
	handlers.databaseQueries.InfoContext(
		ctx,
		"database_query",
		slog.String("correlation_identifier", correlationIdentifier),
		slog.Any("content", content),
	)
}

func LogStartupError(ctx context.Context, handlers Handlers, correlationIdentifier string, err error) {
	handlers.CriticalException(ctx, correlationIdentifier, map[string]string{
		"operation": "startup",
		"error":     err.Error(),
	})
}

func newLogger(output io.Writer, level slog.Level, channel string, baseAttrs []slog.Attr) *slog.Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceAttr,
	})

	attrs := make([]slog.Attr, 0, len(baseAttrs)+1)
	attrs = append(attrs, baseAttrs...)
	attrs = append(attrs, slog.String("channel", channel))

	return slog.New(handler).With(attrsToAny(attrs)...)
}

func replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.TimeKey:
		attr.Key = "timestamp"
	case slog.MessageKey:
		attr.Key = "message"
	case slog.LevelKey:
		level := attr.Value.String()

		if level == "ERROR" {
			level = "CRITICAL"
		}

		attr.Value = slog.StringValue(level)
	}

	return attr
}

func attrsToAny(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs))

	for _, attr := range attrs {
		args = append(args, attr)
	}

	return args
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", value)
	}
}
