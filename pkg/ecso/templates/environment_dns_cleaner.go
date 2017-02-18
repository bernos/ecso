package templates

import "text/template"

var environmentDNSCleanerTemplate = template.Must(template.New("environmentDNSCleanerTemplate").Parse(`
Parameters:
    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    DNSZone:
        Description: Select the DNS zone to clean
        Type: String

Resources:
    TaskDefinition:
        Type: AWS::ECS::TaskDefinition
        Properties:
            Family: !Sub ${EnvironmentName}-dns-cleaner
            TaskRoleArn: !Ref TaskRole
            ContainerDefinitions:
                - Name: dns-cleaner
                  Command:
                      - go-wrapper
                      - run
                      - -region
                      - !Ref AWS::Region
                      - -zone
                      - !Sub "${DNSZone}."
                      - -records
                      - !Sub "*.${EnvironmentName}.${DNSZone}."
                  Essential: true
                  Image: bernos/ecso-dns-cleaner:latest
                  Cpu: 10
                  Memory: 128
                  LogConfiguration:
                    LogDriver: awslogs
                    Options:
                        awslogs-group: !Ref AWS::StackName
                        awslogs-region: !Ref AWS::Region

    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            LogGroupName: !Ref AWS::StackName
            RetentionInDays: 30

    TaskRole:
        Type: AWS::IAM::Role
        Properties:
            RoleName: !Sub ${EnvironmentName}-dns-cleaner-role
            Path: /
            AssumeRolePolicyDocument: |
                {
                    "Statement": [{
                        "Effect": "Allow",
                        "Principal": { "Service": [ "ecs-tasks.amazonaws.com" ]},
                        "Action": [ "sts:AssumeRole" ]
                    }]
                }
            Policies:
                - PolicyName: !Sub ecs-service-${AWS::StackName}
                  PolicyDocument:
                    {
                        "Version": "2012-10-17",
                        "Statement": [{
                                "Effect": "Allow",
                                "Action": [
                                    "ec2:Describe*",
                                    "route53:*"
                                ],
                                "Resource": "*"
                        }]
                    }
`))
