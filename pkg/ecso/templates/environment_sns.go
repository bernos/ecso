package templates

import "text/template"

var environmentSNSTemplate = template.Must(template.New("environmentSNSTemplate").Parse(`
Parameters:
    EnvironmentName:
        Description: An environment name that will be prefixed to resource names
        Type: String

    PagerDutyEndpoint:
        Description: The url of PagerDuty service endpoint to notify
        Type: String

Conditions:
    CreatePagerDutySubscription: {"Fn::Not": [{"Fn::Equals": ["", {"Ref":"PagerDutyEndpoint"}]}]}

Resources:
    AlertsTopic:
        Type: AWS::SNS::Topic
        Properties:
            TopicName: !Sub ${EnvironmentName}-Alerts
            DisplayName: !Sub Infrastructure alerts for ${EnvironmentName}
            Subscription: [{"Fn::If":["CreatePagerDutySubscription", {"Endpoint":{"Ref":"PagerDutyEndpoint"}, "Protocol":"https"},{"Ref":"AWS::NoValue"}]}]

Outputs:
    AlertsTopic:
        Description: A reference to the alerts SNS topic
        Value: !Ref AlertsTopic
`))
