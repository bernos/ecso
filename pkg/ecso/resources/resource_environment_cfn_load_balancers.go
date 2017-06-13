package resources

import "text/template"

func init() {
	EnvironmentCloudFormationTemplates.Add(NewCloudFormationTemplate("load-balancers.yaml", environmentALBTemplate))
}

var environmentALBTemplate = template.Must(template.New("environmentALBTemplate").Parse(`
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
`))
