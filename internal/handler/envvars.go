package handler

import "github.com/Netflix/go-env"

// EnvVars has all environment variables that should be given to Lambda function
type EnvVars struct {
	// From arguments
	TaskTopic   string `env:"TASK_TOPIC"`
	ReportTopic string `env:"REPORT_TOPIC"`
	CacheTable  string `env:"CACHE_TABLE"`

	// Only recvAlert can use because of dependency
	InspectorMashine string `env:"INSPECTOR_MACHINE"`
	ReviewMachine    string `env:"REVIEW_MACHINE"`

	// Utilities
	SentryDSN string `env:"SENTRY_DSN"`
	SentryEnv string `env:"SENTRY_ENVIRONMENT"`
	LogLevel  string `env:"LOG_LEVEL"`

	// From AWS Lambda
	AwsRegion string `env:"AWS_REGION"`
}

// BindEnvVars loads environments variables and set them to EnvVars
func (x *EnvVars) BindEnvVars() error {
	if _, err := env.UnmarshalFromEnviron(x); err != nil {
		Logger.WithError(err).Error("Failed UnmarshalFromEviron")
		return err
	}

	return nil
}
