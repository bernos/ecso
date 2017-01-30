package ecso

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso/services"
)

type Config struct {
	Logger Logger
	// STS                   stsiface.STSAPI
	// CloudFormationService services.CloudFormationService

	// aws services by aws region
	sessions    map[string]*session.Session
	stsClients  map[string]stsiface.STSAPI
	cfnClients  map[string]cloudformationiface.CloudFormationAPI
	s3Clients   map[string]s3iface.S3API
	cfnServices map[string]services.CloudFormationService
	ecsClients  map[string]ecsiface.ECSAPI
	ecsServices map[string]services.ECSService
}

func (c *Config) STSAPI(region string) (stsiface.STSAPI, error) {
	if c.stsClients == nil {
		c.stsClients = make(map[string]stsiface.STSAPI)
	}

	if _, ok := c.stsClients[region]; !ok {
		sess, err := c.getSession(region)

		if err != nil {
			return nil, err
		}

		c.stsClients[region] = sts.New(sess)
	}

	return c.stsClients[region], nil
}

func (c *Config) ECSAPI(region string) (ecsiface.ECSAPI, error) {
	if c.ecsClients == nil {
		c.ecsClients = make(map[string]ecsiface.ECSAPI)
	}

	if _, ok := c.ecsClients[region]; !ok {
		sess, err := c.getSession(region)

		if err != nil {
			return nil, err
		}

		c.ecsClients[region] = ecs.New(sess)
	}

	return c.ecsClients[region], nil
}

func (c *Config) CloudFormationAPI(region string) (cloudformationiface.CloudFormationAPI, error) {
	if c.cfnClients == nil {
		c.cfnClients = make(map[string]cloudformationiface.CloudFormationAPI)
	}

	if _, ok := c.cfnClients[region]; !ok {
		sess, err := c.getSession(region)

		if err != nil {
			return nil, err
		}

		c.cfnClients[region] = cloudformation.New(sess)
	}

	return c.cfnClients[region], nil
}

func (c *Config) S3API(region string) (s3iface.S3API, error) {
	if c.s3Clients == nil {
		c.s3Clients = make(map[string]s3iface.S3API)
	}

	if _, ok := c.s3Clients[region]; !ok {
		sess, err := c.getSession(region)

		if err != nil {
			return nil, err
		}

		c.s3Clients[region] = s3.New(sess)
	}

	return c.s3Clients[region], nil
}

func (c *Config) CloudFormationService(region string) (services.CloudFormationService, error) {
	if c.cfnServices == nil {
		c.cfnServices = make(map[string]services.CloudFormationService)
	}

	if _, ok := c.cfnServices[region]; !ok {
		s3API, err := c.S3API(region)

		if err != nil {
			return nil, err
		}

		cfnAPI, err := c.CloudFormationAPI(region)

		if err != nil {
			return nil, err
		}

		c.cfnServices[region] = services.NewCloudFormationService(region, cfnAPI, s3API, c.Logger.PrefixPrintf("  "))
	}

	return c.cfnServices[region], nil
}

func (c *Config) ECSService(region string) (services.ECSService, error) {
	if c.ecsServices == nil {
		c.ecsServices = make(map[string]services.ECSService)
	}

	if _, ok := c.ecsServices[region]; !ok {
		ecsAPI, err := c.ECSAPI(region)

		if err != nil {
			return nil, err
		}

		c.ecsServices[region] = services.NewECSService(ecsAPI, c.Logger.PrefixPrintf("  "))
	}

	return c.ecsServices[region], nil
}

func (c *Config) getSession(region string) (*session.Session, error) {
	if c.sessions == nil {
		c.sessions = make(map[string]*session.Session)
	}

	if _, ok := c.sessions[region]; !ok {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})

		if err != nil {
			return sess, err
		}

		c.sessions[region] = sess
	}

	return c.sessions[region], nil
}

func NewConfig(options ...func(*Config)) (*Config, error) {
	log := NewLogger(os.Stdout)

	cfg := &Config{
		Logger: log,
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}
