package resources

var environmentDNSCleanerTemplate = `
Parameters:
    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    DNSZone:
        Description: Select the DNS zone to clean
        Type: String

    LogGroupName:
        Description: The name of the cloudwatch log group to send container logs to
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
                        awslogs-group: !Ref LogGroupName
                        awslogs-region: !Ref AWS::Region
                        awslogs-stream-prefix: daemon-services/dns-cleaner

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
`
