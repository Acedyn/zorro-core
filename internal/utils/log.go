package utils

import (
  "sync"
  "log/slog"
)

var logger *slog.Logger
var once sync.Once

// Getter for the logger singleton
func Logger() *slog.Logger {
  once.Do(func() {
    // TODO: Define a log handler
    logger = slog.Default()
  })

  return logger
}
