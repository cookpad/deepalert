import * as cdk from '@aws-cdk/core';
import * as lambda from '@aws-cdk/aws-lambda';
import * as iam from '@aws-cdk/aws-iam';
import * as sns from '@aws-cdk/aws-sns';
import * as sqs from '@aws-cdk/aws-sqs';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as sfn from '@aws-cdk/aws-stepfunctions';
import * as tasks from '@aws-cdk/aws-stepfunctions-tasks';
import {
  SqsEventSource,
  DynamoEventSource,
} from '@aws-cdk/aws-lambda-event-sources';
import { SqsSubscription } from '@aws-cdk/aws-sns-subscriptions';

import * as path from 'path';
import * as fs from 'fs';
export interface Property extends cdk.StackProps {
  assetsPath?: string;

  lambdaRoleARN?: string;
  sfnRoleARN?: string;
  reviewer?: lambda.Function;
  inspectDelay?: cdk.Duration;
  reviewDelay?: cdk.Duration;

  sentryDsn?: string;
  sentryEnv?: string;
  logLevel?: string;
  alertTopicARN?: string;
}

export class DeepAlertStack extends cdk.Stack {
  readonly cacheTable: dynamodb.Table;
  // Messaging
  readonly taskTopic: sns.Topic;
  readonly attributeTopic: sns.Topic;
  readonly reportTopic: sns.Topic;
  readonly alertQueue: sqs.Queue;
  readonly findingQueue: sqs.Queue;
  readonly attributeQueue: sqs.Queue;
  readonly deadLetterQueue: sqs.Queue;

  // Lambda
  receptAlert: lambda.Function;
  dispatchInspection: lambda.Function;
  submitFinding: lambda.Function;
  feedbackAttribute: lambda.Function;
  compileReport: lambda.Function;
  dummyReviewer: lambda.Function;
  submitReport: lambda.Function;
  publishReport: lambda.Function;

  // StepFunctions
  readonly inspectionMachine: sfn.StateMachine;
  readonly reviewMachine: sfn.StateMachine;

