local config = import 'config.json';

{
  local TaskTopic = { 'Fn::ImportValue': config.DeepAlertStackName + '-TaskTopic' },
  local ContentQueue = { 'Fn::ImportValue': config.DeepAlertStackName + '-ContentQueue' },
  local AttributeQueue = { 'Fn::ImportValue': config.DeepAlertStackName + '-AttributeQueue' },
  local ReportTopic = { 'Fn::ImportValue': config.DeepAlertStackName + '-ReportTopic' },

  AWSTemplateFormatVersion: '2010-09-09',
  Description: 'DeepAlert TestStack https://github.com/m-mizutani/deepalert/test',
  Transform: 'AWS::Serverless-2016-10-31',

  Resources: {
    // DynamoDB Tables
    ResultTable: {
      Type: 'AWS::DynamoDB::Table',
      Properties: {
        AttributeDefinitions: [
          {
            AttributeName: 'pk',
            AttributeType: 'S',
          },
          {
            AttributeName: 'sk',
            AttributeType: 'S',
          },
        ],
        KeySchema: [
          {
            AttributeName: 'pk',
            KeyType: 'HASH',
          },
          {
            AttributeName: 'sk',
            KeyType: 'RANGE',
          },
        ],
        BillingMode: 'PAY_PER_REQUEST',
        TimeToLiveSpecification: {
          AttributeName: 'ttl',
          Enabled: true,
        },
      },
    },

    // Lambda Functions
    TestInspector: {
      Type: 'AWS::Serverless::Function',
      Properties: {
        CodeUri: 'build',
        Handler: 'TestInspector',
        Runtime: 'go1.x',
        Timeout: 30,
        MemorySize: 128,
        Role: config.LambdaRoleArn,
        Environment: {
          Variables: {
            RESULT_TABLE: { Ref: 'ResultTable' },
            CONTENT_QUEUE: ContentQueue,
            ATTRIBUTE_QUEUE: AttributeQueue,
          },
        },
        Events: {
          TaskTopic: {
            Type: 'SNS',
            Properties: { Topic: TaskTopic },
          },
        },
      },
    },

    TestEmitter: {
      Type: 'AWS::Serverless::Function',
      Properties: {
        CodeUri: 'build',
        Handler: 'TestEmitter',
        Runtime: 'go1.x',
        Timeout: 30,
        MemorySize: 128,
        Role: config.LambdaRoleArn,
        Environment: {
          Variables: {
            RESULT_TABLE: { Ref: 'ResultTable' },
          },
        },
        Events: {
          ReportTopic: {
            Type: 'SNS',
            Properties: { Topic: ReportTopic },
          },
        },
      },
    },
  },
}
