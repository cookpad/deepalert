package adaptor

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// SNSClientFactory is interface SNSClient constructor
type SNSClientFactory func(region string) SNSClient

// SNSClient is interface of AWS SDK SQS
type SNSClient interface {
	Publish(*sns.PublishInput) (*sns.PublishOutput, error)
}

// NewSNSClient creates actual AWS SNS SDK client
func NewSNSClient(region string) SNSClient {
	ssn := session.New(&aws.Config{Region: aws.String(region)})
	return sns.New(ssn)
}
