package service

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/deepalert/internal/adaptor"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// SNSService is accessor to SQS
type SNSService struct {
	newSQS   adaptor.SNSClientFactory
	queueMap map[string]*sqs.ReceiveMessageOutput
	msgIndex int
}

// NewSNSService is constructor of
func NewSNSService(newSQS adaptor.SNSClientFactory) *SNSService {
	return &SNSService{
		queueMap: make(map[string]*sqs.ReceiveMessageOutput),
		newSQS:   newSQS,
	}
}

func extractSNSRegion(topicARN string) (string, *errors.Error) {
	// topicARN sample: arn:aws:sns:us-east-1:111122223333:my-topic
	arnParts := strings.Split(topicARN, ":")

	if len(arnParts) != 6 {
		return "", errors.New("Invalid SNS topic ARN").With("ARN", topicARN)
	}

	return arnParts[3], nil
}

// Publish is wrapper of sns:Publish of AWS
func (x *SNSService) Publish(msg interface{}, topicARN string) *errors.Error {
	region, daErr := extractSNSRegion(topicARN)
	if daErr != nil {
		return daErr
	}

	client := x.newSQS(region)

	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal message: %v", msg)
	}

	input := sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(raw)),
	}
	resp, err := client.Publish(&input)

	if err != nil {
		return errors.Wrapf(err, "Fail to send SQS message").With("input", input)
	}

	logger.WithField("resp", resp).Trace("Sent SQS message")

	return nil
}
