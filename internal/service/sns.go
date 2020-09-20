package service

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/deepalert/internal/adaptor"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// SNSService is accessor to SQS
type SNSService struct {
	newSNS adaptor.SNSClientFactory
}

// NewSNSService is constructor of
func NewSNSService(newSNS adaptor.SNSClientFactory) *SNSService {
	return &SNSService{
		newSNS: newSNS,
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
func (x *SNSService) Publish(topicARN string, msg interface{}) error {
	region, daErr := extractSNSRegion(topicARN)
	if daErr != nil {
		return daErr
	}

	client, err := x.newSNS(region)
	if err != nil {
		return err
	}

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
