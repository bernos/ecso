package helpers

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

var (
	childTemplateRegexp = regexp.MustCompile(`(\s*)TemplateURL:\s*(\./)*(.+)`)
)

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

// DeploymentResult holds information about a successful cloud formation
// template deployment
type DeploymentResult struct {
	StackID            string
	ChangeSetID        string
	DidRequireUpdating bool
}

// CloudFormationHelper contains high level helper functions for dealing with
// cloud formation
type CloudFormationHelper interface {
	DeleteStack(serviceName string) error
	Deploy(pkg *Package, stackName string, dryRun bool) (*DeploymentResult, error)
	GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error)
	GetStackOutputs(stackName string) (map[string]string, error)
	Package(templateFile, bucket, prefix string, tags, params map[string]string) (*Package, error)
	StackExists(stackName string) (bool, error)
	WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error)
	PackageIsUploadedToS3(pkg *Package) (bool, error)
}

// NewCloudFormationHelper creates a CloudFormationHelper
func NewCloudFormationHelper(region string, cfnClient cloudformationiface.CloudFormationAPI, s3Client s3iface.S3API, stsClient stsiface.STSAPI, w io.Writer) CloudFormationHelper {
	return &cfnHelper{
		w:         w,
		region:    region,
		cfnClient: cfnClient,
		s3Client:  s3Client,
		stsClient: stsClient,
		uploader:  s3manager.NewUploaderWithClient(s3Client),
	}
}

type cfnHelper struct {
	w         io.Writer
	region    string
	cfnClient cloudformationiface.CloudFormationAPI
	uploader  *s3manager.Uploader
	s3Client  s3iface.S3API
	stsClient stsiface.STSAPI
}

func (h *cfnHelper) GetStackOutputs(stackName string) (map[string]string, error) {
	resp, err := h.cfnClient.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return nil, err
	}

	outputs := make(map[string]string)

	for _, stack := range resp.Stacks {
		if *stack.StackName == stackName {
			for _, output := range stack.Outputs {
				outputs[*output.OutputKey] = *output.OutputValue
			}
		}
	}

	return outputs, nil
}

// Package creates a Package from local cloudformation template file. Any child templates in the
// template file will be uploaded to S3, as well as the template file itself. Before the template
// is uploaded, and relative references to child templates will be updated with the fully qualified
// S3 url that they were uploaded to. The resulting Package can be deployed using the Deploy method
func (h *cfnHelper) Package(templateFile, bucket, prefix string, tags, params map[string]string) (*Package, error) {
	pkg := NewPackage(bucket, prefix, h.region)

	fmt.Fprintf(h.w, "Creating deployment package at %s\n", pkg.GetURL())

	if err := h.validateTemplateFile(templateFile); err != nil {
		return nil, err
	}

	templatePrefix := pkg.GetTemplateBucketPrefix()
	basedir := filepath.Dir(templateFile)

	templateBody, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return pkg, err
	}

	if err := h.uploadChildTemplates(basedir, string(templateBody), bucket, templatePrefix); err != nil {
		return pkg, err
	}

	body := updateNestedTemplateURLs(string(templateBody), h.region, bucket, templatePrefix)

	if err := h.validateTemplate([]byte(body)); err != nil {
		return pkg, err
	}

	if err := h.uploadTemplate(strings.NewReader(body), bucket, pkg.GetTemplateBucketKey()); err != nil {
		return pkg, err
	}

	s3Helper := NewS3Helper(h.s3Client, h.region, ui.NewPrefixWriter(h.w, "  "))

	fmt.Fprintf(h.w, "Uploading cloud formation tags to %s\n", pkg.GetTagsBucketKey())
	if err := s3Helper.UploadObjectJSON(tags, bucket, pkg.GetTagsBucketKey()); err != nil {
		return nil, err
	}

	fmt.Fprintf(h.w, "Uploading cloud formation params to %s\n", pkg.GetParamsBucketKey())
	if err := s3Helper.UploadObjectJSON(params, bucket, pkg.GetParamsBucketKey()); err != nil {
		return nil, err
	}

	return pkg, nil
}

func (h *cfnHelper) PackageIsUploadedToS3(pkg *Package) (bool, error) {
	if _, err := h.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(pkg.bucket),
		Key:    aws.String(pkg.GetTemplateBucketKey()),
	}); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// process SDK error
			if awsErr.Code() == "NotFound" {
				return false, nil
			}
		}

		return false, err
	}

	return true, nil
}

