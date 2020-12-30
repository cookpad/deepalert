package api

import (
	"fmt"
	"net/http"

	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/sirupsen/logrus"
)

const (
	contextArgumentKey = "handler.arguments"
	contextRequestID   = "request.id"
	paramReportID      = "report_id"
	paramAlertID       = "alert_id"
)

var logger = golambda.Logger

func getArguments(c *gin.Context) *handler.Arguments {
	// In API code, handler.Arguments must be retrieved. If failed, the process must fail
	ptr, ok := c.Get(contextArgumentKey)
	if !ok {
		logger.With("key", contextArgumentKey).Error("Config is not set in API")
		panic("Config is not set in API")
	}

	args, ok := ptr.(*handler.Arguments)
	if !ok {
		logger.With("key", contextArgumentKey).Error("Config data can not be casted")
		panic("Config data can not be casted")
	}

	return args
}

func getRequestID(c *gin.Context) string {
	// In API code, requestID must be retrieved. If failed, the process must fail
	ptr, ok := c.Get(contextRequestID)
	if !ok {
		logger.With("contextRequestID", contextRequestID).Error("RequestID is not set in API")
		panic("RequestID is not set in API")
	}

	reqID, ok := ptr.(string)
	if !ok {
		logger.With("contextRequestID", contextRequestID).Error("RequestID can not be casted")
		panic("RequestID can not be casted")
	}

	return reqID
}

func wrapErr(msg string) map[string]string {
	return map[string]string{
		"error": msg,
	}
}

func resp(c *gin.Context, data interface{}) {
	reqID := getRequestID(c)
	c.Header("DeepAlert-Request-ID", reqID)

	if err, ok := data.(error); ok {
		if e, ok := err.(*errors.Error); ok {
			fields := logrus.Fields{
				"trace": fmt.Sprintf("%+v", e),
			}
			for k, v := range e.Values {
				fields[k] = v
			}
			logger.With("fields", fields).With("err", err).Error("Request Error")

			if 400 <= e.StatusCode && e.StatusCode < 500 {
				c.JSON(e.StatusCode, wrapErr(e.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, wrapErr("Internal Server Error"))
			}
		} else {
			logger.With("err", err).Error("Request Error (not errors.Error")

			c.JSON(http.StatusInternalServerError, wrapErr("SystemError"))
		}
	} else {
		c.JSON(http.StatusOK, data)
	}
}

// SetupRoute binds route of gin and API
func SetupRoute(r *gin.RouterGroup, args *handler.Arguments) {
	r.Use(func(c *gin.Context) {
		reqID := uuid.New().String()
		logger.
			With("path", c.FullPath()).
			With("params", c.Params).
			With("request_id", reqID).
			With("remote", c.ClientIP()).
			With("ua", c.Request.UserAgent()).
			Info("API request")

		c.Set(contextRequestID, reqID)
		c.Set(contextArgumentKey, args)
		c.Next()
	})

	r.POST("/alert", postAlert)
	r.GET("/alert/:"+paramAlertID+"/report", getReportByAlertID)
	r.GET("/report/:"+paramReportID, getReport)
	r.GET("/report/:"+paramReportID+"/alert", getReportAlerts)
	r.GET("/report/:"+paramReportID+"/section", getSections)
	r.GET("/report/:"+paramReportID+"/attribute", getReportAttributes)
}
