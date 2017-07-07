package util

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

func DirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func AnyError(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}

func ClusterConsoleURL(cluster, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/home?region=%s#/clusters/%s/services", region, region, cluster)
}

func ServiceConsoleURL(service, cluster, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/home?region=%s#/clusters/%s/services/%s/tasks", region, region, cluster, GetIDFromArn(service))
}

func CloudFormationConsoleURL(stackID, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stack/detail?stackId=%s", region, region, url.QueryEscape(stackID))
}

func CloudWatchLogsConsoleURL(logGroup, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=%s", region, region, logGroup)
}

func GetIDFromArn(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}

func VersionFromTime(t time.Time) string {
	return t.Format("2006-01-02_15-04-05")
}
