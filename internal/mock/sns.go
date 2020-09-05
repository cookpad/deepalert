package mock

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/deepalert/internal/adaptor"
)

// NewSNSClient creates mock SNS client
func NewSNSClient(region string) adaptor.SNSClient {
	return &SNSClient{region: region}
}

// SNSClient is mock
type SNSClient struct {
	region string
	input  []*sns.PublishInput
}

// Publish of mock SNSClient only stores sns.PublishInput
func (x *SNSClient) Publish(*sns.PublishInput) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{}, nil
}
