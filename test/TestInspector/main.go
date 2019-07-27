package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
)

func dummyInspector(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
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
		os.Getenv("SUBMIT_TOPIC"), os.Getenv("ATTRIBUTE_TOPIC"))
}
