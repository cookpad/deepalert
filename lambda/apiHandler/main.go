package main

import (
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/m-mizutani/golambda"

	"github.com/deepalert/deepalert/internal/api"
	"github.com/deepalert/deepalert/internal/handler"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args := handler.NewArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		return handleRequest(args, event)
	})
}

func handleRequest(args *handler.Arguments, event golambda.Event) (interface{}, error) {
	var req events.APIGatewayProxyRequest
	if err := event.Bind(&req); err != nil {
		return nil, err
	}

	logger.With("request", req).Info("HTTP request")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	api.SetupRoute(v1, args)

	return ginadapter.New(r).Proxy(req)
}
