package logging

import (
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// Init configures the global logrus instance with the specified log level and formatting.
//
// Parameters:
//   - level: A string representing the desired log level (e.g., "debug", "info", "warn", "error").
//     The level is case-insensitive. If an invalid level is provided, it defaults to "info".
//
// The function sets up the logger with the following configurations:
//   - Log level: Parsed from the input string, defaulting to Info if parsing fails.
//   - Formatter: TextFormatter with full timestamp and custom timestamp format.
//
// This function does not return any value; it modifies the global logger in-place.
func Init(level string) {
	lvl, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
}

// L returns the configured logger (convenience).
func L() *logrus.Logger { return logger }
