package handler

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/deepalert/internal/adaptor"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/m-mizutani/deepalert/internal/repository"
	"github.com/m-mizutani/deepalert/internal/service"
)

// Arguments has environment variables, Event record and adaptor
type Arguments struct {
	EnvVars
	Event interface{}

	NewSQS        adaptor.SNSClientFactory  `json:"-"`
	NewSFn        adaptor.SFnClientFactory  `json:"-"`
	NewRepository adaptor.RepositoryFactory `json:"-"`
}

func newArguments() *Arguments {
	return &Arguments{}
}

// EventRecord is decapslated event data (e.g. Body of SQS event)
type EventRecord []byte

// Bind unmarshal event record to object
func (x EventRecord) Bind(ev interface{}) error {
	if err := json.Unmarshal(x, ev); err != nil {
		return errors.Wrap(err, "Failed json.Unmarshal in DecodeEvent").With("raw", string(x))
	}
	return nil
}

// DecapSQSEvent decapslates wrapped body data in SQSEvent
func (x *Arguments) DecapSQSEvent() ([]EventRecord, error) {
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

// DecapSNSEvent decapslates wrapped body data in SNSEvent
func (x *Arguments) DecapSNSEvent() ([]EventRecord, error) {
	var snsEvent events.SNSEvent
	if err := x.BindEvent(&snsEvent); err != nil {
		return nil, err
	}

	var output []EventRecord
	for _, record := range snsEvent.Records {
		output = append(output, EventRecord(record.SNS.Message))
	}

	return output, nil
}

// BindEvent directly decode event data and unmarshal to ev object.
func (x *Arguments) BindEvent(ev interface{}) error {
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

// SFnService provides service.SFnService with SQS adaptor
func (x *Arguments) SFnService() *service.SFnService {
	if x.NewSFn != nil {
		return service.NewSFnService(x.NewSFn)
	}
	return service.NewSFnService(adaptor.NewSFnClient)
}

// Repository provides data store accessor created by NewDynamoDB. If Arguments.NewRepository is set, this function returns repository object created by NewRepository.
func (x *Arguments) Repository() *service.RepositoryService {
	var ttl int64 = 1800
	if x.NewRepository != nil {
		return service.NewRepositoryService(x.NewRepository(x.CacheTable, x.AwsRegion), ttl)
	}
	return service.NewRepositoryService(repository.NewDynamoDB(x.CacheTable, x.AwsRegion), ttl)
}
