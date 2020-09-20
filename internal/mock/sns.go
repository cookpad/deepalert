package mock

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/deepalert/internal/adaptor"
)

// NewSNSClient creates mock SNS client
func NewSNSClient(region string) (adaptor.SNSClient, error) {
	return &SNSClient{region: region}, nil
}

// SNSClient is mock
type SNSClient struct {
	region string
	input  []*sns.PublishInput
}

// Publish of mock SNSClient only stores sns.PublishInput
func (x *SNSClient) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	x.input = append(x.input, input)
	return &sns.PublishOutput{}, nil
}
