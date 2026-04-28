package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	Log zerolog.Logger
)

// InitLogger initializes the global logger with structured logging
func InitLogger(logLevel string, logToFile bool, logFilePath string) error {
	// Set log level
	level := zerolog.InfoLevel
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	// Configure output writers
	var writers []io.Writer

	// Console output with pretty formatting for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}
	writers = append(writers, consoleWriter)

	// File output if enabled
	if logToFile && logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		writers = append(writers, file)
	}

	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)

	// Initialize global logger
	Log = zerolog.New(multiWriter).With().
		Timestamp().
		Caller().
		Logger()

	Log.Info().Msg("Logger initialized successfully")
	return nil
}

// LogInfo logs informational messages
func LogInfo(message string, fields map[string]interface{}) {
	event := Log.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// LogWarn logs warning messages
func LogWarn(message string, fields map[string]interface{}) {
	event := Log.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// LogError logs error messages
func LogError(message string, err error, fields map[string]interface{}) {
	event := Log.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// LogDebug logs debug messages
func LogDebug(message string, fields map[string]interface{}) {
	event := Log.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// LogFatal logs fatal messages and exits
func LogFatal(message string, err error, fields map[string]interface{}) {
	event := Log.Fatal().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// LogAudit logs audit trail events (for security and compliance)
func LogAudit(action string, userID string, resource string, details map[string]interface{}) {
	event := Log.Info().
		Str("type", "audit").
		Str("action", action).
		Str("user_id", userID).
		Str("resource", resource)
	
	for k, v := range details {
		event = event.Interface(k, v)
	}
	
	event.Msg("Audit log")
}

// LogSecurity logs security-related events (failed logins, unauthorized access)
func LogSecurity(event string, userID string, ipAddress string, details map[string]interface{}) {
	logEvent := Log.Warn().
		Str("type", "security").
		Str("event", event).
		Str("user_id", userID).
		Str("ip_address", ipAddress)
	
	for k, v := range details {
		logEvent = logEvent.Interface(k, v)
	}
	
	logEvent.Msg("Security event")
}

// LogPerformance logs performance metrics
func LogPerformance(operation string, duration time.Duration, details map[string]interface{}) {
	event := Log.Info().
		Str("type", "performance").
		Str("operation", operation).
		Dur("duration_ms", duration)
	
	for k, v := range details {
		event = event.Interface(k, v)
	}
	
	event.Msg("Performance metric")
}

// LogBPM logs BPM-specific events (workflow execution, task completion)
func LogBPM(event string, workflowID string, processID string, details map[string]interface{}) {
	logEvent := Log.Info().
		Str("type", "bpm").
		Str("event", event).
		Str("workflow_id", workflowID).
		Str("process_id", processID)
	
	for k, v := range details {
		logEvent = logEvent.Interface(k, v)
	}
	
	logEvent.Msg("BPM event")
}
