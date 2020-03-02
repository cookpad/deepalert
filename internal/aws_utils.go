package functions

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

// PublishSNS sends general data
func PublishSNS(topicArn, region string, data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(topicArn),
	})

	Logger.WithField("response", resp).Info("Done SNS Publish")

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

func SNStoMessages(snsEvent events.SNSEvent) [][]byte {
	var messages [][]byte
	for _, record := range snsEvent.Records {
		messages = append(messages, []byte(record.SNS.Message))
	}
	return messages
}

func SQStoMessage(sqsEvent events.SQSEvent) [][]byte {
	var messages [][]byte
	for _, record := range sqsEvent.Records {
		messages = append(messages, []byte(record.Body))
	}
	return messages
}

func ExecDelayMachine(stateMachineARN, region string, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := sfn.New(ssn)

	input := sfn.StartExecutionInput{
		Input:           aws.String(string(raw)),
		StateMachineArn: aws.String(stateMachineARN),
	}

	if _, err := svc.StartExecution(&input); err != nil {
		return errors.Wrapf(err, "Fail to execute delay machine: %s %s", stateMachineARN, string(raw))
	}

	return nil
}
