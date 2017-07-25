package config

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type Config struct {
	Version string

	// Base aws session. This session will be copied for each region
	sess *session.Session

	// Map of sessions by aws region
	sessions map[string]*session.Session

	w      io.Writer
	reader io.Reader
}

func (c *Config) getSession(region string) *session.Session {
	if _, ok := c.sessions[region]; !ok {
		c.sessions[region] = c.sess.Copy(&aws.Config{
			Region: aws.String(region),
		})
	}
	return c.sessions[region]
}

func (c *Config) ServiceAPI(region string) api.ServiceAPI {
	sess := c.getSession(region)

	return api.NewServiceAPI(
		cloudformation.New(sess),
		cloudwatchlogs.New(sess),
		ecs.New(sess),
		route53.New(sess),
		s3.New(sess),
		sns.New(sess),
		sts.New(sess))
}

func (c *Config) EnvironmentAPI(region string) api.EnvironmentAPI {
	sess := c.getSession(region)

	return api.NewEnvironmentAPI(
		cloudformation.New(sess),
		cloudwatchlogs.New(sess),
		ecs.New(sess),
		route53.New(sess),
		s3.New(sess),
		sns.New(sess),
		sts.New(sess))
}

func (c *Config) Writer() io.Writer {
	return c.w
}

func (c *Config) Reader() io.Reader {
	return c.reader
}

func (c *Config) ErrWriter() io.Writer {
	return ui.NewErrWriter(c.w)
}

func NewConfig(version string, options ...func(*Config)) (*Config, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	})

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Version:  version,
		w:        os.Stderr,
		reader:   os.Stdin,
		sess:     sess,
		sessions: make(map[string]*session.Session),
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}

func WithAWSSession(sess *session.Session) func(*Config) {
	return func(cfg *Config) {
		cfg.sess = sess
	}
}
