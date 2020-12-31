# DeepAlert

Serverless SOAR (Security Orchestration, Automation and Response) framework for automatic inspection and evaluation of security alert.

## Overview

DeepAlert receives a security alert that is event of interest from security view point and responses the alert automatically. DeepAlert has 3 parts of automatic response.

- **Inspector** investigates entities that are appeared in the alert including IP address, Domain name and store a result: reputation, history of malicious activities, associated cloud instance and etc. Following components are already provided to integrate with your DeepAlert environment. Also you can create own inspector to check logs that is stored into original log storage or log search system.
- **Reviewer** receives the alert with result(s) of Inspector and evaluate severity of the alert. Reviewer should be written by each security operator/administrator of your organization because security policies are differ from organization to organization.
- **Emitter** finally receives the alert with result of Reviewer's severity evaluation. After that, Emitter sends external integrated system. E.g. PagerDuty, Slack, Github Enterprise, etc. Also automatic quarantine can be configured by AWS Lambda function.

![Overview](https://user-images.githubusercontent.com/605953/76850323-80914100-688a-11ea-9c9a-96030094af2c.png)

## Deployment

### Prerequisite

- Tools
  - `aws-cdk` >= 1.75.0
  - `go` >= 1.14
  - `node` >= 14.7.0
  - `npm` >= 6.14.9
- Credential
  - AWS CLI credential to deploy CloudFormation. See [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) for more detail.

### Configure your stack

At first, you need to create AWS CDK repository and install deepalert as a npm module.

```bash
$ mkdir your-stack
$ cd your-stack
$ cdk init --language typescript
$ npm i @deepalert/deepalert
```

Then, edit `./bin/your-stack.ts` as following.

```ts
#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { DeepAlertStack } from '@deepalert/deepalert';

const app = new cdk.App();
new DeepAlertStack(app, 'YourDeepAlert', {});
```

### Deploy your stack

```bash
$ cdk deploy
```

## Alerting

### Alert data schema

```json
{
  "detector": "your-anti-virus",
  "rule_name": "detected malware",
  "rule_id": "detect-malware-by-av",
  "alert_key": "xxxxxxxx",
  "timestamp": "2006-01-02T15:03:04Z",
  "attributes": [
    {
      "type": "ipaddr",
      "key": "IP address of detected machine",
      "value": "10.2.3.4",
      "context": [
        "local",
        "client"
      ],
    },
  ]
}
```


- `detector`: Subject name of monitoring system
- `rule_id`: Machine readable rule identity
- `timestamp`: Detected timestamp
- `rule_name` (optional): Human readable rule name
- `alert_key` (optional): Alert aggregation key if you need
- `attributes` (optional): List of `attribute`
  - `type`: Choose from `ipaddr`, `domain`, `username`, `filehashvalue`, `json` and `url`
  - `key`: Label of the value
  - `value`: Actual value
  - `context`: One or multiple tags describe context of the attribute. See `AttrContext` in [alert.go](alert.go)

### Emit alert via API

`apikey.json` is created in CWD when running `cdk deploy` and it has `X-API-KEY` to access deepalert API.

```bash
$ export AWS_REGION=ap-northeast-1 # set your region
$ export API_KEY=`cat apikey.json  | jq '.["X-API-KEY"]' -r`
$ export API_ID=`aws cloudformation describe-stack-resources --stack-name YourDeepAlert | jq -r '.StackResources[] | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId'`
$ curl -X POST \
  -H "X-API-KEY: $API_KEY" \
  https://$API_ID.execute-api.$AWS_REGION.amazonaws.com/prod/api/v1/alert \
  -d '{
  "detector": "your-anti-virus",
  "rule_name": "detected malware",
  "rule_id": "detect-malware-by-av",
  "alert_key": "xxxxxxxx"
}'
```


### Emit alert via SQS

```bash
$ export QUEUE_URL=`aws cloudformation describe-stack-resources --stack-name YourDeepAlert | jq -r '.StackResources[] | select(.LogicalResourceId | startswith("alertQueue")) | .PhysicalResourceId'`
$ aws sqs send-message --queue-url $QUEUE_URL --message-body '{
  "detector": "your-anti-virus",
  "rule_name": "detected malware",
  "rule_id": "detect-malware-by-av",
  "alert_key": "xxxxxxxx"
}'
```


### Build and deploy Reviewer

See examples and deploy it as Lambda Function.

- Inspector example: [./examples/inspector](./examples/inspector)
- Emitter example: [./examples/emitter](./examples/inspector)

## Development

### Architecture

![architecture overview](https://user-images.githubusercontent.com/605953/103391370-8677ba80-4b5c-11eb-8b96-d44e1d3263a5.png)


### Unit Test

```
$ go test ./...
```

### Integration Test

Move to `./test/workflow/` and run below. Then deploy test stack and execute integration test.

```bash
$ npm i
$ make deploy
$ make test
```

## License

MIT License
