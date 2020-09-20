package errors

import (
	"log"
	"os"
	"time"

	sentry "github.com/getsentry/sentry-go"
)

var sentryEnabled = false

func initSentryErrorHandler() {
	if os.Getenv("SENTRY_DSN") != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: "",
			// Debug: true,
		})
		if err != nil {
			log.Fatalf("Failed sentry.Init: %+v", err)
		}
		sentryEnabled = true
	}
}

func handleSentryError(err *Error) {
	if sentryEnabled {
		eventID := sentry.CaptureException(err)
		if eventID != nil {
			// Add sentry eventID to original error
			_ = err.With("sentry.eventID", eventID)
		}
	}
}

func flushSentryError() {
	if sentryEnabled {
		sentry.Flush(2 * time.Second)
	}
}
