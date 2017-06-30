package helpers

import (
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bernos/ecso/pkg/ecso/log"
)

type S3Helper interface {
	EnsureBucket(bucket string) error
	CreateBucket(bucket string) error
	UploadDir(dir, bucket, prefix string) error
}

type s3Helper struct {
	region   string
	s3Client s3iface.S3API
	logger   log.Logger
}

func NewS3Helper(s3Client s3iface.S3API, region string, logger log.Logger) S3Helper {
	return &s3Helper{
		s3Client: s3Client,
		region:   region,
		logger:   logger,
	}
}

func (h *s3Helper) EnsureBucket(bucket string) error {
	params := &s3.HeadBucketInput{
		Bucket: aws.String(bucket), // Required
	}

	_, err := h.s3Client.HeadBucket(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			return h.CreateBucket(bucket)
		}

		return err
	}

	return nil
}

func (h *s3Helper) CreateBucket(bucket string) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucket), // Required
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(h.region),
		},
	}

	h.logger.Printf("Creating bucket '%s' in region '%s'\n", bucket, h.region)

	_, err := h.s3Client.CreateBucket(params)

	return err
}

func (h *s3Helper) UploadDir(dir, bucket, prefix string) error {
	uploader := s3manager.NewUploaderWithClient(h.s3Client)

	if err := h.EnsureBucket(bucket); err != nil {
		return err
	}

	return filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		reader, err := os.Open(file)
		if err != nil {
			return err
		}

		defer reader.Close()

		key := path.Join(prefix, string(file[len(dir):]))

		params := &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   reader,
		}

		h.logger.Printf("Uploading resource '%s' to 's3://%s/%s'\n", file, bucket, prefix)

		if _, err := uploader.Upload(params); err != nil {
			return err
		}

		return nil
	})
}
