{
  build(ParentStackName):: {
    AWSTemplateFormatVersion: '2010-09-09',
    Description: 'TestStack for DeepAlert https://github.com/m-mizutani/deepalert',
    Transform: 'AWS::Serverless-2016-10-31',
    Globals: {
      Function: {
        Runtime: 'go1.x',
        Timeout: 30,
        MemorySize: 128,
        ReservedConcurrentExecutions: 1,
      },
    },
    Resources: {
      TestInspector: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'TestInspector',
          Role: {
            'Fn::GetAtt': 'LambdaRole.Arn',
          },
          Events: {
            NotifyTopic: {
              Type: 'SNS',
              Properties: {
                Topic: { 'Fn::ImportValue': ParentStackName + '-TaskTopic' },
              },
            },
          },
          Environment: {
            Variables: {
              SUBMIT_TOPIC: { 'Fn::ImportValue': ParentStackName + '-ContentTopic' },
              ATTRIBUTE_TOPIC: { 'Fn::ImportValue': ParentStackName + '-AttributeTopic' },
            },
          },
        },
      },
      TestPublisher: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'TestPublisher',
          Role: {
            'Fn::GetAtt': 'LambdaRole.Arn',
          },
          Events: {
            NotifyTopic: {
              Type: 'SNS',
              Properties: {
                Topic: { 'Fn::ImportValue': ParentStackName + '-ReportTopic' },
              },
            },
          },
        },
      },
      LambdaRole: {
        Type: 'AWS::IAM::Role',
        Properties: {
          AssumeRolePolicyDocument: {
            Version: '2012-10-17',
            Statement: [
              {
                Effect: 'Allow',
                Principal: {
                  Service: ['lambda.amazonaws.com'],
                },
                Action: ['sts:AssumeRole'],
              },
            ],
          },
          Path: '/',
          ManagedPolicyArns: [
            'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
          ],
          Policies: [
            {
              PolicyName: 'AlertResponderLambdaReviewer',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: ['sns:Publish'],
                    Resource: [
                      { 'Fn::ImportValue': ParentStackName + '-ContentTopic' },
                      { 'Fn::ImportValue': ParentStackName + '-AttributeTopic' },
                    ],
                  },
                ],
              },
            },
          ],
        },
      },
    },
  },
}