  constructor(scope: cdk.Construct, id: string, props: Property) {
    super(scope, id, props);

    const lambdaRole = props.lambdaRoleARN
      ? iam.Role.fromRoleArn(this, "LambdaRole", props.lambdaRoleARN, {
        mutable: false,
      })
      : undefined;
    const sfnRole = props.sfnRoleARN
      ? iam.Role.fromRoleArn(this, "SfnRole", props.sfnRoleARN, {
        mutable: false,
      })
      : undefined;

    this.cacheTable = new dynamodb.Table(this, "cacheTable", {
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
      timeToLiveAttribute: "expires_at",
      stream: dynamodb.StreamViewType.NEW_IMAGE,
    });

    // ----------------------------------------------------------------
    // Messaging Channels
    this.taskTopic = new sns.Topic(this, "taskTopic");
    this.attributeTopic = new sns.Topic(this, "attributeTopic");
    this.reportTopic = new sns.Topic(this, "reportTopic");

    this.deadLetterQueue = new sqs.Queue(this, "deadLetterQueue");

    const alertQueueTimeout = cdk.Duration.seconds(30);
    this.alertQueue = new sqs.Queue(this, "alertQueue", {
      visibilityTimeout: alertQueueTimeout,
      deadLetterQueue: {
        maxReceiveCount: 5,
        queue: this.deadLetterQueue,
      },
    });

    if (props.alertTopicARN !== undefined) {
      const alertTopic = sns.Topic.fromTopicArn(this, 'AlertTopic', props.alertTopicARN);
      alertTopic.addSubscription(new SqsSubscription(this.alertQueue));
    }

    const findingQueueTimeout = cdk.Duration.seconds(30);
    this.findingQueue = new sqs.Queue(this, "findingQueue", {
      visibilityTimeout: findingQueueTimeout,
      deadLetterQueue: {
        maxReceiveCount: 5,
        queue: this.deadLetterQueue,
      },
    });

    const attributeQueueTimeout = cdk.Duration.seconds(30);
    this.attributeQueue = new sqs.Queue(this, "attributeQueue", {
      visibilityTimeout: attributeQueueTimeout,
      deadLetterQueue: {
        maxReceiveCount: 5,
        queue: this.deadLetterQueue,
      },
    });


    // ----------------------------------------------------------------
    // Lambda Functions
    const baseEnvVars = {
      TASK_TOPIC: this.taskTopic.topicArn,
      REPORT_TOPIC: this.reportTopic.topicArn,
      CACHE_TABLE: this.cacheTable.tableName,

      SENTRY_DSN: props.sentryDsn || "",
      SENTRY_ENVIRONMENT: props.sentryEnv || "",
      LOG_LEVEL: props.logLevel || "",
    };

    interface LambdaConfig {
      funcName: string;
      events?: lambda.IEventSource[];
      timeout?: cdk.Duration;
      environment?: { [key: string]: string; };
      setToStack: {
        (f: lambda.Function): void;
      };
    }

    const assetsPath = props.assetsPath ?? path.resolve(__dirname, '..', 'build');

    const buildLambdaFunction = (config: LambdaConfig) => {
      const assetPath = path.join(assetsPath, config.funcName);
      if (!fs.existsSync(assetPath)) {
        throw new Error(`Lambda asset not found for "${config.funcName}": ${assetPath} (run "make build" first)`);
      }
      const f = new lambda.Function(this, config.funcName, {
        runtime: lambda.Runtime.PROVIDED_AL2,
        handler: "bootstrap",
        code: lambda.Code.fromAsset(assetPath),
        role: lambdaRole,
        events: config.events,
        timeout: config.timeout,
        environment: config.environment || baseEnvVars,
        deadLetterQueue: this.deadLetterQueue,
      });
      config.setToStack(f);
    };

    // receptAlert is configured later because it requires StepFunctions
    // in environment variables.
    const lambdaConfigs: LambdaConfig[] = [
      {
        funcName: 'submitFinding',
        events: [new SqsEventSource(this.findingQueue)],
        setToStack: (f: lambda.Function) => { this.submitFinding = f; }
      },
      {
        funcName: 'feedbackAttribute',
        events: [new SqsEventSource(this.attributeQueue)],
        timeout: attributeQueueTimeout,
        setToStack: (f: lambda.Function) => { this.feedbackAttribute = f; }
      },
      {
        funcName: 'dispatchInspection',
        setToStack: (f: lambda.Function) => { this.dispatchInspection = f; }
      },
      {
        funcName: 'compileReport',
        setToStack: (f: lambda.Function) => { this.compileReport = f; },
      },
      {
        funcName: 'dummyReviewer',
        setToStack: (f: lambda.Function) => { this.dummyReviewer = f; },
      },
      {
        funcName: 'submitReport',
        setToStack: (f: lambda.Function) => { this.submitReport = f; },
      },
      {
        funcName: 'publishReport',
        events: [
          new DynamoEventSource(this.cacheTable, {
            startingPosition: lambda.StartingPosition.LATEST,
            batchSize: 1,
          }),
        ],
        setToStack: (f: lambda.Function) => { this.publishReport = f; },
      },
    ];

    lambdaConfigs.forEach(buildLambdaFunction);

    this.inspectionMachine = buildInspectionMachine(
      this, id,
      this.dispatchInspection,
      props.inspectDelay,
      sfnRole
    );

    this.reviewMachine = buildReviewMachine(
      this, id,
      this.compileReport,
      props.reviewer || this.dummyReviewer,
      this.submitReport,
      props.reviewDelay,
      sfnRole
    );

    const envVarsWithSF = Object.assign(baseEnvVars, {
      INSPECTOR_MACHINE: this.inspectionMachine.stateMachineArn,
      REVIEW_MACHINE: this.reviewMachine.stateMachineArn,
    });
    buildLambdaFunction({
      funcName: 'receptAlert',
      timeout: alertQueueTimeout,
      events: [new SqsEventSource(this.alertQueue)],
      environment: envVarsWithSF,
      setToStack: (f: lambda.Function) => { this.receptAlert = f; },
    })


    if (lambdaRole === undefined) {
      this.inspectionMachine.grantStartExecution(this.receptAlert);
      this.reviewMachine.grantStartExecution(this.receptAlert);
      this.taskTopic.grantPublish(this.dispatchInspection);
      this.reportTopic.grantPublish(this.publishReport);

      // DynamoDB
      this.cacheTable.grantReadWriteData(this.receptAlert);
      this.cacheTable.grantReadWriteData(this.dispatchInspection);
      this.cacheTable.grantReadWriteData(this.feedbackAttribute);
      this.cacheTable.grantReadWriteData(this.submitFinding);
      this.cacheTable.grantReadWriteData(this.compileReport);
      this.cacheTable.grantReadWriteData(this.submitReport);
      this.cacheTable.grantReadWriteData(this.publishReport);

    }
  }
}

function buildInspectionMachine(
  scope: cdk.Construct,
  stackID: string,
  dispatchInspection: lambda.Function,
  delay?: cdk.Duration,
  sfnRole?: iam.IRole
): sfn.StateMachine {
  const waitTime = delay || cdk.Duration.minutes(5);

  const wait = new sfn.Wait(scope, 'WaitDispatch', {
    time: sfn.WaitTime.duration(waitTime),
  });
  const invokeDispatcher = new tasks.LambdaInvoke(
    scope,
    'InvokeDispatchInspection',
    { lambdaFunction: dispatchInspection }
  );

  const definition = wait.next(invokeDispatcher);

  return new sfn.StateMachine(scope, 'InspectionMachine', {
    stateMachineName: stackID + '-InspectionMachine',
    definition,
    role: sfnRole,
  });
}

function buildReviewMachine(
  scope: cdk.Construct,
  stackID: string,
  compileReport: lambda.Function,
  reviewer: lambda.Function,
  submitReport: lambda.Function,
  delay?: cdk.Duration,
  sfnRole?: iam.IRole
): sfn.StateMachine {
  const waitTime = delay || cdk.Duration.minutes(10);

  const wait = new sfn.Wait(scope, 'WaitCompile', {
    time: sfn.WaitTime.duration(waitTime),
  });

  const definition = wait
    .next(
      new tasks.LambdaInvoke(scope, 'invokeCompileReport', {
        lambdaFunction: compileReport,
        outputPath: '$',
        payloadResponseOnly: true,
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, 'invokeReviewer', {
        lambdaFunction: reviewer,
        resultPath: '$.result',
        outputPath: '$',
        payloadResponseOnly: true,
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, 'invokeSubmitReport', {
        lambdaFunction: submitReport,
      })
    );

  return new sfn.StateMachine(scope, 'ReviewMachine', {
    stateMachineName: stackID + '-ReviewMachine',
    definition,
    role: sfnRole,
  });
}
