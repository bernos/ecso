package helpers

import "fmt"

// Package is our unit of deployment. It refers to a set of cloudformation templates and
// related resources which have been uploaded to S3, and which can be deployed using the
// CloudFormationHelper.Deploy method
type Package struct {
	bucket string
	prefix string
	region string
}

func NewPackage(bucket, prefix, region string) *Package {
	return &Package{bucket, prefix, region}
}

func (p *Package) GetBucketPrefix() string {
	return p.prefix
}

func (p *Package) GetTemplateBucketPrefix() string {
	return fmt.Sprintf("%s/templates", p.GetBucketPrefix())
}

func (p *Package) GetTemplateBucketKey() string {
	return fmt.Sprintf("%s/stack.yaml", p.GetTemplateBucketPrefix())
}

func (p *Package) GetTagsBucketKey() string {
	return fmt.Sprintf("%s/tags.json", p.GetBucketPrefix())
}

func (p *Package) GetParamsBucketKey() string {
	return fmt.Sprintf("%s/params.json", p.GetBucketPrefix())
}

func (p *Package) GetTemplateURL() string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", p.region, p.bucket, p.GetTemplateBucketKey())
}

func (p *Package) GetURL() string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", p.region, p.bucket, p.GetBucketPrefix())
}
