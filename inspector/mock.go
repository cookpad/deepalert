package inspector

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// MockSQSClient is for testing. Just storing sqs.SendMessageInput
type MockSQSClient struct {
	InputMap map[string][]*sqs.SendMessageInput
	Region   string
}

func newMockSQSClient() *MockSQSClient {
	return &MockSQSClient{
		InputMap: make(map[string][]*sqs.SendMessageInput),
	}
}

// NewSQSMock creates a pair of MockSQSClient and constructor that returns the MockSQSClient
func NewSQSMock() (*MockSQSClient, SQSClientFactory) {
	client := newMockSQSClient()
	return client, func(_ string) (SQSClient, error) {
		return client, nil
	}
}

// SendMessage stores input to own struct, not sending. Just for test.
func (x *MockSQSClient) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	url := aws.StringValue(input.QueueUrl)
	if _, ok := x.InputMap[url]; !ok {
		x.InputMap[url] = []*sqs.SendMessageInput{}
	}
	x.InputMap[url] = append(x.InputMap[url], input)
	return &sqs.SendMessageOutput{}, nil
}
