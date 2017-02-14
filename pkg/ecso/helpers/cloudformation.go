package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
)

var (
	childTemplateRegexp = regexp.MustCompile(`(\s*)TemplateURL:\s*(\./)*(.+)`)
)

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
	PackageAndDeploy(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error)
	PackageAndCreate(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error)
	Package(templateFile, prefix string) (string, error)
	DeleteStack(serviceName string) error
	Deploy(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	Create(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	StackExists(stackName string) (bool, error)
	WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error)
	GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error)
	GetStackOutputs(stackName string) (map[string]string, error)
}

// NewCloudFormationHelper creates a CloudFormationHelper
func NewCloudFormationHelper(region string, cfnClient cloudformationiface.CloudFormationAPI, s3Client s3iface.S3API, stsClient stsiface.STSAPI, logger ecso.Logger) CloudFormationHelper {
	return &cfnHelper{
		region:    region,
		cfnClient: cfnClient,
		s3Client:  s3Client,
		stsClient: stsClient,
		uploader:  s3manager.NewUploaderWithClient(s3Client),
		logger:    logger,
	}
}

type cfnHelper struct {
	region    string
	cfnClient cloudformationiface.CloudFormationAPI
	uploader  *s3manager.Uploader
	s3Client  s3iface.S3API
	stsClient stsiface.STSAPI
	logger    ecso.Logger
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

func (h *cfnHelper) PackageAndDeploy(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error) {
	templateBody, err := h.Package(templateFile, prefix)

	if err != nil {
		return nil, err
	}

	return h.Deploy(templateBody, stackName, params, tags, dryRun)

}

func (h *cfnHelper) PackageAndCreate(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error) {
	templateBody, err := h.Package(templateFile, prefix)

	if err != nil {
		return nil, err
	}

	return h.Create(templateBody, stackName, params, tags, dryRun)

}

func (h *cfnHelper) Package(templateFile, prefix string) (string, error) {
	if err := h.validateTemplateFile(templateFile); err != nil {
		return "", err
	}

	bucket, err := h.getDefaultCloudFormationBucket()

	if err != nil {
		return "", err
	}

	basedir := filepath.Dir(templateFile)
	templateBody, err := ioutil.ReadFile(templateFile)

	if err != nil {
		return "", err
	}

	if err := h.ensureBucket(bucket); err != nil {
		return "", err
	}

	if err := h.uploadChildTemplates(basedir, string(templateBody), bucket, prefix, os.Open); err != nil {
		return "", err
	}

	return updateNestedTemplateURLs(string(templateBody), h.region, bucket, prefix), nil
}

func (h *cfnHelper) Create(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
	input := &cloudformation.CreateStackInput{
		StackName:       aws.String(stackName),
		DisableRollback: aws.Bool(true),
		Parameters:      make([]*cloudformation.Parameter, 0),
		Tags:            make([]*cloudformation.Tag, 0),
		TemplateBody:    aws.String(templateBody),
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

	resp, err := h.cfnClient.CreateStack(input)

	if err != nil {
		return nil, err
	}

	result := &DeploymentResult{
		StackID: *resp.StackId,
	}

	childLogger := h.logger.Child()

	cancel := h.LogStackEvents(*resp.StackId, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			childLogger.Printf("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	h.logger.Printf("Waiting for stack creation to complete...\n")

	return result, h.cfnClient.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: resp.StackId,
	})
}

func (h *cfnHelper) Deploy(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
	input := &cloudformation.CreateChangeSetInput{
		StackName:           aws.String(stackName),
		ChangeSetName:       aws.String(fmt.Sprintf("%s-%d", stackName, time.Now().Unix())),
		Parameters:          make([]*cloudformation.Parameter, 0),
		Tags:                make([]*cloudformation.Tag, 0),
		TemplateBody:        aws.String(templateBody),
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
		h.logger.Printf("Updating existing '%s' cloudformation stack\n", stackName)
	} else {
		h.logger.Printf("Creating new '%s' cloudformation stack\n", stackName)
		input.ChangeSetType = aws.String("CREATE")
	}

	h.logger.Printf("Creating changeset...\n")

	changeset, err := h.cfnClient.CreateChangeSet(input)

	if err != nil {
		return nil, err
	}

	result := &DeploymentResult{
		StackID:            *changeset.StackId,
		ChangeSetID:        *changeset.Id,
		DidRequireUpdating: true,
	}

	h.logger.Printf("Waiting for changeset %s to be ready...\n", *changeset.Id)

	if changeSetDescription, err := h.WaitForChangeset(*changeset.Id, cloudformation.ChangeSetStatusCreateComplete, cloudformation.ChangeSetStatusFailed); err != nil {
		return result, err
	} else if len(changeSetDescription.Changes) == 0 {
		result.DidRequireUpdating = false
		return result, nil
	}

	h.logger.Printf("Created changeset %s\n", *changeset.Id)

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

	childLogger := h.logger.Child()

	cancel := h.LogStackEvents(*changeset.StackId, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			childLogger.Printf("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	if exists {
		h.logger.Printf("Waiting for stack update to complete...\n")
		return result, h.cfnClient.WaitUntilStackUpdateComplete(stack)
	}

	h.logger.Printf("Waiting for stack creation to complete...\n")
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

	childLogger := h.logger.Child()

	cancel := h.LogStackEvents(stackName, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			childLogger.Printf("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	h.logger.Printf("Waiting for stack delete to complete...\n")

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

func (h *cfnHelper) ensureBucket(bucket string) error {
	params := &s3.HeadBucketInput{
		Bucket: aws.String(bucket), // Required
	}

	_, err := h.s3Client.HeadBucket(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			return h.createBucket(bucket)
		}

		return err
	}

	return nil
}

func (h *cfnHelper) createBucket(bucket string) error {
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

func (h *cfnHelper) uploadChildTemplates(basedir, templateBody, bucket, prefix string, op func(string) (*os.File, error)) error {
	for _, file := range findNestedTemplateFiles(templateBody) {

		if err := h.validateTemplateFile(filepath.Join(basedir, file)); err != nil {
			return err
		}

		reader, err := op(filepath.Join(basedir, file))

		if err != nil {
			return err
		}

		defer reader.Close()

		params := &s3manager.UploadInput{
			Bucket: &bucket,
			Key:    aws.String(path.Join(prefix, file)),
			Body:   reader,
		}

		h.logger.Printf("Uploading template '%s' to 's3://%s/%s'\n", file, bucket, prefix)

		if _, err := h.uploader.Upload(params); err != nil {
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

func (h *cfnHelper) validateTemplate(body []byte) error {
	_, err := h.cfnClient.ValidateTemplate(&cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(body)),
	})

	return err
}

func (h *cfnHelper) validateTemplateFile(file string) error {
	h.logger.Printf("Validating cloudformation template '%s'...\n", file)
	templateBody, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	return h.validateTemplate(templateBody)
}

func (h *cfnHelper) getDefaultCloudFormationBucket() (string, error) {
	resp, err := h.stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ecso-%s-%s", h.region, *resp.Account), nil
}
