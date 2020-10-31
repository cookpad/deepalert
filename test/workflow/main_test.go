package workflow_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	stackName  = "DeepAlertTestStack"
	apiKeyFile = "apikey.json"
)

func TestWorkflow(t *testing.T) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		t.Skip("AWS_REGION is not set")
	}

	stacks, err := getStackResources(region, stackName)
	require.NoError(t, err)
	restAPIs := stacks.filterByResourceType("AWS::ApiGateway::RestApi")
	require.Equal(t, 1, len(restAPIs))

	restAPI := restAPIs[0]
	baseURL := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/prod/api/v1/", aws.StringValue(restAPI.PhysicalResourceId), region)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	apiKeyPath := path.Join(cwd, apiKeyFile)
	apiKey, err := loadAPIKey(apiKeyPath)
	require.NoError(t, err)

	header := http.Header{
		"X-API-KEY":    []string{apiKey["X-API-KEY"]},
		"content-type": []string{"application/json"},
	}
	client := &http.Client{}

	alert := deepalert.Alert{
		AlertKey: uuid.New().String(),
		RuleID:   "five",
		Detector: "blue",
	}
	alertData, err := json.Marshal(alert)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"alert", bytes.NewReader(alertData))
	require.NoError(t, err)
	req.Header = header
	resp, err := client.Do(req)
	require.NoError(t, err)
	var report deepalert.Report
	respRaw, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	// t.Log(string(respRaw))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, json.Unmarshal(respRaw, &report))
	assert.NotEmpty(t, report.ID)
}

var logger = logging.Logger

func init() {
	// Setup logger
	logger.SetLevel(logrus.InfoLevel)
}

type StackResources []*cloudformation.StackResource

func (x StackResources) filterByResourceType(resourceType string) StackResources {
	var result StackResources
	for _, resource := range x {
		if aws.StringValue(resource.ResourceType) == resourceType {
			result = append(result, resource)
		}
	}
	return result
}

func getStackResources(region, stackName string) (StackResources, error) {
	ssn := session.Must(session.NewSession())
	cfn := cloudformation.New(ssn, aws.NewConfig().WithRegion(region))

	req := cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	}
	resp, err := cfn.DescribeStackResources(&req)
	if err != nil {
		return nil, err
	}

	logger.WithField("resp", resp).Debug("result")

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
