package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Helper interface {
	EnsureBucket(bucket string, w io.Writer) error
	CreateBucket(bucket string, w io.Writer) error
	UploadDir(dir, bucket, prefix string, w io.Writer) error
	UploadObjectJSON(o interface{}, bucket, key string, w io.Writer) error
	DownloadObjectJSON(o interface{}, bucket, key string) error
}

type s3Helper struct {
	region   string
	s3Client s3iface.S3API
}

func NewS3Helper(s3Client s3iface.S3API, region string) S3Helper {
	return &s3Helper{
		s3Client: s3Client,
		region:   region,
	}
}

func (h *s3Helper) EnsureBucket(bucket string, w io.Writer) error {
	params := &s3.HeadBucketInput{
		Bucket: aws.String(bucket), // Required
	}

	_, err := h.s3Client.HeadBucket(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			return h.CreateBucket(bucket, w)
		}

		return err
	}

	return nil
}

func (h *s3Helper) CreateBucket(bucket string, w io.Writer) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucket), // Required
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(h.region),
		},
	}

	fmt.Fprintf(w, "Creating bucket '%s' in region '%s'\n", bucket, h.region)

	_, err := h.s3Client.CreateBucket(params)

	return err
}

func (h *s3Helper) UploadObjectJSON(o interface{}, bucket, key string, w io.Writer) error {
	uploader := s3manager.NewUploaderWithClient(h.s3Client)

	if err := h.EnsureBucket(bucket, w); err != nil {
		return err
	}

	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	params := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(b),
	}

	if _, err := uploader.Upload(params); err != nil {
		return err
	}

	return nil
}

func (h *s3Helper) DownloadObjectJSON(o interface{}, bucket, key string) error {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := h.s3Client.GetObject(params)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(o)
}

func (h *s3Helper) UploadDir(dir, bucket, prefix string, w io.Writer) error {
	uploader := s3manager.NewUploaderWithClient(h.s3Client)

	if err := h.EnsureBucket(bucket, w); err != nil {
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

		fmt.Fprintf(w, "Uploading resource '%s' to 's3://%s/%s'\n", file, bucket, prefix)

		if _, err := uploader.Upload(params); err != nil {
			return err
		}

		return nil
	})
}
