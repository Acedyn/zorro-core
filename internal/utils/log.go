package utils

import (
	"log/slog"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

// Getter for the logger singleton
func Logger() *slog.Logger {
	once.Do(func() {
		// TODO: Define a log handler
		logger = slog.Default()
	})

	return logger
}
