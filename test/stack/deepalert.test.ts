import {
  expect as expectCDK,
  haveResource,
  haveResourceLike,
  countResources,
} from "@aws-cdk/assert";
import * as cdk from "@aws-cdk/core";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";
import * as Deepalert from "../../cdk/deepalert-stack";

const LAMBDA_FUNCTIONS = [
  "submitFinding",
  "feedbackAttribute",
  "dispatchInspection",
  "compileReport",
  "dummyReviewer",
  "submitReport",
  "publishReport",
  "receptAlert",
  "apiHandler",
];

// Create a temporary directory tree with stub bootstrap binaries so
// lambda.Code.fromAsset has real directories to fingerprint.
let assetsPath: string;
let apiKeyPath: string;

beforeAll(() => {
  assetsPath = fs.mkdtempSync(path.join(os.tmpdir(), "deepalert-test-"));
  for (const fn of LAMBDA_FUNCTIONS) {
    const dir = path.join(assetsPath, fn);
    fs.mkdirSync(dir);
    fs.writeFileSync(path.join(dir, "bootstrap"), "");
  }
  apiKeyPath = path.join(assetsPath, "apikey.json");
  fs.writeFileSync(apiKeyPath, JSON.stringify({ "X-API-KEY": "test-key" }));
});

afterAll(() => {
  fs.rmSync(assetsPath, { recursive: true, force: true });
});

function makeStack(props: Partial<Deepalert.Property> = {}): cdk.Stack {
  const app = new cdk.App();
  return new Deepalert.DeepAlertStack(app, "TestStack", {
    assetsPath,
    ...props,
  });
}

describe("DeepAlertStack", () => {
  describe("default stack (enableAPI: false)", () => {
    let stack: cdk.Stack;
    beforeAll(() => { stack = makeStack(); });

    test("creates all 8 core Lambda functions with PROVIDED_AL2 runtime", () => {
      expectCDK(stack).to(countResources("AWS::Lambda::Function", 8));
      expectCDK(stack).to(haveResourceLike("AWS::Lambda::Function", {
        Runtime: "provided.al2",
        Handler: "bootstrap",
      }));
    });

    test("DynamoDB table has TTL and NEW_IMAGE stream enabled", () => {
      expectCDK(stack).to(haveResource("AWS::DynamoDB::Table", {
        TimeToLiveSpecification: { AttributeName: "expires_at", Enabled: true },
        StreamSpecification: { StreamViewType: "NEW_IMAGE" },
        BillingMode: "PAY_PER_REQUEST",
      }));
    });

    test("creates 4 SQS queues (alert, finding, attribute, dead-letter)", () => {
      expectCDK(stack).to(countResources("AWS::SQS::Queue", 4));
    });

    test("functional queues are wired to the dead-letter queue", () => {
      expectCDK(stack).to(haveResourceLike("AWS::SQS::Queue", {
        RedrivePolicy: { maxReceiveCount: 5 },
      }));
    });

    test("creates inspection and review Step Functions state machines", () => {
      expectCDK(stack).to(countResources("AWS::StepFunctions::StateMachine", 2));
      expectCDK(stack).to(haveResourceLike("AWS::StepFunctions::StateMachine", {
        StateMachineName: "TestStack-InspectionMachine",
      }));
      expectCDK(stack).to(haveResourceLike("AWS::StepFunctions::StateMachine", {
        StateMachineName: "TestStack-ReviewMachine",
      }));
    });

    test("does not create API Gateway when enableAPI is false", () => {
      expectCDK(stack).to(countResources("AWS::ApiGateway::RestApi", 0));
    });
  });

  describe("stack with enableAPI: true", () => {
    let stack: cdk.Stack;
    beforeAll(() => {
      stack = makeStack({ enableAPI: true, apiKeyPath });
    });

    test("creates 9 Lambda functions including apiHandler", () => {
      expectCDK(stack).to(countResources("AWS::Lambda::Function", 9));
    });

    test("creates an API Gateway REST API", () => {
      expectCDK(stack).to(countResources("AWS::ApiGateway::RestApi", 1));
    });

    test("creates an API key and usage plan", () => {
      expectCDK(stack).to(countResources("AWS::ApiGateway::ApiKey", 1));
      expectCDK(stack).to(countResources("AWS::ApiGateway::UsagePlan", 1));
    });

    test("API Gateway has alert and report resource paths", () => {
      expectCDK(stack).to(haveResourceLike("AWS::ApiGateway::Resource", {
        PathPart: "alert",
      }));
      expectCDK(stack).to(haveResourceLike("AWS::ApiGateway::Resource", {
        PathPart: "report",
      }));
    });
  });

  describe("stack with alertTopicARN", () => {
    test("subscribes the alert queue to the provided SNS topic", () => {
      const stack = makeStack({
        alertTopicARN: "arn:aws:sns:us-east-1:123456789012:my-topic",
      });
      expectCDK(stack).to(haveResourceLike("AWS::SNS::Subscription", {
        Protocol: "sqs",
      }));
    });
  });

  describe("asset path validation", () => {
    test("throws a clear error when asset directory does not exist", () => {
      expect(() =>
        makeStack({ assetsPath: "/nonexistent/path/" })
      ).toThrow(/Lambda asset not found.*run "make build" first/);
    });
  });
});
