package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/m-mizutani/deepalert"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type cloudWatchLogsHook struct {
	logGroup          string
	region            string
	requestID         string
	reportID          deepalert.ReportID
	funcName          string
	cwlogs            *cloudwatchlogs.CloudWatchLogs
	nextSequenceToken *string
}

func (x *cloudWatchLogsHook) logStream() string {
	if x.funcName != "" && x.requestID != "" && x.reportID != "" {
		return fmt.Sprintf("%s/%s/%s", x.funcName, x.requestID, x.reportID)
	}
	if x.funcName != "" && x.requestID != "" {
		return fmt.Sprintf("%s/%s", x.funcName, x.requestID)
	}
	if x.funcName != "" {
		return fmt.Sprintf("%s/all", x.funcName)
	}
	return "all"
}

// Logger is a global logger for functions
var Logger *logrus.Entry
var loggerHook cloudWatchLogsHook
var loggerBase *logrus.Logger

func setupLogger() {
	loggerBase = logrus.New()
	loggerBase.SetLevel(logrus.InfoLevel)
	loggerBase.SetFormatter(&logrus.JSONFormatter{})

	loggerHook = cloudWatchLogsHook{
		logGroup: os.Getenv("LOG_GROUP"),
		funcName: os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		region:   os.Getenv("AWS_REGION"),
	}
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(loggerHook.region),
	}))
	loggerHook.cwlogs = cloudwatchlogs.New(ssn)

	loggerBase.AddHook(&loggerHook)
	Logger = loggerBase.WithFields(logrus.Fields{})
}

// SetLoggerContext binds context and global logger.
func SetLoggerContext(ctx context.Context, reportID deepalert.ReportID) {
	if ctx != nil {
		lc, _ := lambdacontext.FromContext(ctx)
		loggerHook.requestID = lc.AwsRequestID
	}

	Logger = Logger.WithFields(logrus.Fields{
		"request_id":    loggerHook.requestID,
		"function_name": loggerHook.funcName,
	})

	SetLoggerReportID(reportID)
}

// SetLoggerReportID changes only ReportID of cloudWatchLogsHook.
func SetLoggerReportID(reportID deepalert.ReportID) {
	loggerHook.reportID = reportID
	Logger = Logger.WithField("report_id", reportID)
}

func setLogOutput(output io.Writer) {
	loggerBase.SetOutput(output)
}
func getLogOutput() io.Writer {
	return loggerBase.Out
}

func (x *cloudWatchLogsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func getCloudWatchLogsNextToken(cwlogs *cloudwatchlogs.CloudWatchLogs, logGroup, logStream string) (*string, error) {
	resp, err := cwlogs.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(logStream),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Fail to get sequence token of CW log stream: %s %s", logGroup, logStream)
	}

	if resp.LogStreams == nil || len(resp.LogStreams) != 1 {
		return nil, fmt.Errorf("Unexpected number of LogStream in DescribeLogStreams: %v", resp.LogStreams)
	}

	return resp.LogStreams[0].UploadSequenceToken, nil
}

func createCloudWatchLogsStream(cwlogs *cloudwatchlogs.CloudWatchLogs, logGroup, logStream string) (*string, error) {
	createInput := cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
	}

	if _, err := cwlogs.CreateLogStream(&createInput); err != nil {
		return nil, errors.Wrapf(err, "Fail to create CW Logs LogStream: %s", logStream)
	}

	resp, err := cwlogs.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(logStream),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Fail to get sequence token of CW Logs stream: %s %s", logGroup, logStream)
	}

	if resp.LogStreams == nil || len(resp.LogStreams) != 1 {
		return nil, fmt.Errorf("Unexpected number of LogStream in DescribeLogStreams: %v", resp.LogStreams)
	}

	return resp.LogStreams[0].UploadSequenceToken, nil
}

func (x *cloudWatchLogsHook) Fire(entry *logrus.Entry) error {
	msg, err := json.Marshal(entry.Data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal loggerEntry Data")
	}

	event := cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(msg)),
		Timestamp: aws.Int64(entry.Time.UTC().Unix() * 1000),
	}

	input := cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(x.logGroup),
		LogStreamName: aws.String(x.logStream()),
		LogEvents:     []*cloudwatchlogs.InputLogEvent{&event},
	}

	if x.nextSequenceToken != nil {
		input.SequenceToken = x.nextSequenceToken
	} else {
		token, cwErr := createCloudWatchLogsStream(x.cwlogs, x.logGroup, x.logStream())

		if cwErr != nil {
			switch awsErr := cwErr.(type) {
			case awserr.Error:
				if awsErr.Code() != "ResourceAlreadyExistsException" {
					return err
				}

				token, err := getCloudWatchLogsNextToken(x.cwlogs, x.logGroup, x.logStream())
				if err != nil {
					return err
				}
				input.SequenceToken = token

			default:
				return errors.Wrap(err, "Fail go crete CW Logs stream")
			}
		} else {
			input.SequenceToken = token
		}
	}

	if _, err := x.cwlogs.PutLogEvents(&input); err != nil {
		return errors.Wrapf(err, "Fail to pu CW Logs to %s/%s", x.logGroup, x.logStream())
	}

	return nil
}
