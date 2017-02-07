package templates

import "text/template"

var webServiceCloudFormationTemplate = template.Must(template.New("webServiceCloudFormationFile").Parse(`
Parameters:

    VPC:
        Description: The VPC that the ECS cluster is deployed to
        Type: AWS::EC2::VPC::Id

    Cluster:
        Description: The name of the ECS cluster to deploy to
        Type: String

    DesiredCount:
        Description: The number of instances of the service to run
        Type: Number

    Listener:
        Description: The Application Load Balancer listener to register with
        Type: String

    Path:
        Description: The path to register with the Application Load Balancer
        Type: String

    Port:
        Description: The container port to bind to the ALB
        Type: String

    RoutePriority:
        Description: The priority of Load Balancer listener rule for this service
        Type: String

    TaskDefinition:
        Description: The ARN of the task definition for the service
        Type: String

Resources:

    Service:
        Type: AWS::ECS::Service
        DependsOn: ListenerRule
        Properties:
            Cluster: !Ref Cluster
            Role: !Ref ServiceRole
            DesiredCount: !Ref DesiredCount
            TaskDefinition: !Ref TaskDefinition
            DeploymentConfiguration:
                MaximumPercent: 200
                MinimumHealthyPercent: 100
            LoadBalancers:
                - ContainerName: web
                  ContainerPort: !Ref Port
                  TargetGroupArn: !Ref TargetGroup

    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            LogGroupName: !Ref AWS::StackName
            RetentionInDays: 365

    TargetGroup:
        Type: AWS::ElasticLoadBalancingV2::TargetGroup
        Properties:
            VpcId: !Ref VPC
            Port: 80
            Protocol: HTTP
            Matcher:
                HttpCode: 200-299
            HealthCheckIntervalSeconds: 10
            HealthCheckPath: !Ref Path
            HealthCheckProtocol: HTTP
            HealthCheckTimeoutSeconds: 5
            HealthyThresholdCount: 2

    ListenerRule:
        Type: AWS::ElasticLoadBalancingV2::ListenerRule
        Properties:
            ListenerArn: !Ref Listener
            Priority: !Ref RoutePriority
            Conditions:
                - Field: path-pattern
                  Values:
                    - !Ref Path
            Actions:
                - TargetGroupArn: !Ref TargetGroup
                  Type: forward

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

Outputs:

    TargetGroup:
        Description: Reference to the load balancer target group
        Value: !Ref TargetGroup

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
