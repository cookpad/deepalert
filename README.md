# DeepAlert

Serverless SOAR (Security Orchestration, Automation and Response) framework for automatic inspection and evaluation of security alert.

## Overview

DeepAlert receives a security alert that is event of interest from security view point and responses the alert automatically. DeepAlert has 3 parts of automatic response.

- **Inspector** investigates entities that are appeaered in the alert including IP address, Domain name and store a result: reputation, history of malicious activities, associated cloud instance and etc. Following components are already provided to integrate with your DeepAlert environment. Also you can create own inspector to check logs that is stored into original log storage or log search system.
- **Reviewer** receives the alert with result(s) of Inspector and evaluate severity of the alert. Reviewer should be written by each security operator/administrator of your organization because security policies are differ from organazation to organization.
- **Emitter** finally receives the alert with result of Reviewer's severity evaluation. After that, Emitter sends external integrated system. E.g. PagerDuty, Slack, Github Enterprise, etc. Also automatic quarantine can be configured by AWS Lambda function.

![Overview](https://user-images.githubusercontent.com/605953/76850323-80914100-688a-11ea-9c9a-96030094af2c.png)

## How to use

### Prerequisite

- Tools
  - `awscli` >= 1.16.140
  - `go` >= 1.14
  - `GNU Make` >= 3.81
- Credential
  - AWS CLI credential to deploy CloudFormation. See [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) for more detail.

### Build and deploy Reviewer

See [example](./examples/reviewer) and deploy it as Lambda Function.

### Configuration and deploy DeepAlert

Clone this repository and create two config files, `deploy.jsonnet` and `stack.jsonnet` under `deepalert` directory.

```
$ git clone https://github.com/m-mizutani/deepalert.git
$ cd deepalert
```

**deploy.jsonnet** is for AWS SAM deployment by aws command.

```js
{
  StackName: 'deepalert',          // You can change the stack name.
  CodeS3Bucket: 'YOUR_S3_BUCKET',  // S3 bucket to save code materials for deployment
  CodeS3Prefix: 'functions',       // Prefix of S3 path if youn need (optional)
  Region: 'ap-northeast-1',        // Region to deploy the stack
}
```

**stack.jsonnet** is for building stack template.

```js
local template = import 'template.libsonnet';

// Set Lambda Function's ARN that you deployed as Reviewer
local reviewerArn = 'arn:aws:lambda:ap-northeast-1:123456789xx:function:YOUR_REVIEWER_ARN';

template.build(ReviewerLambdaArn=reviewerArn)
```

Then, deploy DeepAlert stack.

```
$ make deploy
```

### Build and deploy Reviewer

See examples and deploy it as Lambda Function.

- Inspector example: [./examples/inspector](./examples/inspector)
- Emitter example: [./examples/emitter](./examples/inspector)

## Development

### Architecture

![Architecture](https://user-images.githubusercontent.com/605953/76850184-34460100-688a-11ea-92fe-cd8a1226174f.png)

### Unit Test

```
$ make test
```

### Integration Test

At first, deploy DeepAlert stack for test. Then, create a config file into `remote/`.

```json
{
    "StackName": "deepalert-test-stack",
    "CodeS3Bucket": "YOUR_S3_BUCKET",
    "CodeS3Prefix": "SET_IF_YOU_NEED",
    "Region": "ap-northeast-1",
    "DeepAlertStackName": "DEPLOYED_DEEPALERT_STACK_NAME",
    "LambdaRoleArn": "arn:aws:iam::123456xxx:role/YOUR_LAMBDA_ROLE"
}
```

After that, deploy test stack and run test.

```
$ cd remote/
$ make test
```

## License

MIT License
