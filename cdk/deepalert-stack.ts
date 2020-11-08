import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as iam from "@aws-cdk/aws-iam";
import * as sns from "@aws-cdk/aws-sns";
import * as sqs from "@aws-cdk/aws-sqs";
import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as sfn from "@aws-cdk/aws-stepfunctions";
import * as tasks from "@aws-cdk/aws-stepfunctions-tasks";
import * as apigateway from "@aws-cdk/aws-apigateway";
import {
  SqsEventSource,
  DynamoEventSource,
} from "@aws-cdk/aws-lambda-event-sources";
import { SqsSubscription } from '@aws-cdk/aws-sns-subscriptions';

import * as path from "path";
import * as fs from "fs";

// import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";

export interface property extends cdk.StackProps {
  lambdaRoleARN?: string;
  sfnRoleARN?: string;
  reviewer?: lambda.Function;
  inspectDelay?: cdk.Duration;
  reviewDelay?: cdk.Duration;

  apiKeyPath?: string;
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
  readonly noteQueue: sqs.Queue;
  readonly attributeQueue: sqs.Queue;
  readonly deadLetterQueue: sqs.Queue;

  // Lambda
  readonly receptAlert: lambda.Function;
  readonly dispatchInspection: lambda.Function;
  readonly submitNote: lambda.Function;
  readonly feedbackAttribute: lambda.Function;
  readonly compileReport: lambda.Function;
  readonly dummyReviewer: lambda.Function;
  readonly submitReport: lambda.Function;
  readonly publishReport: lambda.Function;
  readonly lambdaError: lambda.Function;
  readonly apiHandler: lambda.Function;

  // StepFunctions
  readonly inspectionMachine: sfn.StateMachine;
  readonly reviewMachine: sfn.StateMachine;

  constructor(scope: cdk.Construct, id: string, props: property) {
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

    const alertQueueTimeout = cdk.Duration.seconds(30);

    this.alertQueue = new sqs.Queue(this, "alertQueue", {
      visibilityTimeout: alertQueueTimeout,
    });

    if (props.alertTopicARN !== undefined) {
      const alertTopic = sns.Topic.fromTopicArn(this, 'AlertTopic', props.alertTopicARN);
      alertTopic.addSubscription(new SqsSubscription(this.alertQueue));
    }

    const noteQueueTimeout = cdk.Duration.seconds(30);
    this.noteQueue = new sqs.Queue(this, "noteQueue", {
      visibilityTimeout: noteQueueTimeout,
    });

    const attributeQueueTimeout = cdk.Duration.seconds(30);
    this.attributeQueue = new sqs.Queue(this, "attributeQueue", {
      visibilityTimeout: attributeQueueTimeout,
    });

    this.deadLetterQueue = new sqs.Queue(this, "deadLetterQueue");

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
    const buildPath = lambda.Code.fromAsset(path.join(__dirname, "../build"));

    this.submitNote = new lambda.Function(this, "submitNote", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "submitNote",
      code: buildPath,
      role: lambdaRole,
      events: [new SqsEventSource(this.noteQueue)],
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
    });

