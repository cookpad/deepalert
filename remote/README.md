# Integration Test

## Setup

Create a setting file as json like following and save as `config.json`.

```json
{
    "StackName": "your-test-stack-name",
    "CodeS3Bucket": "your-bucket",
    "CodeS3Prefix": "functions",
    "Region": "ap-northeast-1",
    "DeepAlertStackName": "your-deepalert-stack-name",
    "LambdaRoleArn": "arn:aws:iam::123456890xxx:role/YourLambdaRole"
}
```

**NOTE**

- `StackName` must be actual deployed CloudFormation stack name
- `LambdaRoleArn` requires permissions to access resources in your deepalert stack

Deploy test stack

```
$ make deploy
```
