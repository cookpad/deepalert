package api

import (
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/usecase"
	"github.com/gin-gonic/gin"
)

func postAlert(c *gin.Context) {
	args := getArguments(c)
	now := time.Now().UTC()

	var alert deepalert.Alert
	if err := c.BindJSON(&alert); err != nil {
		resp(c, wrapUserError(err, "Failed to pase deepalert.Alert"))
		return
	}

	report, err := usecase.HandleAlert(args, alert, now)
	if err != nil {
		resp(c, err)
		return
	}

	resp(c, report)
	return
}
