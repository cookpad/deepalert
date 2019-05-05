package main

import (
	"context"

	"github.com/m-mizutani/deepalert"
)

func dummyHandler(ctx context.Context, attr deepalert.Attribute) (deepalert.ReportContentEntity, error) {
	hostReport := deepalert.ReportHost{
		IPAddr: []string{"10.1.2.3"},
	}
	return &hostReport, nil
}

func main() {
	deepalert.StartInspector(dummyHandler, "dummyHandler")
}
