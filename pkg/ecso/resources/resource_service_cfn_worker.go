package resources

import "text/template"

var WorkerServiceCloudFormationTemplate = NewCloudFormationTemplate("stack.yaml", workerCloudFormationTemplate)

var workerCloudFormationTemplate = template.Must(template.New("workerCloudFormationFile").Parse(`
Parameters:

    AlertsTopic:
        Description: The ARN of the SNS topic to send alarm notifications to
        Type: String

    Cluster:
        Description: The name of the ECS cluster to deploy to
        Type: String

    DesiredCount:
        Description: The number of instances of the service to run
        Type: Number

    TaskDefinition:
        Description: The ARN of the task definition for the service
        Type: String

    Version:
        Description: The version of the service
        Type: String

Resources:

    Service:
        Type: AWS::ECS::Service
        Properties:
            Cluster: !Ref Cluster
            DesiredCount: !Ref DesiredCount
            TaskDefinition: !Ref TaskDefinition
            DeploymentConfiguration:
                MaximumPercent: 200
                MinimumHealthyPercent: 100

    TaskCountAlarm:
        Type: AWS::CloudWatch::Alarm
        Properties:
            AlarmName: !Sub ${Service}-task-count
            AlarmDescription: Not enough tasks running
            Namespace: AWS/ECS
            MetricName: CPUUtilization
            Statistic: SampleCount
            Period: 120
            EvaluationPeriods: 2
            Threshold: !Ref DesiredCount
            ComparisonOperator: LessThanThreshold
            AlarmActions:
                - !Ref AlertsTopic
            Dimensions:
                - Name: ClusterName
                  Value: !Ref Cluster
                - Name: ServiceName
                  Value: !Sub ${Service.Name}

    CPUUtilizationAlarm:
        Type: AWS::CloudWatch::Alarm
        Properties:
            AlarmName: !Sub ${Service}-alarm-cpu-utilisation
            AlarmDescription: CPU utilisation is high
            Namespace: AWS/ECS
            MetricName: CPUUtilization
            Statistic: Maximum
            Period: 60
            EvaluationPeriods: 2
            Threshold: 80
            ComparisonOperator: GreaterThanThreshold
            AlarmActions:
                - !Ref AlertsTopic
            Dimensions:
                - Name: ClusterName
                  Value: !Ref Cluster
                - Name: ServiceName
                  Value: !Sub ${Service.Name}

    MemoryUtilizationAlarm:
        Type: AWS::CloudWatch::Alarm
        Properties:
            AlarmName: !Sub ${Service}-alarm-memory-utilisation
            AlarmDescription: Memory utilisation is high
            Namespace: AWS/ECS
            MetricName: MemoryUtilization
            Statistic: Maximum
            Period: 60
            EvaluationPeriods: 2
            Threshold: 80
            ComparisonOperator: GreaterThanThreshold
            AlarmActions:
                - !Ref AlertsTopic
            Dimensions:
                - Name: ClusterName
                  Value: !Ref Cluster
                - Name: ServiceName
                  Value: !Sub ${Service.Name}

    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            LogGroupName: !Ref AWS::StackName
            RetentionInDays: 365

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

Outputs:

    ServiceRole:
        Description: The IAM role for the service
        Value: !Ref ServiceRole

    CloudWatchLogsGroup:
        Description: Reference to the cloudwatch logs group
        Value: !Ref CloudWatchLogsGroup

    Service:
        Description: Reference to the ecs service
        Value: !Ref Service
`))
