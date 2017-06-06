package templates

import "text/template"

var environmentLoggingTemplate = template.Must(template.New("environmentLoggingTemplate").Parse(`
Description: >
    This template deploys a cloudwatch logs logging group for the ecso environment

Resources:
    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            RetentionInDays: 30

Outputs:
    LogGroup:
        Description: A reference to the CloudWatch logs group
        Value: !Ref CloudWatchLogsGroup
`))
