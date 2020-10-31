package logging

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger is common logger for all deepalert code.
var Logger = logrus.New()

func init() {
	Logger.SetLevel(logrus.InfoLevel)
}

func SetLogLevel(logLevel string) {
	level := logrus.InfoLevel
	switch strings.ToUpper(logLevel) {
	case "TRACE":
		level = logrus.TraceLevel
	case "DEBUG":
		level = logrus.DebugLevel
	case "INFO":
		level = logrus.InfoLevel
	case "ERROR":
		level = logrus.ErrorLevel
	case "FATAL":
		level = logrus.FatalLevel
	case "":
		// nohting to do
	default:
		Logger.WithField("logLevel", logLevel).Warn("Invalid log level, set to INFO")
	}
	Logger.SetLevel(level)
}
