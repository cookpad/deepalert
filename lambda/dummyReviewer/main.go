package main

import (
	"github.com/m-mizutani/golambda"

	"github.com/deepalert/deepalert"
)

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return deepalert.ReportResult{
			Severity: deepalert.SevUnclassified,
			Reason:   "I'm novice",
		}, nil
	})
}
