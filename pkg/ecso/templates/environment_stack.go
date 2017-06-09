package templates

import "text/template"

var environmentStackTemplate = template.Must(template.New("environmentStackTemplate").Parse(`
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
        Default: ""

    KeyPair:
        Description: The keypair to add to each EC2 instance in the cluster
        Type: String

    PagerDutyEndpoint:
        Description: The url of PagerDuty service endpoint to notify
        Type: String
        Default: ""

Resources:
    Alarms:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./alarms.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                Cluster:
                    Fn::GetAtt:
                    - ECS
                    - Outputs.Cluster
                AlertsTopic:
                    Fn::GetAtt:
                    - SNS
                    - Outputs.AlertsTopic
                LoadBalancer:
                    Fn::GetAtt:
                    - ALB
                    - Outputs.LoadBalancer

    SNS:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./sns.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                PagerDutyEndpoint: !Ref PagerDutyEndpoint

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
                LogGroupName:
                    Fn::GetAtt:
                    - Logs
                    - Outputs.LogGroup

    DNSCleanerTaskDefinition:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./dns-cleaner.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                DNSZone: !Ref DNSZone
                LogGroupName:
                    Fn::GetAtt:
                    - Logs
                    - Outputs.LogGroup

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

    Logs:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./logging.yaml
            Parameters:
                LogGroupName: !Ref AWS::StackName

    ECS:
        Type: AWS::CloudFormation::Stack
        Properties:
            TemplateURL: ./ecs-cluster.yaml
            Parameters:
                EnvironmentName: !Ref AWS::StackName
                KeyPair: !Ref KeyPair
                InstanceType: !Ref InstanceType
                ClusterSize: !Ref ClusterSize
                VPC: !Ref VPC
                DNSZone: !Ref DNSZone
                DataDogAPIKey: !Ref DataDogAPIKey
                LogGroupName:
                    Fn::GetAtt:
                    - Logs
                    - Outputs.LogGroup
                NotificationsTopic:
                  Fn::GetAtt:
                  - SNS
                  - Outputs.NotificationsTopic
                SecurityGroup:
                  Fn::GetAtt:
                  - SecurityGroups
                  - Outputs.ECSHostSecurityGroup
                Subnets: { "Fn::Join": [",", { "Ref": "InstanceSubnets" } ] }

Outputs:

    AlertsTopic:
        Description: A reference to the alerts SNS topic
        Value:
          Fn::GetAtt:
            - SNS
            - Outputs.AlertsTopic

    NotificationsTopic:
        Description: A reference to the notifications SNS topic
        Value:
          Fn::GetAtt:
            - SNS
            - Outputs.NotificationsTopic

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

    LogGroup:
        Description: A reference to the CloudWatch logs group
        Value:
          Fn::GetAtt:
            - Logs
            - Outputs.LogGroup

    Listener:
        Description: A reference to the port 80 listener
        Value:
          Fn::GetAtt:
            - ALB
            - Outputs.Listener
`))
