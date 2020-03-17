package main

import (
	"context"
	"log"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/emitter"
)

func handler(ctx context.Context, report deepalert.Report) error {
	log.Println(report.Result.Severity)
	// Or do appropriate action according to report content and severity

	return nil
}

func main() {
	emitter.Start(handler)
}
