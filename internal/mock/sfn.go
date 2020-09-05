package mock

import (
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/m-mizutani/deepalert/internal/adaptor"
)

// NewSFnClient creates mock SNS client
func NewSFnClient(region string) adaptor.SFnClient {
	return &SFnClient{region: region}
}

// SFnClient is mock
type SFnClient struct {
	region string
	input  []*sfn.StartExecutionInput
}

// StartExecution of mock SFnClient only stores sfn.StartExecutionInput
func (x *SFnClient) StartExecution(*sfn.StartExecutionInput) (*sfn.StartExecutionOutput, error) {
	return &sfn.StartExecutionOutput{}, nil
}
