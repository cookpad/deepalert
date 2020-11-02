package main

import (
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	"github.com/deepalert/deepalert/internal/api"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
)

var logger = logging.Logger

func main() {
	handler.StartLambda(handleRequest)
}

func handleRequest(args *handler.Arguments) (handler.Response, error) {
	var req events.APIGatewayProxyRequest
	if err := args.BindEvent(&req); err != nil {
		return nil, err
	}

	logger.WithField("request", req).Info("HTTP request")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	api.SetupRoute(v1, args)

	return ginadapter.New(r).Proxy(req)
}
