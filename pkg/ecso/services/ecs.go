package services

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type ECSService interface {
	LogServiceEvents(service, cluster string, logger func(*ecs.ServiceEvent, error)) (cancel func())
}

func NewECSService(ecsClient ecsiface.ECSAPI, log func(string, ...interface{})) ECSService {
	return &ecsService{
		ecsClient: ecsClient,
		log:       log,
	}
}

type ecsService struct {
	ecsClient ecsiface.ECSAPI
	log       func(string, ...interface{})
}

func (svc *ecsService) LogServiceEvents(service, cluster string, logger func(*ecs.ServiceEvent, error)) (cancel func()) {
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second * 5)

	params := &ecs.DescribeServicesInput{
		Cluster: aws.String(cluster),
		Services: []*string{
			aws.String(service),
		},
	}

	go func() {
		defer ticker.Stop()
		var lastEventID string

		for {
			resp, err := svc.ecsClient.DescribeServices(params)

			if err != nil {
				logger(nil, err)
			} else {
				if len(resp.Services) != 1 {
					logger(nil, fmt.Errorf("Expected to find 1 service, but found %d", len(resp.Services)))
				} else {
					newEvents := resp.Services[0].Events[:1]

					if lastEventID != "" {
						newEvents = resp.Services[0].Events

						for i, event := range resp.Services[0].Events {
							if *event.Id == lastEventID {
								newEvents = resp.Services[0].Events[:i]
								break
							}
						}
					}

					for i := len(newEvents) - 1; i >= 0; i-- {
						logger(newEvents[i], nil)
						lastEventID = *newEvents[i].Id
					}
				}
			}

			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()

	return func() {
		close(done)
	}

}
