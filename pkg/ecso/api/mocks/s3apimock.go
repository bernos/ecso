package mocks

import "github.com/aws/aws-sdk-go/service/s3/s3iface"

type S3APIMock struct {
	s3iface.S3API
}
