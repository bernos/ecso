package resources

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
