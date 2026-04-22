module github.com/cookpad/deepalert/test/workflow

go 1.26

toolchain go1.26.1

replace github.com/cookpad/deepalert => ../..

require (
	github.com/aws/aws-lambda-go v1.53.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/cookpad/deepalert v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/guregu/dynamo v1.23.0
	github.com/m-mizutani/golambda v1.1.3
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/getsentry/sentry-go v0.43.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/m-mizutani/goerr v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
