package templates

import "text/template"

var environmentClusterTemplate = template.Must(template.New("environmentClusterTemplate").Parse(`
Description: >
    This template deploys an ECS cluster to the provided VPC and subnets
    using an Auto Scaling Group

Parameters:

    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    InstanceType:
        Description: Which instance type should we use to build the ECS cluster?
        Type: String
        Default: c4.large

    ClusterSize:
        Description: How many ECS hosts do you want to initially deploy?
        Type: Number
        Default: 4

    VPC:
        Description: Choose which VPC this ECS cluster should be deployed to
        Type: AWS::EC2::VPC::Id

    Subnets:
        Description: Choose which subnets this ECS cluster should be deployed to
        Type: List<AWS::EC2::Subnet::Id>

    SecurityGroup:
        Description: Select the Security Group to use for the ECS cluster hosts
        Type: AWS::EC2::SecurityGroup::Id

    DNSZone:
        Description: Select the DNS zone to use for service discovery
        Type: String

    DataDogAPIKey:
        Description: Please provide your datadog API key
        Type: String
        Default: ""

    NotificationsTopic:
        Description: The arn of the sns topic to send ecs cluster events to
        Type: String

Mappings:

    # These are the latest ECS optimized AMIs as of November 2016:
    #
    #   amzn-ami-2016.09.b-amazon-ecs-optimized
    #   ECS agent:    1.13.1
    #   Docker:       1.11.2
    #   ecs-init:     1.13.1-1
    #
    # You can find the latest available on this page of our documentation:
    # http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-optimized_AMI.html
    # (note the AMI identifier is region specific)

    AWSRegionToAMI:
        us-east-1:
            AMI: ami-d69c74c0
        us-east-2:
            AMI: ami-64270201
        us-west-1:
            AMI: ami-bc90c2dc
        us-west-2:
            AMI: ami-8e7bc4ee
        eu-west-1:
            AMI: ami-48f9a52e
        eu-central-1:
            AMI: ami-6b428d04
        ap-northeast-1:
            AMI: ami-372f5450
        ap-southeast-1:
            AMI: ami-69208a0a
        ap-southeast-2:
            AMI: ami-307f7853

Resources:

    ECSCluster:
        Type: AWS::ECS::Cluster
        Properties:
            ClusterName: !Ref EnvironmentName

    CloudWatchEvents:
        Type: AWS::Events::Rule
        Properties:
            Name: !Sub ${EnvironmentName}-Events
            State: ENABLED
            EventPattern:
                source:
                    - "aws.ecs"
                detail-type:
                    - "ECS Task State Change"
                    - "ECS Container Instance State Change"
                detail:
                    clusterArn:
                        - !Sub arn:aws:ecs:${AWS::Region}:${AWS::AccountId}:cluster/${EnvironmentName}
            Targets:
                - Arn: !GetAtt EventLoggingLambda.Arn
                  Id: !Sub ${EnvironmentName}-Events-Lambda-Target

    InvokeLambdaPermission:
        Type: AWS::Lambda::Permission
        Properties:
            FunctionName:
                Ref: EventLoggingLambda
            Action: "lambda:InvokeFunction"
            Principal: "events.amazonaws.com"
            SourceArn: !GetAtt CloudWatchEvents.Arn

    CloudWatchLogsGroup:
        Type: AWS::Logs::LogGroup
        Properties:
            RetentionInDays: 7

    ECSAutoScalingGroup:
        Type: AWS::AutoScaling::AutoScalingGroup
        Properties:
            VPCZoneIdentifier: !Ref Subnets
            LaunchConfigurationName: !Ref ECSLaunchConfiguration
            MinSize: !Ref ClusterSize
            MaxSize: !Ref ClusterSize
            DesiredCapacity: !Ref ClusterSize
            Tags:
                - Key: Name
                  Value: !Sub ${EnvironmentName} ECS host
                  PropagateAtLaunch: true
        CreationPolicy:
            ResourceSignal:
                Timeout: PT15M
        UpdatePolicy:
            AutoScalingRollingUpdate:
                MinInstancesInService: 1
                MaxBatchSize: 1
                PauseTime: PT15M
                WaitOnResourceSignals: true

    ECSLaunchConfiguration:
        Type: AWS::AutoScaling::LaunchConfiguration
        Properties:
            ImageId:  !FindInMap [AWSRegionToAMI, !Ref "AWS::Region", AMI]
            InstanceType: !Ref InstanceType
            SecurityGroups:
                - !Ref SecurityGroup
            IamInstanceProfile: !Ref ECSInstanceProfile
            UserData:
                "Fn::Base64": !Sub |
                    #!/bin/bash
                    echo ECS_CLUSTER=${ECSCluster} >> /etc/ecs/ecs.config

                    yum install -y aws-cfn-bootstrap aws-cli jq

                    /opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration --configsets install_all
                    /opt/aws/bin/cfn-signal -e $? --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSAutoScalingGroup

        Metadata:
            AWS::CloudFormation::Init:
                configSets:
                    install_all:
                        - install_cfn
                        - install_logs
                        - install_ecssd_agent
                        - install_dd_agent
                        - install_dns_cleaner

                install_dns_cleaner:
                    files:
                        "/etc/init/dns-cleaner.conf":
                            mode: "000644"
                            owner: root
                            group: root
                            content: !Sub |
                                description "Amazon EC2 Container Service (start task on instance boot)"
                                author "Amazon Web Services"
                                start on started ecs

                                script
                                    exec 2>>/var/log/ecs/ecs-start-task.log
                                    set -x
                                    until curl -s http://localhost:51678/v1/metadata
                                    do
                                        sleep 1
                                    done

                                    # Grab the container instance ARN and AWS region from instance metadata
                                    instance_arn=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .ContainerInstanceArn' | awk -F/ '{print $NF}' )
                                    cluster=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .Cluster' | awk -F/ '{print $NF}' )
                                    region=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .ContainerInstanceArn' | awk -F: '{print $4}')

                                    # Specify the task definition to run at launch
                                    task_definition=${EnvironmentName}-dns-cleaner

                                    # Run the AWS CLI start-task command to start your task on this container instance
                                    aws ecs start-task --cluster $cluster --task-definition $task_definition --container-instances $instance_arn --started-by $instance_arn --region $region
                                end script

                install_dd_agent:
                    files:
                        "/etc/init/datadog-agent.conf":
                            mode: "000644"
                            owner: root
                            group: root
                            content: !Sub |
                                description "Amazon EC2 Container Service (start task on instance boot)"
                                author "Amazon Web Services"
                                start on started ecs

                                script
                                    exec 2>>/var/log/ecs/ecs-start-task.log
                                    set -x
                                    until curl -s http://localhost:51678/v1/metadata
                                    do
                                        sleep 1
                                    done

                                    # Grab the container instance ARN and AWS region from instance metadata
                                    instance_arn=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .ContainerInstanceArn' | awk -F/ '{print $NF}' )
                                    cluster=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .Cluster' | awk -F/ '{print $NF}' )
                                    region=$(curl -s http://localhost:51678/v1/metadata | jq -r '. | .ContainerInstanceArn' | awk -F: '{print $4}')

                                    # Specify the task definition to run at launch
                                    task_definition=${EnvironmentName}-datadog-agent

                                    # Set the datadog api key. If this is empty, we won't actually start the container
                                    dd_api_key=${DataDogAPIKey}

                                    if [ -n "$dd_api_key" ]; then
                                        # Run the AWS CLI start-task command to start your task on this container instance
                                        aws ecs start-task --cluster $cluster --task-definition $task_definition --container-instances $instance_arn --started-by $instance_arn --region $region --overrides "{\"containerOverrides\":[{\"name\":\"dd-agent\", \"environment\":[{\"name\":\"DD_API_KEY\",\"value\":\"$dd_api_key\"}]}]}"
                                    fi
                                end script

                install_cfn:
                    files:
                        "/etc/cfn/cfn-hup.conf":
                            mode: "000400"
                            owner: root
                            group: root
                            content: !Sub |
                                [main]
                                stack=${AWS::StackId}
                                region=${AWS::Region}

                        "/etc/cfn/hooks.d/cfn-auto-reloader.conf":
                            content: !Sub |
                                [cfn-auto-reloader-hook]
                                triggers=post.update
                                path=Resources.ContainerInstances.Metadata.AWS::CloudFormation::Init
                                action=/opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration --configsets install_all
                                runas=root

                    services:
                        sysvinit:
                            cfn-hup:
                                enabled: true
                                ensureRunning: true
                                files:
                                    - /etc/cfn/cfn-hup.conf
                                    - /etc/cfn/hooks.d/cfn-auto-reloader.conf

                install_logs:
                    packages:
                        yum:
                            awslogs: []

                    commands:
                        01_create_state_directory:
                            command: "mkdir -p /var/awslogs/state"

                    files:
                        "/etc/awslogs/awscli.conf":
                            mode: "000400"
                            owner: root
                            group: root
                            content: !Sub |
                                [plugins]
                                cwlogs = cwlogs
                                [default]
                                region = ${AWS::Region}

                        "/etc/awslogs/awslogs.conf":
                            mode: "000400"
                            owner: root
                            group: root
                            content: !Sub |
                                [general]
                                state_file = /var/awslogs/state/agent-state
                                [/var/log/ecs/ecs-start-task.log]
                                file = /var/log/ecs/ecs-start-task.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/ecs-start-task.log
                                datetime_format =
                                [/var/log/cloud-init.log]
                                file = /var/log/cloud-init.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/cloud-init.log
                                datetime_format =
                                [/var/log/cloud-init-output.log]
                                file = /var/log/cloud-init-output.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/cloud-init-output.log
                                datetime_format =
                                [/var/log/cfn-init.log]
                                file = /var/log/cfn-init.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/cfn-init.log
                                datetime_format =
                                [/var/log/cfn-hup.log]
                                file = /var/log/cfn-hup.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/cfn-hup.log
                                datetime_format =
                                [/var/log/cfn-wire.log]
                                file = /var/log/cfn-wire.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/cfn-wire.log
                                datetime_format =
                                [/var/log/ecssd_agent.log]
                                file = /var/log/ecssd_agent.log
                                log_group_name = ${CloudWatchLogsGroup}
                                log_stream_name = {instance_id}/ecssd-agent.log
                                datetime_format = %Y-%m-%dT%H:%M:%S%z

                    services:
                        sysvinit:
                            awslogs:
                                enabled: true
                                unsureRunning: true
                                files:
                                    - /etc/awslogs/awslogs.conf

                install_ecssd_agent:
                    commands:
                        start_ecssd_agent:
                            command: "start ecssd-agent"

                    files:
                        "/etc/init/ecssd-agent.conf":
                            mode: "000644"
                            owner: root
                            group: root
                            content: !Sub |
                                description "Amazon EC2 Container Service Discovery"
                                author "Javieros Ros"
                                start on stopped rc RUNLEVEL=[345]
                                exec /usr/local/bin/ecssd_agent ${DNSZone} >> /var/log/ecssd_agent.log 2>&1
                                respawn

                        "/usr/local/bin/ecssd_agent":
                            source: https://github.com/awslabs/service-discovery-ecs-dns/releases/download/1.2/ecssd_agent
                            mode: "000755"
                            owner: root
                            group: root

    # This IAM Role is attached to all of the ECS hosts. It is based on the default role
    # published here:
    # http://docs.aws.amazon.com/AmazonECS/latest/developerguide/instance_IAM_role.html
    #
    # You can add other IAM policy statements here to allow access from your ECS hosts
    # to other AWS services. Please note that this role will be used by ALL containers
    # running on the ECS host.
    ECSRole:
        Type: AWS::IAM::Role
        Properties:
            Path: /
            RoleName: !Sub ${EnvironmentName}-ECSRole-${AWS::Region}
            AssumeRolePolicyDocument: |
                {
                    "Statement": [{
                        "Action": "sts:AssumeRole",
                        "Effect": "Allow",
                        "Principal": {
                            "Service": "ec2.amazonaws.com"
                        }
                    }]
                }
            Policies:
                - PolicyName: ecs-service
                  PolicyDocument: |
                    {
                        "Statement": [{
                            "Effect": "Allow",
                            "Action": [
                                "ecs:CreateCluster",
                                "ecs:DeregisterContainerInstance",
                                "ecs:DiscoverPollEndpoint",
                                "ecs:Poll",
                                "ecs:RegisterContainerInstance",
                                "ecs:StartTelemetrySession",
                                "ecs:StartTask",
                                "ecs:Submit*",
                                "logs:CreateLogGroup",
                                "logs:CreateLogStream",
                                "logs:PutLogEvents",
                                "logs:DescribeLogStreams",
                                "ecr:BatchCheckLayerAvailability",
                                "ecr:BatchGetImage",
                                "ecr:GetDownloadUrlForLayer",
                                "ecr:GetAuthorizationToken",
                                "route53:*",
                                "elasticloadbalancing:DescribeLoadBalancers"
                            ],
                            "Resource": "*"
                        }]
                    }

    ECSInstanceProfile:
        Type: AWS::IAM::InstanceProfile
        Properties:
            Path: /
            Roles:
                - !Ref ECSRole

    EventLoggingLambda:
        Type: AWS::Lambda::Function
        Properties:
            Handler: index.handler
            Role: !GetAtt LambdaExecutionRole.Arn
            Runtime: nodejs4.3
            FunctionName: !Sub ${EnvironmentName}-ECS-Event-Logger
            Code:
                ZipFile: !Sub |
                  exports.handler = function(event, context) {
                      console.log(JSON.stringify(event, null, 2))
                  }

    LambdaExecutionRole:
        Type: AWS::IAM::Role
        Properties:
            AssumeRolePolicyDocument:
              Version: '2012-10-17'
              Statement:
              - Effect: Allow
                Principal:
                  Service:
                  - lambda.amazonaws.com
                Action:
                - sts:AssumeRole
            Path: "/"
            Policies:
            - PolicyName: root
              PolicyDocument:
                Version: '2012-10-17'
                Statement:
                - Effect: Allow
                  Action:
                  - logs:*
                  Resource: arn:aws:logs:*:*:*

Outputs:

    Cluster:
        Description: A reference to the ECS cluster
        Value: !Ref ECSCluster

    LogGroup:
        Description: A reference to the CloudWatch logs group
        Value: !Ref CloudWatchLogsGroup
`))
