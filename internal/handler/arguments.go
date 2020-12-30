package handler

import (
	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/repository"
	"github.com/deepalert/deepalert/internal/service"
)

// Arguments has environment variables, Event record and adaptor
type Arguments struct {
	EnvVars

	NewSNS        adaptor.SNSClientFactory  `json:"-"`
	NewSFn        adaptor.SFnClientFactory  `json:"-"`
	NewRepository adaptor.RepositoryFactory `json:"-"`
}

// NewArguments is constructor of Arguments
func NewArguments() *Arguments {
	return &Arguments{}
}

// SNSService provides service.SNSService with SQS adaptor
func (x *Arguments) SNSService() *service.SNSService {
	if x.NewSNS != nil {
		return service.NewSNSService(x.NewSNS)
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
func (x *Arguments) Repository() (*service.RepositoryService, error) {
	var ttl int64 = 1800
	var repo adaptor.Repository

	if x.NewRepository != nil {
		repo = x.NewRepository(x.AwsRegion, x.CacheTable)
	} else {
		dynamodb, err := repository.NewDynamoDB(x.AwsRegion, x.CacheTable)
		if err != nil {
			return nil, err
		}
		repo = dynamodb
	}

	return service.NewRepositoryService(repo, ttl), nil
}
