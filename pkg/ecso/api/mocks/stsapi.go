package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type STSMock struct {
	stsiface.STSAPI

	getCallerIdentity func(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}

func (mock *STSMock) GetCallerIdentityReturns(output *sts.GetCallerIdentityOutput, err error) {
	mock.getCallerIdentity = func(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
		return output, err
	}
}

func (mock *STSMock) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if mock.getCallerIdentity != nil {
		return mock.getCallerIdentity(input)
	}
	return nil, fmt.Errorf("Not implemented")
}
