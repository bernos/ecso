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
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/util"
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
	// PackageAndCreate(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error)
	Package(templateFile, prefix string) (string, error)
	DeleteStack(serviceName string) error
	DeployTemplateURL(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	// CreateTemplateURL(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	StackExists(stackName string) (bool, error)
	WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error)
	GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error)
	GetStackOutputs(stackName string) (map[string]string, error)
}

// NewCloudFormationHelper creates a CloudFormationHelper
func NewCloudFormationHelper(region string, cfnClient cloudformationiface.CloudFormationAPI, s3Client s3iface.S3API, stsClient stsiface.STSAPI, logger log.Logger) CloudFormationHelper {
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
	logger    log.Logger
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

	return h.DeployTemplateURL(templateBody, stackName, params, tags, dryRun)

}

func (h *cfnHelper) PackageAndCreate(stackName, templateFile, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error) {
	templateBody, err := h.Package(templateFile, prefix)

	if err != nil {
		return nil, err
	}

	return h.CreateTemplateURL(templateBody, stackName, params, tags, dryRun)

}

func (h *cfnHelper) Package(templateFile, prefix string) (string, error) {
	if err := h.validateTemplateFile(templateFile); err != nil {
		return "", err
	}

	bucket, err := util.GetEcsoBucket(h.stsClient, h.region)
	if err != nil {
		return "", err
	}

	templatePrefix := path.Join(prefix, "templates")
	basedir := filepath.Dir(templateFile)
	templateBody, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", err
	}

	if err := h.uploadChildTemplates(basedir, string(templateBody), bucket, templatePrefix); err != nil {
		return "", err
	}

	body := updateNestedTemplateURLs(string(templateBody), h.region, bucket, templatePrefix)
	key := path.Join(templatePrefix, "stack.yaml")

	if err := h.uploadTemplate(strings.NewReader(body), bucket, key); err != nil {
		return "", err
	}

	// TODO: upload tags and params. Return prefix to package in S3, and change deploy/create funcs so that
	// they accept the base prefix, rather than the full template url
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", h.region, bucket, key), nil
}

func (h *cfnHelper) CreateTemplateURL(templateURL, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
	input := &cloudformation.CreateStackInput{
		StackName:       aws.String(stackName),
		DisableRollback: aws.Bool(true),
		Parameters:      make([]*cloudformation.Parameter, 0),
		Tags:            make([]*cloudformation.Tag, 0),
		TemplateURL:     aws.String(templateURL),
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

func (h *cfnHelper) DeployTemplateURL(templateURL, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
	input := &cloudformation.CreateChangeSetInput{
		StackName:           aws.String(stackName),
		ChangeSetName:       aws.String(fmt.Sprintf("%s-%d", stackName, time.Now().Unix())),
		Parameters:          make([]*cloudformation.Parameter, 0),
		Tags:                make([]*cloudformation.Tag, 0),
		TemplateURL:         aws.String(templateURL),
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

func (h *cfnHelper) uploadChildTemplates(basedir, templateBody, bucket, prefix string) error {
	files := findNestedTemplateFiles(templateBody)

	for _, file := range files {
		if err := h.validateTemplateFile(filepath.Join(basedir, file)); err != nil {
			return err
		}
	}

	s3Helper := NewS3Helper(h.s3Client, h.region, h.logger)
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

	h.logger.Printf("Uploading cloudformation template to 's3://%s/%s'\n", bucket, key)

	if _, err := h.uploader.Upload(params); err != nil {
		return err
	}

	return nil
}

func (h *cfnHelper) validateTemplateFile(file string) error {
	h.logger.Printf("Validating cloudformation template '%s'...\n", file)
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
