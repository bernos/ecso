package resources

var environmentLoggingTemplate = `
Description: >
    This template deploys a cloudwatch logs logging group for the ecso environment

Parameters:
    LogGroupName:
        Description: The name of the log group to create
        Type: String

Resources:
    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            LogGroupName: !Ref LogGroupName
            RetentionInDays: 30

Outputs:
    LogGroup:
        Description: A reference to the CloudWatch logs group
        Value: !Ref CloudWatchLogsGroup
`
