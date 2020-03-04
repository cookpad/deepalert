package inspector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/deepalert"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// InspectHandler is a function type of callback of inspector.
type InspectHandler func(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error)

// Logger is github.com/sirupsen/logrus logger and exported to be controlled from external module.
var Logger = logrus.New()

func init() {
	// If running as Lambda function
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		Logger.SetFormatter(&logrus.JSONFormatter{})
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.WarnLevel)
		Logger.SetFormatter(&logrus.TextFormatter{})
	}
}

type reportIDKey struct{}

var contextKey = &reportIDKey{}

// ReportIDFromCtx extracts ReportID from context. The function is available in handler called by Start
func ReportIDFromCtx(ctx context.Context) (*deepalert.ReportID, bool) {
	lc, ok := ctx.Value(contextKey).(*deepalert.ReportID)
	return lc, ok
}

// Arguments is parameters to invoke Start(). All arguments are required.
type Arguments struct {
	// Handler is callback function of Start(). Handler mayu be called multiply.
	Handler InspectHandler

	// Author indicates owner of new attributes and contents. It does not require explicit unique name, but unique name helps your debugging and troubleshooting.
	Author string

	// AttrQueueURL is URL to send new attributes discovered inspector (e.g. a new related IP address). It should be exported CloudFormation value and can be imported by Fn::ImportValue: + YOU_STACK_NAME-AttributeQueue to your inspector CloudFormation stack.
	AttrQueueURL string

	// ContentQueueURL is also URL to send contents generated inspector (e.g. IP address is blacklisted or not). It should be exported CloudFormation value and can be imported by Fn::ImportValue: + YOU_STACK_NAME-ContentQueue to your inspector CloudFormation stack.
	ContentQueueURL string
}

// Start is a wrapper of Inspector.
func Start(args Arguments) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		Logger.WithFields(logrus.Fields{
			"event": event,
			"args":  args,
		}).Info("Start inspector")

		// Check Arguments
		if args.Handler == nil {
			return fmt.Errorf("Handler is not set in emitter.Argument")
		}
		if args.Author == "" {
			return fmt.Errorf("Author is not set in emitter.Argument")
		}
		if args.AttrQueueURL == "" {
			return fmt.Errorf("AttrQueueURL is not set in emitter.Argument")
		}
		if args.ContentQueueURL == "" {
			return fmt.Errorf("ContentQueueURL is not set in emitter.Argument")
		}

		for _, record := range event.Records {
			var task deepalert.Task
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &task); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal task: %s", msg)
			}

			if err := handleTask(ctx, args, task); err != nil {
				return err
			}
		}

		Logger.Info("Exit inspector normally")
		return nil
	})
}

func handleTask(ctx context.Context, args Arguments, task deepalert.Task) error {
	Logger.WithField("task", task).Trace("Start handler")
	newCtx := context.WithValue(ctx, contextKey, &task.ReportID)

	result, err := args.Handler(newCtx, task.Attribute)
	if err != nil {
		return errors.Wrapf(err, "Fail to handle task: %v", task)
	}

	if result == nil {
		return nil
	}

	// Sending entities
	for _, entity := range result.Contents {
		section := deepalert.ReportSection{
			ReportID:  task.ReportID,
			Attribute: task.Attribute,
			Author:    args.Author,
			Type:      entity.Type(),
			Content:   entity,
		}
		Logger.WithField("section", section).Trace("Sending section")

		if err := sendSQS(section, args.ContentQueueURL); err != nil {
			return errors.Wrapf(err, "Fail to publish ReportContent to %s: %v", args.ContentQueueURL, section)
		}
	}

	var newAttrs []deepalert.Attribute
	for _, attr := range result.NewAttributes {
		if attr.Timestamp == nil {
			attr.Timestamp = task.Attribute.Timestamp
		}
		newAttrs = append(newAttrs, attr)
	}

	// Sending new attributes
	if len(result.NewAttributes) > 0 {
		attrReport := deepalert.ReportAttribute{
			ReportID:   task.ReportID,
			Original:   task.Attribute,
			Attributes: newAttrs,
			Author:     args.Author,
		}

		Logger.WithField("ReportAttribute", attrReport).Trace("Sending new attributes")
		if err := sendSQS(attrReport, args.AttrQueueURL); err != nil {
			return errors.Wrapf(err, "Fail to publish ReportAttribute to %s: %v", args.AttrQueueURL, attrReport)
		}
	}

	Logger.Trace("Exit handler normally")
	return nil
}
