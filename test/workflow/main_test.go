package workflow_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/logging"
	"github.com/deepalert/deepalert/test/workflow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apiKeyFile = "apikey.json"
)

var (
	logger        = logging.Logger
	mainStackName = "DeepAlertTestStack"
	testStackName = "DeepAlertTestWorkflowStack"
)

func init() {
	logging.SetLogLevel(os.Getenv("LOG_LEVEL"))
	if v, ok := os.LookupEnv("DEEPALERT_TEST_STACK_NAME"); ok {
		mainStackName = v
	}
	if v, ok := os.LookupEnv("DEEPALERT_WORKFLOW_STACK_NAME"); ok {
		testStackName = v
	}
}

func TestWorkflow(t *testing.T) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		t.Skip("AWS_REGION is not set")
	}
	client, err := newDAClient(region)
	require.NoError(t, err)

	repo, err := newRepository(region)
	require.NoError(t, err)

	testResults := func(t *testing.T, report *deepalert.Report) {
		require.NoError(t, expBackOff(10, func(_ uint) bool {
			respReport, err := client.Request("GET", fmt.Sprintf("report/%s", report.ID), nil)
			require.NoError(t, err)

			if respReport.StatusCode == 200 {
				var gotReport deepalert.Report
				require.NoError(t, unmarshal(respReport.Body, &gotReport))
				assert.Equal(t, report.ID, gotReport.ID)
				return true
			}

			return false
		}))

		require.NoError(t, expBackOff(10, func(_ uint) bool {
			results, err := repo.GetEmitterResult(report.ID)
			require.NoError(t, err)

			if len(results) == 2 {
				for _, res := range results {
					var tmp deepalert.Report
					require.NoError(t, json.Unmarshal([]byte(res.Data), &tmp))
					assert.Equal(t, report.ID, tmp.ID)
					t.Log(tmp)
					if tmp.Status == deepalert.StatusPublished {
						t.Log(tmp)
						return true
					}
				}
			}
			return false
		}))
	}

	t.Run("Sends an alert via API GW", func(t *testing.T) {
		alert := deepalert.Alert{
			AlertKey: uuid.New().String(),
			RuleID:   "five",
			Detector: "blue",
			Attributes: []deepalert.Attribute{
				{
					Type:  deepalert.TypeIPAddr,
					Value: "198.51.100.1",
				},
			},
		}

		resp, err := client.Request("POST", "alert", alert)
		require.NoError(t, err)

		var report deepalert.Report
		respRaw, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		// t.Log(string(respRaw))
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, json.Unmarshal(respRaw, &report))
		assert.NotEmpty(t, report.ID)

		testResults(t, &report)
	})

	t.Run("Sends an alert via SQS", func(t *testing.T) {
		ssn := session.Must(session.NewSession())
		sqsClient := sqs.New(ssn, aws.NewConfig().WithRegion(region))

		stacks, err := getStackResources(region, mainStackName)
		require.NoError(t, err)

		queues := stacks.
			filterByResourceType("AWS::SQS::Queue").filterByLogicalResourceId("alertQueue9836D344")
		require.Equal(t, 1, len(queues))

		alert := deepalert.Alert{
			AlertKey: uuid.New().String(),
			RuleID:   "six",
			Detector: "blue",
			Attributes: []deepalert.Attribute{
				{
					Type:  deepalert.TypeIPAddr,
					Value: "198.51.100.1",
				},
			},
		}
		raw, err := json.Marshal(alert)
		require.NoError(t, err)

		input := &sqs.SendMessageInput{
			QueueUrl:    aws.String(*queues[0].PhysicalResourceId),
			MessageBody: aws.String(string(raw)),
		}

		_, err = sqsClient.SendMessage(input)
		require.NoError(t, err)

		var report deepalert.Report
		require.NoError(t, expBackOff(10, func(_ uint) bool {

			resp, err := client.Request("GET", "alert/"+alert.AlertID()+"/report", nil)
			require.NoError(t, err)

			respRaw, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			if resp.StatusCode == http.StatusOK {
				t.Log("resp", string(respRaw))
				require.NoError(t, json.Unmarshal(respRaw, &report))
				assert.NotEmpty(t, report.ID)
				return true
			}
			return false
		}))

		testResults(t, &report)
	})
}

