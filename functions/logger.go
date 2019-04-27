package functions

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

// Logger is a global logger for functions
var Logger *logrus.Entry

func setupLogger() {
	loggerBase := logrus.New()
	loggerBase.SetLevel(logrus.InfoLevel)
	loggerBase.SetFormatter(&logrus.JSONFormatter{})

	Logger = loggerBase.WithFields(logrus.Fields{})
}

// SetLoggerContext binds context and global logger.
func SetLoggerContext(ctx context.Context) {
	lc, _ := lambdacontext.FromContext(ctx)
	Logger = Logger.WithField("request_id", lc.AwsRequestID)
}
