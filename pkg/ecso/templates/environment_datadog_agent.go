package templates

import "text/template"

var environmentDataDogTemplate = template.Must(template.New("environmentDataDogTemplate").Parse(`
Parameters:
    DataDogAPIKey:
        Description: Please provide your datadog API key
        Type: String

    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

Resources:
    TaskDefinition:
        Type: AWS::ECS::TaskDefinition
        Properties:
            Family: !Sub ${EnvironmentName}-datadog-agent
            Volumes:
                - Name: docker_sock
                  Host:
                    SourcePath: /var/run/docker.sock
                - Name: proc
                  Host:
                    SourcePath: /proc/
                - Name: cgroup
                  Host:
                    SourcePath: /cgroup/
            ContainerDefinitions:
                - Name: dd-agent
                  Essential: true
                  Image: datadog/docker-dd-agent:latest
                  Cpu: 10
                  Memory: 128
                  PortMappings:
                    - HostPort: 8125
                      ContainerPort: 8125
                      Protocol: udp
                  Environment:
                    - Name: DD_API_KEY
                      Value: !Ref DataDogAPIKey
                    - Name: DD_TAGS
                      Value: !Sub ecs-cluster:${EnvironmentName}
                    - Name: NON_LOCAL_TRAFFIC
                      Value: 1
                    - Name: SERVICE_8125_NAME
                      Value: !Sub datadog.${EnvironmentName}
                  MountPoints:
                    - ContainerPath: /var/run/docker.sock
                      SourceVolume: docker_sock
                    - ContainerPath: /host/sys/fs/cgroup
                      SourceVolume: cgroup
                      ReadOnly: true
                    - ContainerPath: /host/proc
                      SourceVolume: proc
                      ReadOnly: true
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

    # This IAM Role grants the service access to register/unregister with the
    # Application Load Balancer (ALB). It is based on the default documented here:
    # http://docs.aws.amazon.com/AmazonECS/latest/developerguide/service_IAM_role.html
    ServiceRole:
        Type: AWS::IAM::Role
        Properties:
            RoleName: !Sub ecs-service-${AWS::StackName}
            Path: /
            AssumeRolePolicyDocument: |
                {
                    "Statement": [{
                        "Effect": "Allow",
                        "Principal": { "Service": [ "ecs.amazonaws.com" ]},
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
                                    "ec2:AuthorizeSecurityGroupIngress",
                                    "ec2:Describe*",
                                    "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
                                    "elasticloadbalancing:Describe*",
                                    "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
                                    "elasticloadbalancing:DeregisterTargets",
                                    "elasticloadbalancing:DescribeTargetGroups",
                                    "elasticloadbalancing:DescribeTargetHealth",
                                    "elasticloadbalancing:RegisterTargets"
                                ],
                                "Resource": "*"
                        }]
                    }
`))
