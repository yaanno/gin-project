package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config represents the logger configuration
type Config struct {
	Level       string
	FilePath    string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	EnableFile  bool
	Development bool
}

// New creates a new logger with the specified configuration
func New(cfg Config) zerolog.Logger {
	// Multiple writers
	var writers []io.Writer

	// Always add console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	writers = append(writers, consoleWriter)

	// Optional file logging
	if cfg.EnableFile {
		// Ensure logs directory exists
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			panic(err)
		}

		// Configure log rotation
		rotatingWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge, // days
		}
		writers = append(writers, rotatingWriter)
	}

	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure logger
	logger := zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return logger
}

// Default returns a default configured logger
func Default() zerolog.Logger {
	defaultConfig := Config{
		Level:       "debug",
		FilePath:    "logs/app.log",
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      28,
		EnableFile:  true,
		Development: false,
	}
	return New(defaultConfig)
}

func init() {
	var _ = New(Config{
		Level:       "debug",
		FilePath:    "logs/app.log",
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      28,
		EnableFile:  true,
		Development: false,
	})
}
