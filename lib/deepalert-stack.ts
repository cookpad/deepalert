import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as iam from "@aws-cdk/aws-iam";
import * as sns from "@aws-cdk/aws-sns";
import * as sqs from "@aws-cdk/aws-sqs";
import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as sfn from "@aws-cdk/aws-stepfunctions";
import * as tasks from "@aws-cdk/aws-stepfunctions-tasks";
import {
  SqsEventSource,
  SnsEventSource,
} from "@aws-cdk/aws-lambda-event-sources";
import * as path from "path";

// import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";

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
  // readonly alertQueue: sqs.Queue;
  readonly contentQueue: sqs.Queue;
  readonly attributeQueue: sqs.Queue;
  readonly deadLetterQueue: sqs.Queue;

  // Lambda
  readonly recvAlert: lambda.Function;
  readonly dispatchInspection: lambda.Function;
  readonly submitContent: lambda.Function;
  readonly feedbackAttribute: lambda.Function;
  readonly compileReport: lambda.Function;
  readonly dummyReviewer: lambda.Function;
  readonly publishReport: lambda.Function;
  readonly lambdaError: lambda.Function;

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
    /*
    this.alertQueue = new sqs.Queue(this, "alertQueue", {
      visibilityTimeout: alertQueueTimeout,
    });
    this.alertTopic.addSubscription(new SqsSubscription(this.alertQueue));
*/

    const contentQueueTimeout = cdk.Duration.seconds(30);
    this.contentQueue = new sqs.Queue(this, "contentQueue", {
      visibilityTimeout: contentQueueTimeout,
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
    };
    const buildPath = lambda.Code.asset(path.join(__dirname, "../build"));

    this.submitContent = new lambda.Function(this, "submitContent", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "submitContent",
      code: buildPath,
      role: lambdaRole,
      events: [new SqsEventSource(this.contentQueue)],
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
    this.publishReport = new lambda.Function(this, "publishReport", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "publishReport",
      code: buildPath,
      role: lambdaRole,
      environment: baseEnvVars,
      deadLetterQueue: this.deadLetterQueue,
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
      this.publishReport,
      props.reviewDelay,
      sfnRole
    );

    this.recvAlert = new lambda.Function(this, "recvAlert", {
      runtime: lambda.Runtime.GO_1_X,
      handler: "recvAlert",
      code: buildPath,
      timeout: alertQueueTimeout,
      role: lambdaRole,
      events: [new SnsEventSource(this.alertTopic)],
      environment: Object.assign(baseEnvVars, {
        INSPECTOR_MACHINE: this.inspectionMachine.stateMachineArn,
        REVIEW_MACHINE: this.reviewMachine.stateMachineArn,
      }),
      deadLetterQueue: this.deadLetterQueue,
    });

    if (lambdaRole === undefined) {
      this.inspectionMachine.grantStartExecution(this.recvAlert);
      this.reviewMachine.grantStartExecution(this.recvAlert);
      this.taskTopic.grantPublish(this.dispatchInspection);

      // DynamoDB
      this.cacheTable.grantReadWriteData(this.recvAlert);
      this.cacheTable.grantReadWriteData(this.dispatchInspection);
      this.cacheTable.grantReadWriteData(this.feedbackAttribute);
      this.cacheTable.grantReadWriteData(this.submitContent);
      this.cacheTable.grantReadWriteData(this.compileReport);
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
  publishReport: lambda.Function,
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
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, "invokeReviewer", {
        lambdaFunction: reviewer,
        resultPath: "$.result",
      })
    )
    .next(
      new tasks.LambdaInvoke(scope, "invokePublishReport", {
        lambdaFunction: publishReport,
      })
    );

  return new sfn.StateMachine(scope, "ReviewMachine", {
    definition,
    role: sfnRole,
  });
}
