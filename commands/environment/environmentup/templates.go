package environmentup

var templates = map[string]string{
	"stack.yaml":           stackTemplate,
	"ecs-cluster.yaml":     clusterTemplate,
	"load-balancers.yaml":  albTemplate,
	"security-groups.yaml": securityGroupTemplate,
}

var stackTemplate = `
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

Resources:

    SecurityGroups:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./security-groups.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                VPC: !Ref VPC

    ALB:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./load-balancers.yaml
            Parameters:
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
                InstanceType: t2.large
                ClusterSize: 4
                VPC: !Ref VPC
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

var clusterTemplate = `
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
                    yum install -y aws-cfn-bootstrap

                    # export MYVAR=foobar
                    # echo 'OPTIONS="-e FOO=bar"' > /etc/sysconfig/docker

                    /opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration
                    /opt/aws/bin/cfn-signal -e $? --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSAutoScalingGroup

        Metadata:
            AWS::CloudFormation::Init:
                config:
                    commands:
                        01_add_instance_to_cluster:
                            command: !Sub echo ECS_CLUSTER=${ECSCluster} >> /etc/ecs/ecs.config
                    files:
                        "/etc/cfn/cfn-hup.conf":
                            mode: 000400
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
                                action=/opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration

                    services:
                        sysvinit:
                            cfn-hup:
                                enabled: true
                                ensureRunning: true
                                files:
                                    - /etc/cfn/cfn-hup.conf
                                    - /etc/cfn/hooks.d/cfn-auto-reloader.conf

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
                                "ecs:Submit*",
                                "logs:CreateLogStream",
                                "logs:PutLogEvents",
                                "ecr:BatchCheckLayerAvailability",
                                "ecr:BatchGetImage",
                                "ecr:GetDownloadUrlForLayer",
                                "ecr:GetAuthorizationToken"
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

var albTemplate = `
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

Resources:

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

var securityGroupTemplate = `
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