func (h *cfnHelper) Deploy(pkg *Package, stackName string, dryRun bool) (*DeploymentResult, error) {
	fmt.Fprintf(h.w, "Deploying package from %s\n", pkg.GetURL())

	s3Helper := NewS3Helper(h.s3Client, h.region, ui.NewPrefixWriter(h.w, "  "))

	versionExists, err := h.PackageIsUploadedToS3(pkg)
	if err != nil {
		return nil, err
	}

	if !versionExists {
		return nil, fmt.Errorf("Deployment package not found at %s", pkg.GetBucketPrefix())
	}

	// Download params and tags
	params := make(map[string]string)
	tags := make(map[string]string)

	fmt.Fprintf(h.w, "Downloading stack params \n")
	if err := s3Helper.DownloadObjectJSON(&params, pkg.bucket, pkg.GetParamsBucketKey()); err != nil {
		return nil, err
	}

	fmt.Fprintf(h.w, "Downloading stack tags \n")
	if err := s3Helper.DownloadObjectJSON(&tags, pkg.bucket, pkg.GetTagsBucketKey()); err != nil {
		return nil, err
	}

	input := &cloudformation.CreateChangeSetInput{
		StackName:           aws.String(stackName),
		ChangeSetName:       aws.String(fmt.Sprintf("%s-%d", stackName, time.Now().Unix())),
		Parameters:          make([]*cloudformation.Parameter, 0),
		Tags:                make([]*cloudformation.Tag, 0),
		TemplateURL:         aws.String(pkg.GetTemplateURL()),
		UsePreviousTemplate: aws.Bool(false),
		Capabilities: []*string{
			aws.String("CAPABILITY_NAMED_IAM"),
			aws.String("CAPABILITY_IAM"),
		},
	}

	for k, v := range params {
		input.Parameters = append(input.Parameters, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}

	for k, v := range tags {
		input.Tags = append(input.Tags, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	exists, err := h.StackExists(stackName)

	if err != nil {
		return nil, err
	}

	if exists {
		input.ChangeSetType = aws.String("UPDATE")
		fmt.Fprintf(h.w, "Updating existing '%s' cloudformation stack\n", stackName)
	} else {
		fmt.Fprintf(h.w, "Creating new '%s' cloudformation stack\n", stackName)
		input.ChangeSetType = aws.String("CREATE")
	}

	fmt.Fprintf(h.w, "Creating changeset...\n")

	changeset, err := h.cfnClient.CreateChangeSet(input)

	if err != nil {
		return nil, err
	}

	result := &DeploymentResult{
		StackID:            *changeset.StackId,
		ChangeSetID:        *changeset.Id,
		DidRequireUpdating: true,
	}

	fmt.Fprintf(h.w, "Waiting for changeset %s to be ready...\n", *changeset.Id)

	if changeSetDescription, err := h.WaitForChangeset(*changeset.Id, cloudformation.ChangeSetStatusCreateComplete, cloudformation.ChangeSetStatusFailed); err != nil {
		return result, err
	} else if len(changeSetDescription.Changes) == 0 {
		result.DidRequireUpdating = false
		return result, nil
	}

	fmt.Fprintf(h.w, "Created changeset %s\n", *changeset.Id)

	if dryRun {
		return result, nil
	}

	if _, err := h.cfnClient.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName: changeset.Id,
		StackName:     changeset.StackId,
	}); err != nil {
		return result, err
	}

	stack := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	childWriter := ui.NewPrefixWriter(h.w, "  ")

	cancel := h.LogStackEvents(*changeset.StackId, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			fmt.Fprintf(childWriter, "%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	if exists {
		fmt.Fprintf(h.w, "Waiting for stack update to complete...\n")
		return result, h.cfnClient.WaitUntilStackUpdateComplete(stack)
	}

	fmt.Fprintf(h.w, "Waiting for stack creation to complete...\n")
	return result, h.cfnClient.WaitUntilStackCreateComplete(stack)
}

func (h *cfnHelper) GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error) {
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeset),
	}

	return h.cfnClient.DescribeChangeSet(params)
}

