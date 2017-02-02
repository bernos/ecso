package services

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
)

var (
	childTemplateRegexp = regexp.MustCompile(`(\s*)TemplateURL:\s*(\./)*(.+)`)
)

type DeploymentResult struct {
	StackID            string
	ChangeSetID        string
	DidRequireUpdating bool
}

type CloudFormationService interface {
	PackageAndDeploy(stackName, templateFile, bucket, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error)
	PackageAndCreate(stackName, templateFile, bucket, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error)
	Package(templateFile, bucket, prefix string) (string, error)
	DeleteStack(serviceName string) error
	Deploy(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	Create(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error)
	StackExists(stackName string) (bool, error)
	WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error)
	GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error)
	GetStackOutputs(stackName string) (map[string]string, error)
	LogStackOutputs(stackName string, logfn func(name, value string)) error
}

func NewCloudFormationService(region string, cfnClient cloudformationiface.CloudFormationAPI, s3Client s3iface.S3API, log func(string, ...interface{})) CloudFormationService {
	return &cfnService{
		region:    region,
		cfnClient: cfnClient,
		s3Client:  s3Client,
		uploader:  s3manager.NewUploaderWithClient(s3Client),
		log:       log,
	}
}

type cfnService struct {
	region    string
	cfnClient cloudformationiface.CloudFormationAPI
	uploader  *s3manager.Uploader
	s3Client  s3iface.S3API
	log       func(string, ...interface{})
}

func (svc *cfnService) GetStackOutputs(stackName string) (map[string]string, error) {
	resp, err := svc.cfnClient.DescribeStacks(&cloudformation.DescribeStacksInput{
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

func (svc *cfnService) PackageAndDeploy(stackName, templateFile, bucket, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error) {
	templateBody, err := svc.Package(templateFile, bucket, prefix)

	if err != nil {
		return nil, err
	}

	return svc.Deploy(templateBody, stackName, params, tags, dryRun)

}

func (svc *cfnService) PackageAndCreate(stackName, templateFile, bucket, prefix string, tags, params map[string]string, dryRun bool) (*DeploymentResult, error) {
	templateBody, err := svc.Package(templateFile, bucket, prefix)

	if err != nil {
		return nil, err
	}

	return svc.Create(templateBody, stackName, params, tags, dryRun)

}

func (svc *cfnService) Package(templateFile, bucket, prefix string) (string, error) {
	// TODO: validate the template

	basedir := filepath.Dir(templateFile)
	templateBody, err := ioutil.ReadFile(templateFile)

	if err != nil {
		return "", err
	}

	if err := svc.ensureBucket(bucket); err != nil {
		return "", err
	}

	if err := svc.uploadChildTemplates(basedir, string(templateBody), bucket, prefix, os.Open); err != nil {
		return "", err
	}

	return updateNestedTemplateURLs(string(templateBody), svc.region, bucket, prefix), nil
}

func (svc *cfnService) Create(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
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

	resp, err := svc.cfnClient.CreateStack(input)

	if err != nil {
		return nil, err
	}

	result := &DeploymentResult{
		StackID: *resp.StackId,
	}

	cancel := svc.LogStackEvents(*resp.StackId, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			svc.log("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	svc.log("Waiting for stack creation to complete...\n")

	return result, svc.cfnClient.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: resp.StackId,
	})
}

func (svc *cfnService) Deploy(templateBody, stackName string, params, tags map[string]string, dryRun bool) (*DeploymentResult, error) {
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

	exists, err := svc.StackExists(stackName)

	if err != nil {
		return nil, err
	}

	if exists {
		input.ChangeSetType = aws.String("UPDATE")
		svc.log("Updating existing '%s' cloudformation stack\n", stackName)
	} else {
		svc.log("Creating new '%s' cloudformation stack\n", stackName)
		input.ChangeSetType = aws.String("CREATE")
	}

	svc.log("Creating changeset...\n")

	changeset, err := svc.cfnClient.CreateChangeSet(input)

	if err != nil {
		return nil, err
	}

	result := &DeploymentResult{
		StackID:            *changeset.StackId,
		ChangeSetID:        *changeset.Id,
		DidRequireUpdating: true,
	}

	svc.log("Waiting for changeset %s to be ready...\n", *changeset.Id)

	if changeSetDescription, err := svc.WaitForChangeset(*changeset.Id, cloudformation.ChangeSetStatusCreateComplete, cloudformation.ChangeSetStatusFailed); err != nil {
		return result, err
	} else if len(changeSetDescription.Changes) == 0 {
		result.DidRequireUpdating = false
		return result, nil
	}

	svc.log("Created changeset %s\n", *changeset.Id)

	if dryRun {
		return result, nil
	}

	if _, err := svc.cfnClient.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName: changeset.Id,
		StackName:     changeset.StackId,
	}); err != nil {
		return result, err
	}

	stack := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	cancel := svc.LogStackEvents(*changeset.StackId, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			svc.log("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	if exists {
		svc.log("Waiting for stack update to complete...\n")
		return result, svc.cfnClient.WaitUntilStackUpdateComplete(stack)
	} else {
		svc.log("Waiting for stack creation to complete...\n")
		return result, svc.cfnClient.WaitUntilStackCreateComplete(stack)
	}
}

func (svc *cfnService) GetChangeSet(changeset string) (*cloudformation.DescribeChangeSetOutput, error) {
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeset),
	}

	return svc.cfnClient.DescribeChangeSet(params)
}

