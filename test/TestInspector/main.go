package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
)

func dummyInspector(ctx context.Context, attr deepalert.Attribute) (deepalert.ReportContentEntity, error) {
	hostReport := deepalert.ReportHost{
		IPAddr: []string{"10.1.2.3"},
	}
	return &hostReport, nil
}

func main() {
	deepalert.StartInspector(dummyInspector, "dummyInspector", os.Getenv("SUBMIT_TOPIC"))
}