func (h *cfnHelper) DeleteStack(stackName string) error {
	_, err := h.cfnClient.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return err
	}

	childWriter := ui.NewPrefixWriter(h.w, "  ")

	cancel := h.LogStackEvents(stackName, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			fmt.Fprintf(childWriter, "%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	fmt.Fprintf(h.w, "Waiting for stack delete to complete...\n")

	return h.cfnClient.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
}

func (h *cfnHelper) StackExists(stackName string) (bool, error) {
	params := &cloudformation.ListStacksInput{
		StackStatusFilter: []*string{
			aws.String(cloudformation.StackStatusCreateComplete),
			aws.String(cloudformation.StackStatusCreateFailed),
			aws.String(cloudformation.StackStatusCreateInProgress),
			aws.String(cloudformation.StackStatusRollbackComplete),
			aws.String(cloudformation.StackStatusRollbackFailed),
			aws.String(cloudformation.StackStatusRollbackInProgress),
			aws.String(cloudformation.StackStatusUpdateComplete),
			aws.String(cloudformation.StackStatusUpdateCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusUpdateInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackComplete),
			aws.String(cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackFailed),
			aws.String(cloudformation.StackStatusUpdateRollbackInProgress),
		},
	}

	found := false

	err := h.cfnClient.ListStacksPages(params, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
		for _, stack := range page.StackSummaries {
			if *stack.StackName == stackName {
				found = true
				return false
			}
		}
		return true
	})

	return found, err
}

func (h *cfnHelper) WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error) {
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeset),
	}

	start := time.Now().UTC()
	timeout := time.Second * 60 * 20

	for {
		resp, err := h.cfnClient.DescribeChangeSet(params)

		if err != nil {
			return resp, err
		}

		for _, s := range status {
			if s == *resp.Status {
				return resp, nil
			}
		}

		if time.Since(start) > timeout {
			return resp, fmt.Errorf("Changeset %s failed to reach state %s within %s", changeset, status, timeout)
		}

		time.Sleep(time.Second * 5)
	}
}

func (h *cfnHelper) LogStackEvents(stackID string, logger func(*cloudformation.StackEvent, error)) (cancel func()) {
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second * 5)

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackID),
	}

	go func() {
		defer ticker.Stop()
		var lastEventID string

		for {
			resp, err := h.cfnClient.DescribeStackEvents(params)

			if err != nil {
				logger(nil, err)
			} else {
				if len(resp.StackEvents) > 0 {
					newEvents := resp.StackEvents[:1]

					if lastEventID != "" {
						newEvents = resp.StackEvents

						for i, event := range resp.StackEvents {
							if *event.EventId == lastEventID {
								newEvents = resp.StackEvents[:i]
								break
							}
						}
					}

					for i := len(newEvents) - 1; i >= 0; i-- {
						logger(newEvents[i], nil)
						lastEventID = *newEvents[i].EventId
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

func (h *cfnHelper) uploadChildTemplates(basedir, templateBody, bucket, prefix string) error {
	files := findNestedTemplateFiles(templateBody)

	for _, file := range files {
		if err := h.validateTemplateFile(filepath.Join(basedir, file)); err != nil {
			return err
		}
	}

	s3Helper := NewS3Helper(h.s3Client, h.region, ui.NewPrefixWriter(h.w, "  "))
	err := s3Helper.EnsureBucket(bucket)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := h.uploadTemplateFile(basedir, file, bucket, prefix); err != nil {
			return err
		}
	}

	return nil
}

func findNestedTemplateFiles(templateBody string) []string {
	files := make([]string, 0)
	matches := childTemplateRegexp.FindAllStringSubmatch(templateBody, -1)

	for _, match := range matches {
		files = append(files, match[3])
	}

	return files
}

func updateNestedTemplateURLs(templateBody, region, bucket, prefix string) string {
	repl := fmt.Sprintf("${1}TemplateURL: https://s3-%s.amazonaws.com/%s/%s/$3", region, bucket, prefix)
	return childTemplateRegexp.ReplaceAllString(templateBody, repl)
}

func (h *cfnHelper) uploadTemplateFile(basedir, file, bucket, prefix string) error {
	reader, err := os.Open(path.Join(basedir, file))
	if err != nil {
		return err
	}

	defer reader.Close()

	key := path.Join(prefix, file)

	return h.uploadTemplate(reader, bucket, key)
}

func (h *cfnHelper) uploadTemplate(r io.Reader, bucket, key string) error {
	params := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	}

	fmt.Fprintf(h.w, "Uploading cloudformation template to 's3://%s/%s'\n", bucket, key)

	if _, err := h.uploader.Upload(params); err != nil {
		return err
	}

	return nil
}

func (h *cfnHelper) validateTemplateFile(file string) error {
	fmt.Fprintf(h.w, "Validating cloudformation template '%s'...\n", file)
	templateBody, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	return h.validateTemplate(templateBody)
}

func (h *cfnHelper) validateTemplate(body []byte) error {
	_, err := h.cfnClient.ValidateTemplate(&cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(body)),
	})

	return err
}