func (svc *cfnService) DeleteStack(stackName string) error {
	_, err := svc.cfnClient.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return err
	}

	cancel := svc.LogStackEvents(stackName, func(ev *cloudformation.StackEvent, err error) {
		if ev != nil {
			svc.log("%s: %s\n", *ev.LogicalResourceId, *ev.ResourceStatus)
		}
	})

	defer cancel()

	svc.log("Waiting for stack delete to complete...\n")

	return svc.cfnClient.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
}

func (svc *cfnService) StackExists(stackName string) (bool, error) {
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

	err := svc.cfnClient.ListStacksPages(params, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
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

func (svc *cfnService) WaitForChangeset(changeset string, status ...string) (*cloudformation.DescribeChangeSetOutput, error) {
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeset),
	}

	start := time.Now().UTC()
	timeout := time.Second * 60 * 20

	for {
		resp, err := svc.cfnClient.DescribeChangeSet(params)

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

func (svc *cfnService) LogStackEvents(stackID string, logger func(*cloudformation.StackEvent, error)) (cancel func()) {
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second * 5)

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackID),
	}

	go func() {
		defer ticker.Stop()
		var lastEventID string

		for {
			resp, err := svc.cfnClient.DescribeStackEvents(params)

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

func (svc *cfnService) LogStackOutputs(stackName string, logfn func(name, value string)) error {
	outputs, err := svc.GetStackOutputs(stackName)

	if err != nil {
		return err
	}

	for k, v := range outputs {
		logfn(k, v)
	}

	return nil
}

func (svc *cfnService) ensureBucket(bucket string) error {
	params := &s3.HeadBucketInput{
		Bucket: aws.String(bucket), // Required
	}

	_, err := svc.s3Client.HeadBucket(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			return svc.createBucket(bucket)
		}

		return err
	}

	return nil
}

func (svc *cfnService) createBucket(bucket string) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucket), // Required
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(svc.region),
		},
	}

	svc.log("Creating bucket '%s' in region '%s'\n", bucket, svc.region)

	_, err := svc.s3Client.CreateBucket(params)

	return err
}

func (svc *cfnService) uploadChildTemplates(basedir, templateBody, bucket, prefix string, op func(string) (*os.File, error)) error {
	for _, file := range findNestedTemplateFiles(templateBody) {
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

		svc.log("Uploading template '%s' to 's3://%s/%s'\n", file, bucket, prefix)

		if _, err := svc.uploader.Upload(params); err != nil {
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

	// repl = "${1}TemplateURL: https://bucket/foo/$3"
	return childTemplateRegexp.ReplaceAllString(templateBody, repl)
}
