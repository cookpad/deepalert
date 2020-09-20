package adaptor

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
)

// SFnClientFactory is interface SFnClient constructor
type SFnClientFactory func(region string) (SFnClient, error)

// SFnClient is interface of AWS SDK SQS
type SFnClient interface {
	StartExecution(*sfn.StartExecutionInput) (*sfn.StartExecutionOutput, error)
}

// NewSFnClient creates actual AWS SFn SDK client
func NewSFnClient(region string) (SFnClient, error) {
	ssn, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return sfn.New(ssn), nil
}
