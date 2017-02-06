package templates

var EnvironmentTemplates = map[string]string{
	"stack.yaml":           environmentStackTemplate,
	"ecs-cluster.yaml":     environmentClusterTemplate,
	"load-balancers.yaml":  environmentALBTemplate,
	"security-groups.yaml": environmentSecurityGroupTemplate,
	"dd-agent.yaml":        environmentDataDogTemplate,
}

var environmentStackTemplate = `
Description: >

    This template deploys a highly available ECS cluster using an AutoScaling Group, with
    ECS hosts distributed across multiple Availability Zones.

Parameters:

    VPC:
        Description: Choose which VPC this ECS cluster should be deployed to
        Type: AWS::EC2::VPC::Id

    InstanceSubnets:
        Description: Choose which subnets this ECS cluster instances should be deployed to
        Type: List<AWS::EC2::Subnet::Id>

    ALBSubnets:
        Description: Choose which subnets the application load balancer should be deployed to
        Type: List<AWS::EC2::Subnet::Id>

    ClusterSize:
        Description: Choose the number of container instances to add to the cluster
        Type: Number

    InstanceType:
        Description: Choose the type of EC2 instance to add to the cluster
        Type: String
        Default: t2.large

    DNSZone:
        Description: Select the DNS zone to use for service discovery
        Type: String

    DataDogAPIKey:
        Description: Please provide your datadog API key
        Type: String

Resources:

    SecurityGroups:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./security-groups.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                VPC: !Ref VPC

    DataDogTaskDefinition:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./dd-agent.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                DataDogAPIKey: !Ref DataDogAPIKey

    ALB:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./load-balancers.yaml
            Parameters:
                DNSZone: !Ref DNSZone
                EnvironmentName: !Ref AWS::StackName
                VPC: !Ref VPC
                Subnets: { "Fn::Join": [",", { "Ref": "ALBSubnets" } ] }
                SecurityGroup:
                  Fn::GetAtt:
                  - SecurityGroups
                  - Outputs.LoadBalancerSecurityGroup

    ECS:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./ecs-cluster.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                InstanceType: !Ref InstanceType
                ClusterSize: !Ref ClusterSize
                VPC: !Ref VPC
                DNSZone: !Ref DNSZone
                SecurityGroup:
                  Fn::GetAtt:
                  - SecurityGroups
                  - Outputs.ECSHostSecurityGroup
                Subnets: { "Fn::Join": [",", { "Ref": "InstanceSubnets" } ] }

Outputs:

    VPC:
        Description: The VPC ID
        Value: !Ref VPC

    Cluster:
        Description: A reference to the ECS cluster.
        Value:
          Fn::GetAtt:
            - ECS
            - Outputs.Cluster

    RecordSet:
        Description: A reference to the DNS recordset
        Value:
          Fn::GetAtt:
            - ALB
            - Outputs.RecordSet

    LoadBalancer:
        Description: A reference to the application load balancer.
        Value:
          Fn::GetAtt:
            - ALB
            - Outputs.LoadBalancer

    LoadBalancerUrl:
        Description: The URL of the ALB
        Value:
          Fn::GetAtt:
            - ALB
            - Outputs.LoadBalancerUrl

    Listener:
        Description: A reference to the port 80 listener
        Value:
          Fn::GetAtt:
            - ALB
            - Outputs.Listener
`

var environmentClusterTemplate = `
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

                install_dd_agent:
                    commands:
                        01_start_ecs:
                            command: "start ecs"
                        02_install_dd_agent:
                            command: "/usr/local/bin/install-dd-agent"

                    files:
                        "/usr/local/bin/install-dd-agent":
                            mode: "000755"
                            owner: root
                            group: root
                            content: !Sub |
                                #!/bin/bash
                                WAIT=0

                                while [ $WAIT -ne 10 ] && [ -z "$metadata" ]; do
                                    metadata=$(curl -s http://localhost:51678/v1/metadata)
                                    sleep $(( WAIT++ ))
                                done

                                echo "$metadata" > /etc/metadata.txt

                                instance_arn=$(cat /etc/metadata.txt | jq -r '. | .ContainerInstanceArn' | awk -F/ '{print $NF}' )

                                echo "aws ecs start-task --cluster ${EnvironmentName} --task-definition ${EnvironmentName}-datadog-agent --container-instances $instance_arn --region ${AWS::Region}" >> /etc/rc.local

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

Outputs:

    Cluster:
        Description: A reference to the ECS cluster
        Value: !Ref ECSCluster
`

