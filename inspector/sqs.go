package inspector

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/deepalert/deepalert/internal/errors"
)

// SQSClient is interface of AWS SDK SQS. Need to have only SendMessage()
type SQSClient interface {
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

// SQSClientFactory is constructor of SQSClient with region
type SQSClientFactory func(region string) (SQSClient, error)

func newAwsSQSClient(region string) (SQSClient, error) {
	ssn, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	client := sqs.New(ssn)
	return client, nil
}

// Sample: https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/some-queue-name
var regexSqsURL = regexp.MustCompile(`https://sqs.([a-z0-9-]+).amazonaws.com`)

func extractRegionFromURL(url string) (*string, error) {
	if m := regexSqsURL.FindStringSubmatch(url); len(m) == 2 {
		return &m[1], nil
	}
	return nil, fmt.Errorf("Invalid SQS URL foramt: %v", url)
}

func sendSQS(newSQS SQSClientFactory, msg interface{}, targetURL string) error {
	region, err := extractRegionFromURL(targetURL)
	if err != nil {
		return err
	}

	client, err := newSQS(*region)
	if err != nil {
		return errors.Wrap(err, "Failed to create a new SQS client")
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal message").With("msg", msg)
	}

	input := sqs.SendMessageInput{
		QueueUrl:    &targetURL,
		MessageBody: aws.String(string(raw)),
	}
	resp, err := client.SendMessage(&input)

	if err != nil {
		return errors.Wrap(err, "Failed to send SQS message").With("input", input)
	}

	Logger.WithField("resp", resp).Trace("Sent SQS message")

	return nil
}
