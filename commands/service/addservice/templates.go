package addservice

import "text/template"

var composeFileTemplate = template.Must(template.New("composeFile").Parse(`
version: '2'

volumes:
  nginxdata: {}

services:
  web:
    image: nginx:latest
    mem_limit: 20000000
    logging:
      driver: awslogs
      options:
        awslogs-region: ${ECSO_AWS_REGION}
        awslogs-group: ${ECSO_CLUSTER_NAME}-{{.Service.Name}}
    ports:
      - "0:80"
    volumes:
      - nginxdata:/usr/share/nginx/html/:ro
    command: /bin/bash -c "echo \"server { location / { root /usr/share/nginx/html; try_files \$$uri /index.html =404; } }\" > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"
  backend:
    image: busybox:latest
    mem_limit: 10000000
    logging:
      driver: awslogs
      options:
        awslogs-region: ${ECSO_AWS_REGION}
        awslogs-group: ${ECSO_CLUSTER_NAME}-{{.Service.Name}}
    volumes:
      - nginxdata:/nginx
    command: sh -c "while true; do echo \"This is the {{.Service.Name}} service <p><pre>` + "`env`" + `</pre></p> \" > /nginx/index.html; sleep 3; done"
`))

var cloudFormationTemplate = template.Must(template.New("cloudFormationFile").Parse(`
Parameters:

    VPC:
        Description: The VPC that the ECS cluster is deployed to
        Type: AWS::EC2::VPC::Id

    Listener:
        Description: The Application Load Balancer listener to register with
        Type: String

    Path:
        Description: The path to register with the Application Load Balancer
        Type: String

    RoutePriority:
        Description: The priority of Load Balancer listener rule for this service
        Type: String

Resources:

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

    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            LogGroupName: !Ref AWS::StackName
            RetentionInDays: 365

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
`))
