package api

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api/mocks"
)

func NewEnvironmentAPIWithMockAWSServices() *environmentAPI {
	return &environmentAPI{
		cloudformationAPI: &mocks.CloudFormationAPIMock{},
		cloudwatchlogsAPI: &mocks.CloudWatchLogsAPIMock{},
		ecsAPI:            &mocks.ECSAPIMock{},
		route53API:        &mocks.Route53APIMock{},
		s3API:             &mocks.S3APIMock{},
		snsAPI:            &mocks.SNSAPIMock{},
		stsAPI:            &mocks.STSMock{},
	}
}

func TestGetCurrentAWSAccount(t *testing.T) {
	stsMock := &mocks.STSMock{}
	stsMock.GetCallerIdentityReturns(&sts.GetCallerIdentityOutput{
		Account: aws.String("abc123"),
	}, nil)

	api := NewEnvironmentAPIWithMockAWSServices()
	api.stsAPI = stsMock

	account, err := api.GetCurrentAWSAccount()

	if err != nil {
		t.Error(err)
	}

	if account != "abc123" {
		t.Errorf("Want empty string, got %s", account)
	}
}

func TestGetEcsoBucket(t *testing.T) {
	account := "abc123"
	region := "a-region"
	expect := fmt.Sprintf("ecso-%s-%s", region, account)
	env := &ecso.Environment{Region: region}
	stsMock := &mocks.STSMock{}
	stsMock.GetCallerIdentityReturns(&sts.GetCallerIdentityOutput{
		Account: aws.String(account),
	}, nil)

	api := NewEnvironmentAPIWithMockAWSServices()
	api.stsAPI = stsMock

	result, err := api.GetEcsoBucket(env)
	if err != nil {
		t.Error(err)
	}

	if result != expect {
		t.Errorf("Want '%s', got '%s'.", expect, result)
	}
}
