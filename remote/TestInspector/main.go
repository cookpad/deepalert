package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/inspector"
)

func dummyInspector(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	// tableName := os.Getenv("RESULT_TABLE")
	// reportID, _ := deepalert.ReportIDFromCtx(ctx)

	hostReport := deepalert.ReportHost{
		IPAddr: []string{"10.1.2.3"},
		Owner:  []string{"superman"},
	}

	newAttr := deepalert.Attribute{
		Key:   "username",
		Value: "mizutani",
		Type:  deepalert.TypeUserName,
	}

	return &deepalert.TaskResult{
		Contents:      []deepalert.ReportContent{&hostReport},
		NewAttributes: []deepalert.Attribute{newAttr},
	}, nil
}

func main() {
	inspector.Start(dummyInspector, "dummyInspector",
		os.Getenv("ATTRIBUTE_QUEUE"), os.Getenv("CONTENT_QUEUE"))
}
