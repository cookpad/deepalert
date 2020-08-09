#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as lambda from "@aws-cdk/aws-lambda";

import { DeepAlertStack } from "../../../lib/deepalert-stack";
import { SnsEventSource } from "@aws-cdk/aws-lambda-event-sources";

interface properties extends cdk.StackProps {
  readonly deepalert: DeepAlertStack;
}

export class WorkflowStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: properties) {
    super(scope, id, props);

    const table = new dynamodb.Table(this, "resultTable", {
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
    });

    const buildPath = lambda.Code.asset("./build");

    const testInspector = new lambda.Function(this, "testInspector", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "testInspector",
      timeout: cdk.Duration.seconds(30),
      code: buildPath,
      events: [new SnsEventSource(props.deepalert.taskTopic)],
      environment: {
        RESULT_TABLE: table.tableName,
        CONTENT_QUEUE: props.deepalert.contentQueue.queueUrl,
        ATTRIBUTE_QUEUE: props.deepalert.attributeQueue.queueUrl,
      },
    });

    const testEmitter = new lambda.Function(this, "testEmitter", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "testEmitter",
      timeout: cdk.Duration.seconds(30),
      code: buildPath,
      events: [new SnsEventSource(props.deepalert.reportTopic)],
      environment: {
        RESULT_TABLE: table.tableName,
      },
    });

    table.grantReadWriteData(testInspector);
    table.grantReadWriteData(testEmitter);
    props.deepalert.contentQueue.grantSendMessages(testInspector);
    props.deepalert.attributeQueue.grantSendMessages(testInspector);
  }
}

const deepalertStackName =
  process.env.DEEPALERT_TEST_STACK_NAME || "DeepAlertTestStack";
const workflowStackName =
  process.env.DEEPALERT_WORKFLOW_STACK_NAME || "DeepAlertTestWorkflowStack";

const app = new cdk.App();
const deepalert = new DeepAlertStack(app, deepalertStackName, {
  stackName: deepalertStackName,
});
new WorkflowStack(app, workflowStackName, {
  deepalert: deepalert,
  stackName: workflowStackName,
});