    this.feedbackAttribute = new lambda.Function(this, "feedbackAttribute", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "feedbackAttribute",
      code: buildPath,
      timeout: attributeQueueTimeout,
      role: lambdaRole,
      events: [new SqsEventSource(this.attributeQueue)],
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
    });

    this.dispatchInspection = new lambda.Function(this, "dispatchInspection", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "dispatchInspection",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
    });
    this.compileReport = new lambda.Function(this, "compileReport", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "compileReport",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
    });
    this.dummyReviewer = new lambda.Function(this, "dummyReviewer", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "dummyReviewer",
      code: buildPath,
      role: lambdaRole,
      deadLetterQueue: this.deadLetterQueue,
    });
    this.submitReport = new lambda.Function(this, "submitReport", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "submitReport",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
    });
    this.publishReport = new lambda.Function(this, "publishReport", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "publishReport",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
      events: [
        new DynamoEventSource(this.cacheTable, {
          startingPosition: lambda.StartingPosition.LATEST,
          batchSize: 1,
        }),
      ],
    });
    this.lambdaError = new lambda.Function(this, "lambdaError", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "lambdaError",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      events: [new SqsEventSource(this.deadLetterQueue)],
    });

    this.inspectionMachine = buildInspectionMachine(
      this,
      this.dispatchInspection,
      props.inspectDelay,
      sfnRole
    );

    this.reviewMachine = buildReviewMachine(
      this,
      this.compileReport,
      props.reviewer || this.dummyReviewer,
      this.submitReport,
      props.reviewDelay,
      sfnRole
    );

    this.receptAlert = new lambda.Function(this, "receptAlert", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "receptAlert",
      code: buildPath,
      timeout: alertQueueTimeout,
      role: lambdaRole,
      events: [new SqsEventSource(this.alertQueue)],
      environment: Object.assign(baseEnvVars, {
        INSPECTOR_MACHINE: this.inspectionMachine.stateMachineArn,
        REVIEW_MACHINE: this.reviewMachine.stateMachineArn,
      }),
      deadLetterQueue: this.deadLetterQueue,
    });

    // API handler
    this.apiHandler = new lambda.Function(this, "apiHandler", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "apiHandler",
      code: buildPath,
      role: lambdaRole,
      timeout: cdk.Duration.seconds(120),
      memorySize: 2048,
      environment: baseEnvVars,
    });

    const api = new apigateway.LambdaRestApi(this, "deepalertAPI", {
      handler: this.apiHandler,
      proxy: false,
      cloudWatchRole: false,
      endpointTypes: [apigateway.EndpointType.REGIONAL],
      policy: new iam.PolicyDocument({
        statements: [
          new iam.PolicyStatement({
            actions: ["execute-api:Invoke"],
            resources: ["execute-api:/*/*"],
            effect: iam.Effect.ALLOW,
            principals: [new iam.AnyPrincipal()],
          }),
        ],
      }),
    });
    const apiKey = api.addApiKey('APIKey', {
      value: getAPIKey(props.apiKeyPath),
    })
    api.addUsagePlan('UsagePlan', {
      apiKey: apiKey,
    }).addApiStage({
      stage: api.deploymentStage,
    })

    const apiOpt = { apiKeyRequired: true};
    const v1 = api.root.addResource("api").addResource("v1",);
    const alertAPI = v1.addResource("alert");
    alertAPI.addMethod("POST", undefined, apiOpt);
    alertAPI.addResource("{alert_id}").addResource("report").addMethod("GET", undefined, apiOpt);

    const reportAPI = v1.addResource("report");
    const reportAPIwithID = reportAPI.addResource("{report_id}");
    reportAPIwithID.addMethod("GET", undefined, apiOpt);
    reportAPIwithID.addResource("alert").addMethod("GET", undefined, apiOpt);
    reportAPIwithID.addResource("attribute").addMethod("GET", undefined, apiOpt);
    reportAPIwithID.addResource("section").addMethod("GET", undefined, apiOpt);

    if (lambdaRole === undefined) {
      this.inspectionMachine.grantStartExecution(this.receptAlert);
      this.inspectionMachine.grantStartExecution(this.apiHandler);
      this.reviewMachine.grantStartExecution(this.receptAlert);
      this.reviewMachine.grantStartExecution(this.apiHandler);
      this.taskTopic.grantPublish(this.dispatchInspection);
      this.reportTopic.grantPublish(this.publishReport);

      // DynamoDB
      this.cacheTable.grantReadWriteData(this.receptAlert);
      this.cacheTable.grantReadWriteData(this.dispatchInspection);
      this.cacheTable.grantReadWriteData(this.feedbackAttribute);
      this.cacheTable.grantReadWriteData(this.submitNote);
      this.cacheTable.grantReadWriteData(this.compileReport);
      this.cacheTable.grantReadWriteData(this.submitReport);
      this.cacheTable.grantReadWriteData(this.publishReport);
      this.cacheTable.grantReadWriteData(this.apiHandler);
    }
  }
}

function buildInspectionMachine(
  scope: cdk.Construct,
  dispatchInspection: lambda.Function,
  delay?: cdk.Duration,
  sfnRole?: iam.IRole
): sfn.StateMachine {
  const waitTime = delay || cdk.Duration.minutes(5);

  const wait = new sfn.Wait(scope, "WaitDispatch", {
    time: sfn.WaitTime.duration(waitTime),
  });
  const invokeDispatcher = new tasks.LambdaInvoke(
    scope,
    "InvokeDispatchInspection",
    { lambdaFunction: dispatchInspection }
  );

  const definition = wait.next(invokeDispatcher);

  return new sfn.StateMachine(scope, "InspectionMachine", {
    definition,
    role: sfnRole,
  });
}

function buildReviewMachine(
  scope: cdk.Construct,
  compileReport: lambda.Function,
  reviewer: lambda.Function,
  submitReport: lambda.Function,
  delay?: cdk.Duration,
  sfnRole?: iam.IRole
): sfn.StateMachine {
  const waitTime = delay || cdk.Duration.minutes(10);

  const wait = new sfn.Wait(scope, "WaitCompile", {
    time: sfn.WaitTime.duration(waitTime),
  });

  const definition = wait
    .next(
      new tasks.LambdaInvoke(scope, "invokeCompileReport", {
        lambdaFunction: compileReport,
        outputPath: "$",
        payloadResponseOnly: true,
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, "invokeReviewer", {
        lambdaFunction: reviewer,
        resultPath: "$.result",
        outputPath: "$",
        payloadResponseOnly: true,
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, "invokeSubmitReport", {
        lambdaFunction: submitReport,
      })
    );

  return new sfn.StateMachine(scope, "ReviewMachine", {
    definition,
    role: sfnRole,
  });
}

function getAPIKey(apiKeyPath?: string): string {
  if (apiKeyPath === undefined) {
    apiKeyPath = path.join(__dirname, "apikey.json");
  }

  if (fs.existsSync(apiKeyPath)) {
    console.log("Read API key from: ", apiKeyPath);
    const buf = fs.readFileSync(apiKeyPath)
    const keyData = JSON.parse(buf.toString());
    return keyData['X-API-KEY'];
  } else {
    const literals = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    const length = 32;
    const apiKey = Array.from(Array(length)).map(()=>literals[Math.floor(Math.random()*literals.length)]).join('');
    fs.writeFileSync(apiKeyPath, JSON.stringify({'X-API-KEY': apiKey}))
    console.log("Generatedd and wrote API key to: ", apiKeyPath);
    return apiKey;
  }
}
