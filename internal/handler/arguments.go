package handler

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/deepalert/internal/adaptor"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/m-mizutani/deepalert/internal/service"
)

// Arguments has environment variables, Event record and adaptor
type Arguments struct {
	EnvVars
	Event interface{}

	NewSQS adaptor.SNSClientFactory `json:"-"`
}

// EventRecord is decapslated event data (e.g. Body of SQS event)
type EventRecord []byte

// Bind unmarshal event record to object
func (x EventRecord) Bind(ev interface{}) *errors.Error {
	if err := json.Unmarshal(x, ev); err != nil {
		return errors.Wrap(err, "Failed json.Unmarshal in DecodeEvent").With("raw", string(x))
	}
	return nil
}

// DecapSQSEvent decapslates wrapped body data in SQSEvent
func (x *Arguments) DecapSQSEvent() ([]EventRecord, *errors.Error) {
	var sqsEvent events.SQSEvent
	if err := x.BindEvent(&sqsEvent); err != nil {
		return nil, err
	}

	var output []EventRecord
	for _, record := range sqsEvent.Records {
		output = append(output, EventRecord(record.Body))
	}

	return output, nil
}

// BindEvent directly decode event data and unmarshal to ev object.
func (x *Arguments) BindEvent(ev interface{}) *errors.Error {
	raw, err := json.Marshal(x.Event)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal lambda event in BindEvent").With("event", x.Event)
	}

	if err := json.Unmarshal(raw, ev); err != nil {
		return errors.Wrap(err, "Failed json.Unmarshal in BindEvent").With("raw", string(raw))
	}

	return nil
}

// SNSService provides service.SNSService with SQS adaptor
func (x *Arguments) SNSService() *service.SNSService {
	if x.NewSQS != nil {
		return service.NewSNSService(x.NewSQS)
	}
	return service.NewSNSService(adaptor.NewSNSClient)
}
