package templates

// {
// "AWSTemplateFormatVersion": "2010-09-09",
// "Description": "SNS Topics to handle application notifications",
// "Parameters": {
// "TemplateBaseUrl": {
// "Type": "String"
// },
// "Version": {
// "Type": "String"
// },
// "ApplicationName": {
// "Type": "String",
// "Description": "The name of the application that the topic belongs to."
// },
// "Environment": {
// "Type": "String",
// "Description": "The name of the environment we deploying to."
// },
// "PagerDutyEndpoint": {
// "Type": "String",
// "Description": "The name of the pager duty endpoint."
// },
// "LambdaARN": {
// "Type": "String",
// "Description": "The arn of the lambda endpoint to push to slack."
// }
// },
// "Conditions": {
// "CreatePagerDutySubscription": {
// "Fn::Not": [
// {
// "Fn::Equals": [
// "",
// {
// "Ref": "PagerDutyEndpoint"
// }
// ]
// }
// ]
// },
// "CreateLambdaSubscription": {
// "Fn::Not": [
// {
// "Fn::Equals": [
// "",
// {
// "Ref": "LambdaARN"
// }
// ]
// }
// ]
// }
// },
// "Resources": {
// "NotificationsTopic": {
// "Type": "AWS::SNS::Topic",
// "Properties": {
// "DisplayName": {
// "Fn::Join": [
// " ",
// [
// {
// "Ref": "ApplicationName"
// },
// {
// "Ref": "Environment"
// },
// " notifications sns topic"
// ]
// ]
// },
// "Subscription": [
// {
// "Fn::If": [
// "CreateLambdaSubscription",
// {
// "Endpoint": {
// "Ref": "LambdaARN"
// },
// "Protocol": "lambda"
// },
// {
// "Ref": "AWS::NoValue"
// }
// ]
// }
// ],
// "TopicName": {
// "Fn::Join": [
// "-",
// [
// {
// "Ref": "ApplicationName"
// },
// {
// "Ref": "Environment"
// },
// "notifications"
// ]
// ]
// }
// }
// },
// "PagerDutyTopic": {
// "Type": "AWS::SNS::Topic",
// "Properties": {
// "DisplayName": {
// "Fn::Join": [
// " ",
// [
// {
// "Ref": "ApplicationName"
// },
// {
// "Ref": "Environment"
// },
// " pager duty sns topic"
// ]
// ]
// },
// "Subscription": [
// {
// "Fn::If": [
// "CreatePagerDutySubscription",
// {
// "Endpoint": {
// "Ref": "PagerDutyEndpoint"
// },
// "Protocol": "https"
// },
// {
// "Ref": "AWS::NoValue"
// }
// ]
// }
// ],
// "TopicName": {
// "Fn::Join": [
// "-",
// [
// {
// "Ref": "ApplicationName"
// },
// {
// "Ref": "Environment"
// },
// "pagerduty"
// ]
// ]
// }
// }
// }
// },
// "Outputs": {
// "NotificationsTopic": {
// "Value": {
// "Ref": "NotificationsTopic"
// }
// },
// "PagerDutyTopic": {
// "Value": {
// "Ref": "PagerDutyTopic"
// }
// }
// }
// }
