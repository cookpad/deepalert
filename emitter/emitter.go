package emitter

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/m-mizutani/golambda"
)

// SNSEventToReport extracts set of deepalert.Report from events.SNSEvent
func SNSEventToReport(event events.SNSEvent) ([]*deepalert.Report, error) {
	var reports []*deepalert.Report
	for _, record := range event.Records {
		var report deepalert.Report
		msg := record.SNS.Message
		if err := json.Unmarshal([]byte(msg), &report); err != nil {
			return nil, golambda.WrapError(err, "Fail to unmarshal report").With("msg", msg)
		}

		reports = append(reports, &report)
	}

	return reports, nil
}
