package resources

import "text/template"

func init() {
	EnvironmentCloudFormationTemplates.Add(NewCloudFormationTemplate("dd-agent.yaml", environmentDataDogTemplate))
}

var environmentDataDogTemplate = template.Must(template.New("environmentDataDogTemplate").Parse(`
Parameters:
    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String
    LogGroupName:
        Description: The name of the cloudwatch log group to send container logs to
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
                        awslogs-group: !Ref LogGroupName
                        awslogs-region: !Ref AWS::Region
                        awslogs-stream-prefix: daemon-services/datadog
`))
