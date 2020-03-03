package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/test"
	"github.com/pkg/errors"
)

func dummyInspector(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	tableName := os.Getenv("RESULT_TABLE")
	repo := test.NewRepository(os.Getenv("AWS_REGION"), tableName)
	reportID, _ := deepalert.ReportIDFromCtx(ctx)

	if err := repo.PutInspectorResult(*reportID, attr.Key, attr.Value); err != nil {
		return nil, errors.Wrapf(err, "Fail to put result to %v", tableName)
	}

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
	deepalert.StartInspector(dummyInspector, "dummyInspector",
		os.Getenv("ATTRIBUTE_QUEUE"), os.Getenv("CONTENT_QUEUE"))
}
