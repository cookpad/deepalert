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

// Logger is a global logger for functions
var Logger *logrus.Entry
var loggerHook cloudWatchLogsHook
var loggerBase *logrus.Logger

func setupLogger() {
	loggerBase = logrus.New()
	loggerBase.SetLevel(logrus.InfoLevel)
	loggerBase.SetFormatter(&logrus.JSONFormatter{})

	loggerHook = cloudWatchLogsHook{
		logGroup:  os.Getenv("LOG_GROUP"),
		logStream: os.Getenv("LOG_STREAM"),
		region:    os.Getenv("AWS_REGION"),
	}
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(loggerHook.region),
	}))
	loggerHook.cwlogs = cloudwatchlogs.New(ssn)

	loggerBase.AddHook(&loggerHook)
	Logger = loggerBase.WithFields(logrus.Fields{})
}

// SetLoggerContext binds context and global logger.
func SetLoggerContext(ctx context.Context, funcName string, reportID deepalert.ReportID) {
	if ctx != nil {
		lc, _ := lambdacontext.FromContext(ctx)
		loggerHook.requestID = lc.AwsRequestID
	}
	loggerHook.funcName = funcName

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

func setLogDestination(logGroup, logStream, region string) {
	loggerHook.logGroup = logGroup
	loggerHook.logStream = logStream
	loggerHook.region = region

	// Redefine
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(loggerHook.region),
	}))
	loggerHook.cwlogs = cloudwatchlogs.New(ssn)
}

func setLogOutput(output io.Writer) {
	loggerBase.SetOutput(output)
}
func getLogOutput() io.Writer {
	return loggerBase.Out
}

type cloudWatchLogsHook struct {
	logGroup          string
	logStream         string
	region            string
	requestID         string
	reportID          deepalert.ReportID
	funcName          string
	cwlogs            *cloudwatchlogs.CloudWatchLogs
	nextSequenceToken *string
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
		LogStreamName: aws.String(x.logStream),
		LogEvents:     []*cloudwatchlogs.InputLogEvent{&event},
	}

	if x.nextSequenceToken != nil {
		input.SequenceToken = x.nextSequenceToken
	} else {
		input.SequenceToken, err = getCloudWatchLogsNextToken(x.cwlogs, x.logGroup, x.logStream)
		if err != nil {
			return err
		}
	}

	for n := 0; n < 10; n++ {
		if resp, err := x.cwlogs.PutLogEvents(&input); err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "InvalidSequenceTokenException" {
				fmt.Printf("[CW Logs Retry] %d\n", n)
				input.SequenceToken, err = getCloudWatchLogsNextToken(x.cwlogs, x.logGroup, x.logStream)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else if resp.NextSequenceToken != nil {
			x.nextSequenceToken = resp.NextSequenceToken
			break
		}
	}

	return nil
}