func unmarshal(reader io.Reader, data interface{}) error {
	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(raw, data); err != nil {
		return err
	}
	return nil
}

type stackResources []*cloudformation.StackResource

func (x stackResources) filterByResourceType(resourceType string) stackResources {
	var result stackResources
	for _, resource := range x {
		if aws.StringValue(resource.ResourceType) == resourceType {
			result = append(result, resource)
		}
	}
	return result
}

func (x stackResources) filterByLogicalResourceId(LogicalResourceId string) stackResources {
	var result stackResources
	for _, resource := range x {
		if strings.Contains(aws.StringValue(resource.LogicalResourceId), LogicalResourceId) {
			result = append(result, resource)
		}
	}
	return result
}

func getStackResources(region, stackName string) (stackResources, error) {
	ssn := session.Must(session.NewSession())
	cfn := cloudformation.New(ssn, aws.NewConfig().WithRegion(region))

	req := cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	}
	resp, err := cfn.DescribeStackResources(&req)
	if err != nil {
		return nil, err
	}

	return resp.StackResources, nil
}

type httpHeader map[string]string

func loadAPIKey(path string) (httpHeader, error) {
	hdr := make(httpHeader)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &hdr); err != nil {
		return nil, err
	}

	return hdr, nil
}

type daClient struct {
	BaseURL string
	header  map[string][]string
}

func newDAClient(region string) (*daClient, error) {
	stacks, err := getStackResources(region, mainStackName)
	if err != nil {
		return nil, err
	}

	restAPIs := stacks.filterByResourceType("AWS::ApiGateway::RestApi")
	if len(restAPIs) != 1 {
		return nil, fmt.Errorf("Invalid number of AWS::ApiGateway::RestApi")
	}

	restAPI := restAPIs[0]
	baseURL := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/prod/api/v1/", aws.StringValue(restAPI.PhysicalResourceId), region)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	apiKeyPath := path.Join(cwd, apiKeyFile)
	apiKey, err := loadAPIKey(apiKeyPath)
	if err != nil {
		return nil, err
	}

	client := &daClient{
		BaseURL: baseURL,
		header: http.Header{
			"X-API-KEY":    []string{apiKey["X-API-KEY"]},
			"content-type": []string{"application/json"},
		},
	}

	return client, nil
}

func (x *daClient) Request(method, path string, data interface{}) (*http.Response, error) {
	var reader io.Reader
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}

	url := x.BaseURL + path
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	for key, values := range x.header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return client.Do(req)
}

// Exponential backoff timer
var errRetryMaxExceeded = fmt.Errorf("RetryMax is exceeded")

func expBackOff(retryMax uint, callback func(count uint) bool) error {

	for i := uint(0); i < retryMax; i++ {
		logger.Tracef("Callback(%d, %p)", i, callback)
		if exit := callback(i); exit {
			return nil
		}

		if i+1 < retryMax {
			wait := math.Pow(float64(i)/3.14, 2)
			if wait > 10 {
				wait = 10
			}
			time.Sleep(time.Duration(wait * float64(time.Second)))
		}
	}

	return errRetryMaxExceeded
}

func newRepository(region string) (*workflow.Repository, error) {
	stacks, err := getStackResources(region, testStackName)
	if err != nil {
		return nil, err
	}

	tables := stacks.filterByResourceType("AWS::DynamoDB::Table")
	if len(tables) != 1 {
		return nil, fmt.Errorf("Invalid number of AWS::DynamoDB::Table")
	}

	logger.WithField("table", *tables[0]).Debug("Test table")
	return workflow.NewRepository(region, *tables[0].PhysicalResourceId)
}
