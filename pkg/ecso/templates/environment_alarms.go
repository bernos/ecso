package templates

// Resources:
//   ELBHealthyHostMinCountAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Min instance count" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Min instance count" ]]}
//       Namespace: AWS/ELB
//       MetricName: HealthyHostCount
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Average
//       Period: 120
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmMinHostCount
//           DefaultValue: 2
//       ComparisonOperator: LessThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   ELBHealthyHostMaxCountAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Max instance count" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Max instance count" ]]}
//       Namespace: AWS/ELB
//       MetricName: HealthyHostCount
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Average
//       Period: 120
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmMaxHostCount
//           DefaultValue: 10
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   ELBLatencyAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "ELB Latency." ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "ELB Latency." ]]}
//       Namespace: AWS/ELB
//       MetricName: Latency
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Average
//       Period: 300
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmLatency
//           DefaultValue: "0.8"
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   ELBHTTP5xxAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 5xx rate" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 5xx rate" ]]}
//       Namespace: AWS/ELB
//       MetricName: HTTPCode_Backend_5XX
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Sum
//       Period: 300
//       EvaluationPeriods: 1
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmHTTP5xxErrors
//           DefaultValue: "10"
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   ELBHTTP4xxAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 4xx rate" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 4xx rate" ]]}
//       Namespace: AWS/ELB
//       MetricName: HTTPCode_Backend_4XX
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Sum
//       Period: 60
//       EvaluationPeriods: 1
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmHTTP4xxErrors
//           DefaultValue: "10"
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   ELBHTTP2xxAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 2xx rate" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Backend HTTP 2xx rate" ]]}
//       Namespace: AWS/ELB
//       MetricName: HTTPCode_Backend_2XX
//       Dimensions:
//         - Name: LoadBalancerName
//           Value: { "Ref" : "AWSEBLoadBalancer" }
//       Statistic: Sum
//       Period: 60
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmHTTP2xxCount
//           DefaultValue: "5"
//       ComparisonOperator: LessThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   HighCPUAlarm:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "CPU usage" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "CPU usage" ]]}
//       Namespace: AWS/EC2
//       MetricName: CPUUtilization
//       Dimensions:
//         - Name: AutoScalingGroupName
//           Value: { "Ref" : "AWSEBAutoScalingGroup" }
//       Statistic: Average
//       Period: 300
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmCPU
//           DefaultValue: "80"
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }

//   DiskUsage:
//     Type: AWS::CloudWatch::Alarm
//     Properties:
//       AlarmName: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Disk usage" ]]}
//       AlarmDescription: { "Fn::Join" : [" ", [{ "Fn::GetOptionSetting": {"OptionName" : "AlarmPrefix", "DefaultValue" : "Unnamed application" } }, {"Ref":"AWSEBEnvironmentName"}, "Disk usage" ]]}
//       Namespace: AWS/ElasticBeanstalk
//       MetricName: RootFilesystemUtil
//       Dimensions:
//         - Name: EnvironmentName
//           Value: { "Ref" : "AWSEBEnvironmentName" }
//       Statistic: Maximum
//       Period: 300
//       EvaluationPeriods: 2
//       Threshold:
//         Fn::GetOptionSetting:
//           OptionName: AlarmDiskUsage
//           DefaultValue: "60"
//       ComparisonOperator: GreaterThanThreshold
//       ActionsEnabled: { "Fn::GetOptionSetting" : { "OptionName" : "AlarmActionsEnabled", "DefaultValue" : "false" } }
//       AlarmActions:
//         - { "Fn::GetOptionSetting": {"OptionName" : "NotificationsARN", "DefaultValue" : "Unnamed notification ARN" } }
//         - { "Fn::GetOptionSetting": {"OptionName" : "PagerDutyARN", "DefaultValue" : "Unnamed notification ARN" } }
