package inspector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/m-mizutani/golambda"
)

// InspectHandler is a function type of callback of inspector.
type InspectHandler func(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error)

// Logger is github.com/m-mizutani/golambda logger and exported to be controlled from external module.
var Logger = golambda.Logger

type reportIDKey struct{}

var contextKey = &reportIDKey{}

// ReportIDFromCtx extracts ReportID from context. The function is available in handler called by Start
func ReportIDFromCtx(ctx context.Context) (*deepalert.ReportID, bool) {
	lc, ok := ctx.Value(contextKey).(*deepalert.ReportID)
	return lc, ok
}

// Arguments is parameters to invoke Start().
type Arguments struct {
	Context context.Context

	Tasks []*deepalert.Task

	// Handler is callback function of Start(). Handler mayu be called multiply. (Required)
	Handler InspectHandler

	// HandlerData is data for Handler. deepalert/inspector never access HandlerData and set additional argument if you need in Handler (optional)
	HandlerData interface{}

	// Author indicates owner of new attributes and contents. It does not require explicit unique name, but unique name helps your debugging and troubleshooting. (Required)
	Author string

	// AttrQueueURL is URL to send new attributes discovered inspector (e.g. a new related IP address). It should be exported CloudFormation value and can be imported by Fn::ImportValue: + YOU_STACK_NAME-AttributeQueue to your inspector CloudFormation stack. (Required)
	AttrQueueURL string

	// FindingQueueURL is also URL to send contents generated inspector (e.g. IP address is blacklisted or not). It should be exported CloudFormation value and can be imported by Fn::ImportValue: + YOU_STACK_NAME-ContentQueue to your inspector CloudFormation stack. (Required)
	FindingQueueURL string

	// NewSQS is constructor of SQSClient that is interface of AWS SDK. This function is to set stub for testing. If NewSQS is nil, use default constructor, newAwsSQSClient. (Optional)
	NewSQS SQSClientFactory
}

// Start is a wrapper of Inspector.
func Start(args Arguments) error {
	for _, task := range args.Tasks {
		if err := HandleTask(args.Context, task, args); err != nil {
			return err
		}
	}

	return nil
}

// SNSEventToTasks extracts deepalert.Task from SNS Event
func SNSEventToTasks(event events.SNSEvent) ([]*deepalert.Task, error) {
	var results []*deepalert.Task

	for _, record := range event.Records {
		var task deepalert.Task
		msg := record.SNS.Message
		if err := json.Unmarshal([]byte(msg), &task); err != nil {
			return nil, golambda.WrapError(err, "Fail to unmarshal task").With("msg", msg)
		}

		results = append(results, &task)
	}

	return results, nil
}

// HandleTask is called with task by task. It's exported for testing
func HandleTask(ctx context.Context, task *deepalert.Task, args Arguments) error {
	Logger.
		With("task", task).
		With("ctx", ctx).
		With("Author", args.Author).
		With("AttrQueueURL", args.AttrQueueURL).
		With("FindingQueueURL", args.FindingQueueURL).
		Info("Start inspector")

	// Check Arguments
	if args.Handler == nil {
		return fmt.Errorf("Handler is not set in inspector.Argument")
	}
	if args.Author == "" {
		return fmt.Errorf("Author is not set in inspector.Argument")
	}
	if args.AttrQueueURL == "" {
		return fmt.Errorf("AttrQueueURL is not set in inspector.Argument")
	}
	if args.FindingQueueURL == "" {
		return fmt.Errorf("FindingQueueURL is not set in inspector.Argument")
	}
	if task == nil {
		return fmt.Errorf("Task is nil")
	}

	if args.NewSQS == nil {
		args.NewSQS = newAwsSQSClient
	}

	newCtx := context.WithValue(ctx, contextKey, &task.ReportID)

	result, err := args.Handler(newCtx, *task.Attribute)
	if err != nil {
		return golambda.WrapError(err, "Fail to handle task").With("task", task)
	}

	if result == nil {
		return nil
	}

	// Sending entities
	for _, entity := range result.Contents {
		finding := deepalert.Finding{
			ReportID:  task.ReportID,
			Attribute: *task.Attribute,
			Author:    args.Author,
			Type:      entity.Type(),
			Content:   entity,
		}
		Logger.With("finding", finding).Trace("Sending finding")

		if err := sendSQS(args.NewSQS, finding, args.FindingQueueURL); err != nil {
			return golambda.WrapError(err, "Fail to publish ReportContent").With("url", args.FindingQueueURL).With("finding", finding)
		}
	}

	var newAttrs []*deepalert.Attribute
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
			OriginAttr: *task.Attribute,
			Attributes: newAttrs,
			Author:     args.Author,
		}

		Logger.With("ReportAttribute", attrReport).Trace("Sending new attributes")
		if err := sendSQS(args.NewSQS, attrReport, args.AttrQueueURL); err != nil {
			return golambda.WrapError(err, "Fail to publish ReportAttribute").With("url", args.AttrQueueURL).With("report", attrReport)
		}
	}

	Logger.Trace("Exit handler normally")
	return nil
}