var environmentALBTemplate = `
Description: >
    This template deploys an Application Load Balancer that exposes our various ECS services.
    We create them it a seperate nested template, so it can be referenced by all of the other nested templates.

Parameters:

    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    VPC:
        Type: AWS::EC2::VPC::Id
        Description: Choose which VPC the Applicaion Load Balancer should be deployed to

    Subnets:
        Description: Choose which subnets the Applicaion Load Balancer should be deployed to
        Type: List<AWS::EC2::Subnet::Id>

    SecurityGroup:
        Description: Select the Security Group to apply to the Applicaion Load Balancer
        Type: AWS::EC2::SecurityGroup::Id

    DNSZone:
        Description: Select the DNS zone the loadbalancer will be added to
        Type: String

Resources:
    RecordSet:
        Type: AWS::Route53::RecordSet
        Properties:
            Type: A
            HostedZoneName: !Sub ${DNSZone}.
            Name: !Sub ${EnvironmentName}.${DNSZone}
            AliasTarget:
                DNSName: !GetAtt LoadBalancer.DNSName
                HostedZoneId: !GetAtt LoadBalancer.CanonicalHostedZoneID

    LoadBalancer:
        Type: AWS::ElasticLoadBalancingV2::LoadBalancer
        Properties:
            Name: !Ref EnvironmentName
            Scheme: internal
            Subnets: !Ref Subnets
            SecurityGroups:
                - !Ref SecurityGroup
            Tags:
                - Key: Name
                  Value: !Ref EnvironmentName

    LoadBalancerListener:
        Type: AWS::ElasticLoadBalancingV2::Listener
        Properties:
            LoadBalancerArn: !Ref LoadBalancer
            Port: 80
            Protocol: HTTP
            DefaultActions:
                - Type: forward
                  TargetGroupArn: !Ref DefaultTargetGroup

    # We define a default target group here, as this is a mandatory Parameters
    # when creating an Application Load Balancer Listener. This is not used, instead
    # a target group is created per-service in each service template (../services/*)
    DefaultTargetGroup:
        Type: AWS::ElasticLoadBalancingV2::TargetGroup
        Properties:
            Name: !Sub ${EnvironmentName}-default
            VpcId: !Ref VPC
            Port: 80
            Protocol: HTTP

Outputs:
    RecordSet:
        Description: A reference to the DNS recordset
        Value: !Ref RecordSet

    LoadBalancer:
        Description: A reference to the Application Load Balancer
        Value: !Ref LoadBalancer

    LoadBalancerUrl:
        Description: The URL of the ALB
        Value: !Sub http://${LoadBalancer.DNSName}

    Listener:
        Description: A reference to a port 80 listener
        Value: !Ref LoadBalancerListener
`

var environmentSecurityGroupTemplate = `
Description: >
    This template contains the security groups required by our entire stack.
    We create them in a seperate nested template, so they can be referenced
    by all of the other nested templates.

Parameters:

    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    VPC:
        Type: AWS::EC2::VPC::Id
        Description: Choose which VPC the security groups should be deployed to

Resources:

    # This security group defines who/where is allowed to access the ECS hosts directly.
    # By default we're just allowing access from the load balancer.  If you want to SSH
    # into the hosts, or expose non-load balanced services you can open their ports here.
    ECSHostSecurityGroup:
        Type: AWS::EC2::SecurityGroup
        Properties:
            VpcId: !Ref VPC
            GroupDescription: Access to the ECS hosts and the tasks/containers that run on them
            SecurityGroupIngress:
                # Only allow inbound access to ECS from the ELB
                - SourceSecurityGroupId: !Ref LoadBalancerSecurityGroup
                  IpProtocol: -1
            Tags:
                - Key: Name
                  Value: !Sub ${EnvironmentName}-ECS-Hosts

    # This security group defines who/where is allowed to access the Application Load Balancer.
    # By default, we've opened this up to the public internet (0.0.0.0/0) but can you restrict
    # it further if you want.
    LoadBalancerSecurityGroup:
        Type: AWS::EC2::SecurityGroup
        Properties:
            VpcId: !Ref VPC
            GroupDescription: Access to the load balancer that sits in front of ECS
            SecurityGroupIngress:
                # Allow access from anywhere to our ECS services
                - CidrIp: 0.0.0.0/0
                  IpProtocol: -1
            Tags:
                - Key: Name
                  Value: !Sub ${EnvironmentName}-LoadBalancers

Outputs:

    ECSHostSecurityGroup:
        Description: A reference to the security group for ECS hosts
        Value: !Ref ECSHostSecurityGroup

    LoadBalancerSecurityGroup:
        Description: A reference to the security group for load balancers
        Value: !Ref LoadBalancerSecurityGroup
`

var environmentDataDogTemplate = `
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
`
