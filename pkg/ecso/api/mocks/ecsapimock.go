package mocks

import "github.com/aws/aws-sdk-go/service/ecs/ecsiface"

type ECSAPIMock struct {
	ecsiface.ECSAPI
}
