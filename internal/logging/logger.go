package logger

import "github.com/sirupsen/logrus"

// Logger is common logger for all deepalert code.
var Logger = logrus.New()

func init() {
	Logger.SetLevel(logrus.InfoLevel)
}
