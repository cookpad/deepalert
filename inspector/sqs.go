package inspector

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

type sqsClient interface {
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

var newSqsClient = newAwsSqsClient

func newAwsSqsClient(region string) sqsClient {
	ssn := session.New(&aws.Config{Region: aws.String(region)})
	client := sqs.New(ssn)
	return client
}

// Sample: https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/some-queue-name
var regexSqsURL = regexp.MustCompile(`https://sqs.([a-z0-9-]+).amazonaws.com`)

func extractRegionFromURL(url string) (*string, error) {
	if m := regexSqsURL.FindStringSubmatch(url); len(m) == 2 {
		return &m[1], nil
	}
	return nil, fmt.Errorf("Invalid SQS URL foramt: %v", url)
}

func sendSQS(msg interface{}, targetURL string) error {
	region, err := extractRegionFromURL(targetURL)
	if err != nil {
		return err
	}

	client := newSqsClient(*region)

	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal message: %v", msg)
	}

	input := sqs.SendMessageInput{
		QueueUrl:    &targetURL,
		MessageBody: aws.String(string(raw)),
	}
	resp, err := client.SendMessage(&input)

	if err != nil {
		return errors.Wrapf(err, "Fail to send SQS message: %v", input)
	}

	Logger.WithField("resp", resp).Trace("Sent SQS message")

	return nil
}
