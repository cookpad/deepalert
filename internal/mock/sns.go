package mock

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/deepalert/deepalert/internal/adaptor"
)

// NewSNSClient creates mock SNS client
func NewSNSClient(region string) (adaptor.SNSClient, error) {
	return &SNSClient{Region: region}, nil
}

// SNSClient is mock
type SNSClient struct {
	Region string
	Input  []*sns.PublishInput
}

// Publish of mock SNSClient only stores sns.PublishInput
func (x *SNSClient) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	x.Input = append(x.Input, input)
	return &sns.PublishOutput{}, nil
}

// NewMockSNSClientSet returns a pair of SNSClient and SNSClientFactory
func NewMockSNSClientSet() (*SNSClient, adaptor.SNSClientFactory) {
	client := &SNSClient{}
	return client, func(region string) (adaptor.SNSClient, error) {
		client.Region = region
		return client, nil
	}
}
