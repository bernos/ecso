package helpers

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type ECSHelper interface {
	LogServiceEvents(service, cluster string, logger func(*ecs.ServiceEvent, error)) (cancel func())
}

func NewECSHelper(ecsClient ecsiface.ECSAPI, w io.Writer) ECSHelper {
	return &ecsHelper{
		w:         w,
		ecsClient: ecsClient,
	}
}

type ecsHelper struct {
	w         io.Writer
	ecsClient ecsiface.ECSAPI
}

func (h *ecsHelper) LogServiceEvents(service, cluster string, logger func(*ecs.ServiceEvent, error)) (cancel func()) {
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
			resp, err := h.ecsClient.DescribeServices(params)

			if err != nil {
				logger(nil, err)
			} else {
				if len(resp.Services) != 1 {
					logger(nil, fmt.Errorf("Expected to find 1 service, but found %d", len(resp.Services)))
				} else {
					if len(resp.Services[0].Events) > 0 {
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
