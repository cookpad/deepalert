package api_test

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/api"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handleRequest(args *handler.Arguments, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	route := gin.New()
	v1 := route.Group("/api/v1")
	api.SetupRoute(v1, args)
	return ginadapter.New(route).Proxy(req)
}

func toBody(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", v)
	}
	return string(raw)
}

func TestCreateReport(t *testing.T) {
	t.Run("Get report after creating", func(tt *testing.T) {
		_, repoFactory := mock.NewMockRepositorySet()
		args := &handler.Arguments{
			NewRepository: repoFactory,
			NewSFn:        mock.NewSFnClient,
			NewSNS:        mock.NewSNSClient,
			EnvVars: handler.EnvVars{
				InspectorMashine: "arn:aws:states:us-east-1:111122223333:stateMachine:inspect",
				ReviewMachine:    "arn:aws:states:us-east-1:111122223333:stateMachine:review",
				ReportTopic:      "arn:aws:sns:us-east-1:111122223333:report",
			},
		}
		alert := &deepalert.Alert{
			Detector: "testDetector",
			RuleID:   "r1",
			RuleName: "testRule",
		}

		postResp, err := handleRequest(args, events.APIGatewayProxyRequest{
			Path:       "/api/v1/alert",
			HTTPMethod: "POST",
			Body:       toBody(alert),
		})
		require.NoError(tt, err)
		assert.Equal(tt, 200, postResp.StatusCode)

		var report deepalert.Report
		require.NoError(tt, json.Unmarshal([]byte(postResp.Body), &report))
		assert.NotEqual(tt, deepalert.ReportID(""), report.ID)

		getReportResp, err := handleRequest(args, events.APIGatewayProxyRequest{
			Path:       fmt.Sprintf("/api/v1/report/%s", report.ID),
			HTTPMethod: "GET",
		})
		require.NoError(tt, err)
		var getReport deepalert.Report
		require.NoError(tt, json.Unmarshal([]byte(getReportResp.Body), &getReport))
		assert.Equal(tt, report.ID, getReport.ID)
		require.Equal(tt, 1, len(getReport.Alerts))
		assert.Equal(tt, alert, getReport.Alerts[0])

		getResp, err := handleRequest(args, events.APIGatewayProxyRequest{
			Path:       fmt.Sprintf("/api/v1/report/%s/alert", report.ID),
			HTTPMethod: "GET",
		})
		require.NoError(tt, err)
		assert.Equal(tt, 200, getResp.StatusCode)
		var alerts []*deepalert.Alert
		require.NoError(tt, json.Unmarshal([]byte(getResp.Body), &alerts))
		require.Equal(tt, 1, len(alerts))
		assert.Equal(tt, "testRule", alerts[0].RuleName)
	})
}

func TestErrorCase(t *testing.T) {
	t.Run("Invalid path", func(tt *testing.T) {
		_, repoFactory := mock.NewMockRepositorySet()
		args := &handler.Arguments{NewRepository: repoFactory}

		req := events.APIGatewayProxyRequest{
			Path:       "/api/v0/report",
			HTTPMethod: "POST",
		}

		resp, err := handleRequest(args, req)
		require.NoError(tt, err)
		assert.Equal(tt, 404, resp.StatusCode)
	})
}
