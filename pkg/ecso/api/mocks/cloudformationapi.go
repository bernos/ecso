package mocks

import "github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

type CloudFormationAPIMock struct {
	cloudformationiface.CloudFormationAPI
}
