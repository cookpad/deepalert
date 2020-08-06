import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as iam from "@aws-cdk/aws-iam";
import * as sns from "@aws-cdk/aws-sns";
import * as sqs from "@aws-cdk/aws-sqs";
import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as sfn from "@aws-cdk/aws-stepfunctions";
import * as tasks from "@aws-cdk/aws-stepfunctions-tasks";
import { SqsEventSource } from "@aws-cdk/aws-lambda-event-sources";
import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";
import { NodejsFunction } from "@aws-cdk/aws-lambda-nodejs";
import * as path from "path";

export interface property extends cdk.StackProps {
  lambdaRoleARN?: string;
  sfnRoleARN?: string;
  reviewer?: lambda.Function;
  inspectDelay?: cdk.Duration;
  reviewDelay?: cdk.Duration;
}

export class DeepAlertStack extends cdk.Stack {
  readonly cacheTable: dynamodb.Table;
  // Messaging
  readonly alertTopic: sns.Topic;
  readonly taskTopic: sns.Topic;
  readonly contentTopic: sns.Topic;
  readonly attributeTopic: sns.Topic;
  readonly reportTopic: sns.Topic;
  readonly alertQueue: sqs.Queue;
  readonly contentQueue: sqs.Queue;
  readonly attributeQueue: sqs.Queue;

  // Lambda
  readonly recvAlert: lambda.Function;
  readonly dispatchInspection: lambda.Function;
  readonly submitContent: lambda.Function;
  readonly feedbackAttribute: lambda.Function;
  readonly compileReport: lambda.Function;
  readonly dummyReviewer: lambda.Function;
  readonly publishReport: lambda.Function;
  readonly lambdaError: lambda.Function;
  readonly stepFunctionError: lambda.Function;

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
    });

    // ----------------------------------------------------------------
    // Messaging Channels
    this.alertTopic = new sns.Topic(this, "alertTopic");
    this.taskTopic = new sns.Topic(this, "taskTopic");
    this.contentTopic = new sns.Topic(this, "contentTopic");
    this.attributeTopic = new sns.Topic(this, "attributeTopic");
    this.reportTopic = new sns.Topic(this, "reportTopic");

    const alertQueueTimeout = cdk.Duration.seconds(30);
    this.alertQueue = new sqs.Queue(this, "alertQueue", {
      visibilityTimeout: alertQueueTimeout,
    });
    this.alertTopic.addSubscription(new SqsSubscription(this.alertQueue));

    const contentQueueTimeout = cdk.Duration.seconds(30);
    this.contentQueue = new sqs.Queue(this, "contentQueue", {
      visibilityTimeout: contentQueueTimeout,
    });

    const attributeQueueTimeout = cdk.Duration.seconds(30);
    this.attributeQueue = new sqs.Queue(this, "attributeQueue", {
      visibilityTimeout: attributeQueueTimeout,
    });

    // ----------------------------------------------------------------
    // Lambda Functions
    const baseEnvVars = {
      TASK_TOPIC: this.taskTopic.topicArn,
      REPORT_TOPIC: this.reportTopic.topicArn,
      CACHE_TABLE: this.cacheTable.tableName,
    };
    const nodeModulesLayer = new lambda.LayerVersion(this, "NodeModulesLayer", {
      code: lambda.AssetCode.fromAsset("????"),
      compatibleRuntimes: [lambda.Runtime.NODEJS_10_X],
    });
    this.recvAlert = new NodejsFunction(this, "recvAlert", {
      entry: path.join(__dirname, "lambda/recvAlert.js"),
      handler: "main",
      timeout: alertQueueTimeout,
      role: lambdaRole,
      events: [new SqsEventSource(this.alertQueue)],
      environment: Object.assign(baseEnvVars, {
        INSPECTOR_MACHINE: "",
        REVIEW_MACHINE: "",
      }),
    });

    this.submitContent = new NodejsFunction(this, "submitContent", {
      entry: path.join(__dirname, "lambda/submitContent.js"),
      handler: "main",
      role: lambdaRole,
      events: [new SqsEventSource(this.alertQueue)],
      environment: baseEnvVars,
    });

    this.feedbackAttribute = new NodejsFunction(this, "feedbackAttribute", {
      entry: path.join(__dirname, "lambda/feedbackAttribute.js"),
      handler: "main",
      timeout: attributeQueueTimeout,
      role: lambdaRole,
      events: [new SqsEventSource(this.attributeQueue)],
      environment: baseEnvVars,
    });

    this.dispatchInspection = new NodejsFunction(this, "dispatchInspection", {
      entry: path.join(__dirname, "lambda/dispatchInspection.js"),
      handler: "main",
      role: lambdaRole,
      environment: baseEnvVars,
    });
    this.compileReport = new NodejsFunction(this, "compileReport", {
      entry: path.join(__dirname, "lambda/compileReport.js"),
      handler: "main",
      role: lambdaRole,
      environment: baseEnvVars,
    });
    this.dummyReviewer = new NodejsFunction(this, "dummyReviewer", {
      entry: path.join(__dirname, "lambda/dummyReviewer.js"),
      handler: "main",
      role: lambdaRole,
    });
    this.publishReport = new NodejsFunction(this, "publishReport", {
      entry: path.join(__dirname, "lambda/publishReport.js"),
      handler: "main",
      role: lambdaRole,
      environment: baseEnvVars,
    });
    this.stepFunctionError = new NodejsFunction(this, "stepFunctionError", {
      entry: path.join(__dirname, "lambda/stepFunctionError.js"),
      handler: "main",
      role: lambdaRole,
      environment: baseEnvVars,
    });
    this.lambdaError = new NodejsFunction(this, "lambdaError", {
      entry: path.join(__dirname, "lambda/lambdaError.js"),
      handler: "main",
      role: lambdaRole,
      environment: baseEnvVars,
    });

    this.inspectionMachine = buildInspectionMachine(
      this,
      this.dispatchInspection,
      this.stepFunctionError,
      props.inspectDelay,
      sfnRole
    );

    this.inspectionMachine = buildReviewMachine(
      this,
      this.compileReport,
      props.reviewer || this.dummyReviewer,
      this.publishReport,
      this.stepFunctionError,
      props.reviewDelay,
      sfnRole
    );
  }
}

