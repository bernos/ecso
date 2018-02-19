package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type CloudFormationAPIMock struct {
	cloudformationiface.CloudFormationAPI

	describeStacks func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
}

func (mock *CloudFormationAPIMock) DescribeStacksReturns(output *cloudformation.DescribeStacksOutput, err error) {
	mock.describeStacks = func(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
		return output, err
	}
}

func (mock *CloudFormationAPIMock) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if mock.describeStacks != nil {
		return mock.describeStacks(input)
	}
	return nil, fmt.Errorf("Not implemented")
}
