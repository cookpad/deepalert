package errors

// Flush emits all errors
func Flush() {
	flushSentryError()
}