function buildInspectionMachine(
  scope: cdk.Construct,
  dispatchInspection: lambda.Function,
  errorHandler: lambda.Function,
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
  const invokeErrorHandler = new tasks.LambdaInvoke(
    scope,
    "InvokeErrorHandler",
    { lambdaFunction: errorHandler }
  );

  const definition = wait
    .next(invokeDispatcher)
    .next(
      new sfn.Choice(scope, "Job Complete?").when(
        sfn.Condition.stringEquals("$.status", "FAILED"),
        invokeErrorHandler
      )
    );

  return new sfn.StateMachine(scope, "InspectionMachine", {
    definition,
    role: sfnRole,
  });
}

function buildReviewMachine(
  scope: cdk.Construct,
  compileReport: lambda.Function,
  reviewer: lambda.Function,
  publishReport: lambda.Function,
  errorHandler: lambda.Function,
  delay?: cdk.Duration,
  sfnRole?: iam.IRole
): sfn.StateMachine {
  const waitTime = delay || cdk.Duration.minutes(10);

  const invokeErrorHandler = new tasks.LambdaInvoke(
    scope,
    "InvokeErrorHandler",
    { lambdaFunction: errorHandler }
  );
  const condFailed = sfn.Condition.stringEquals("$.status", "FAILED");
  const condSucceeded = sfn.Condition.stringEquals("$.status", "SUCCEEDED");

  const wait = new sfn.Wait(scope, "WaitCompile", {
    time: sfn.WaitTime.duration(waitTime),
  });

  const definition = wait
    .next(
      new sfn.Choice(scope, "Job Complete?")
        .when(condFailed, invokeErrorHandler)
        .when(
          condSucceeded,
          new tasks.LambdaInvoke(scope, "invokeCompileReport", {
            lambdaFunction: compileReport,
          })
        )
    )
    .next(
      new sfn.Choice(scope, "Job Complete?")
        .when(condFailed, invokeErrorHandler)
        .when(
          sfn.Condition.stringEquals("$.status", "SUCCEEDED"),
          new tasks.LambdaInvoke(scope, "invokeReviewer", {
            lambdaFunction: reviewer,
          })
        )
    )
    .next(
      new sfn.Choice(scope, "Job Complete?")
        .when(condFailed, invokeErrorHandler)
        .when(
          sfn.Condition.stringEquals("$.status", "SUCCEEDED"),
          new tasks.LambdaInvoke(scope, "invokePublishReport", {
            lambdaFunction: publishReport,
          })
        )
    );

  return new sfn.StateMachine(scope, "InspectionMachine", {
    definition,
    role: sfnRole,
  });
}
