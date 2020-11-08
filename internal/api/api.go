package api

import (
	"fmt"
	"net/http"

	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	contextArgumentKey = "handler.arguments"
	contextRequestID   = "request.id"
	paramReportID      = "report_id"
	paramAlertID       = "alert_id"
)

var logger = logging.Logger

func getArguments(c *gin.Context) *handler.Arguments {
	// In API code, handler.Arguments must be retrieved. If failed, the process must fail
	ptr, ok := c.Get(contextArgumentKey)
	if !ok {
		logger.Fatalf("Config is not set in API as '%s'", contextArgumentKey)
		return nil
	}

	args, ok := ptr.(*handler.Arguments)
	if !ok {
		logger.Fatalf("Config data as '%s' can not be casted", contextArgumentKey)
		return nil
	}

	return args
}

func getRequestID(c *gin.Context) string {
	// In API code, requestID must be retrieved. If failed, the process must fail
	ptr, ok := c.Get(contextRequestID)
	if !ok {
		logger.Fatalf("RequestID is not set in API as '%s'", contextRequestID)
	}

	reqID, ok := ptr.(string)
	if !ok {
		logger.Fatalf("RequestID as '%s' can not be casted", contextRequestID)
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
			logger.WithFields(fields).WithError(err).Error("Request Error")

			if 400 <= e.StatusCode && e.StatusCode < 500 {
				c.JSON(e.StatusCode, wrapErr(e.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, wrapErr("Internal Server Error"))
			}
		} else {
			logger.WithError(err).Error("Request Error (not errors.Error")

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
		logger.WithFields(logrus.Fields{
			"path":       c.FullPath(),
			"params":     c.Params,
			"request_id": reqID,
			"remote":     c.ClientIP(),
			"ua":         c.Request.UserAgent(),
		}).Info("API request")

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
