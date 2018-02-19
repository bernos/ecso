package helpers

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bernos/ecso/pkg/ecso/api/mocks"
)

func newCloudFormationHelperWithMocks() *cfnHelper {
	s3Client := &mocks.S3APIMock{}

	return &cfnHelper{
		region:    "test-region",
		cfnClient: &mocks.CloudFormationAPIMock{},
		s3Client:  s3Client,
		stsClient: &mocks.STSMock{},
		uploader:  s3manager.NewUploaderWithClient(s3Client),
	}
}

func TestGetStackOutputs(t *testing.T) {
	expect := map[string]string{
		"foo": "bar",
		"baz": "cat",
	}

	mock := &mocks.CloudFormationAPIMock{}
	mock.DescribeStacksReturns(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			&cloudformation.Stack{
				StackName: aws.String("foo"),
				Outputs: []*cloudformation.Output{
					&cloudformation.Output{
						OutputKey:   aws.String("foo"),
						OutputValue: aws.String("bar"),
					},
					&cloudformation.Output{
						OutputKey:   aws.String("baz"),
						OutputValue: aws.String("cat"),
					},
				},
			},
			&cloudformation.Stack{
				StackName: aws.String("bar"),
				Outputs: []*cloudformation.Output{
					&cloudformation.Output{
						OutputKey:   aws.String("food"),
						OutputValue: aws.String("bard"),
					},
					&cloudformation.Output{
						OutputKey:   aws.String("bazd"),
						OutputValue: aws.String("catd"),
					},
				},
			},
		},
	}, nil)

	helper := newCloudFormationHelperWithMocks()
	helper.cfnClient = mock

	result, err := helper.GetStackOutputs("foo")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, expect) {
		t.Errorf("Expected %s, got %s", expect, result)
	}
}

func TestFindNestedTemplateFiles(t *testing.T) {
	wants := []string{
		"infrastructure/security-groups.yaml",
		"infrastructure/load-balancers.yaml",
		"infrastructure/ecs-cluster.yaml",
	}

	gots := findNestedTemplateFiles(MustReadFile(t, "./testdata/root_template.yaml"))

	if len(wants) != len(gots) {
		t.Errorf("Want %v, got %v", wants, gots)
	} else {
		for i, want := range wants {
			if want != gots[i] {
				t.Errorf("Want %s, got %s", want, gots[i])
			}
		}
	}
}

func TestUpdateNestedTemplateURLs(t *testing.T) {
	want := MustReadFile(t, "./testdata/packaged_root_template.yaml")
	got := updateNestedTemplateURLs(MustReadFile(t, "./testdata/root_template.yaml"), "ap-southeast-2", "bucketname", "my/bucket/prefix")

	if want != got {
		t.Errorf("Want %s, got %s", want, got)
	}
}

func MustReadFile(t *testing.T, filename string) string {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		t.Fatalf("Failed to ReadFile(%s) : %s", filename, err.Error())
	}

	return string(data)
}
