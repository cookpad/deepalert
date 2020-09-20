package inspector

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/deepalert"
	"github.com/pkg/errors"
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

func (x *MockSQSClient) GetSections(url string) ([]*deepalert.ReportSection, error) {
	queues, ok := x.InputMap[url]
	if !ok {
		return nil, nil
	}

	var output []*deepalert.ReportSection
	for _, q := range queues {
		var section deepalert.ReportSection
		if err := json.Unmarshal([]byte(*q.MessageBody), &section); err != nil {
			return nil, errors.Wrap(err, "Failed to parse section queue")
		}

		output = append(output, &section)
	}

	return output, nil
}

func (x *MockSQSClient) GetAttributes(url string) ([]*deepalert.ReportAttribute, error) {
	queues, ok := x.InputMap[url]
	if !ok {
		return nil, nil
	}

	var output []*deepalert.ReportAttribute
	for _, q := range queues {
		var attr deepalert.ReportAttribute
		if err := json.Unmarshal([]byte(*q.MessageBody), &attr); err != nil {
			return nil, errors.Wrap(err, "Failed to parse attribute queue")
		}

		output = append(output, &attr)
	}

	return output, nil
}
