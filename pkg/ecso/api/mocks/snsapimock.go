package mocks

import "github.com/aws/aws-sdk-go/service/sns/snsiface"

type SNSAPIMock struct {
	snsiface.SNSAPI
}
