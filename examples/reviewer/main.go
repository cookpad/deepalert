package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
)

func main() {
	lambda.Start(evaluate)
}

// Example to evaluate security alert of suspicious activity on AWS
func evaluate(ctx context.Context, report deepalert.Report) (*deepalert.ReportResult, error) {
	for _, alert := range report.Alerts {
		// Skip if alert ruleID is not matched
		if alert.RuleID != "your_alert_rule_id" {
			return nil, nil
		}
	}

	// Extract results of Inspector
	for _, section := range report.Sections {
		for _, host := range section.Hosts {
			for _, owner := range host.Owner {
				// If source host is owned by your company
				if owner == "YOUR_COMPANY" {
					return &deepalert.ReportResult{
						// Evaluate the alert as safe (no action required)
						Severity: deepalert.SevSafe,
						Reason:   "The device accessing to G Suite is owned by YOUR_COMPANY.",
					}, nil
				}
			}
		}
	}

	return nil, nil
}
