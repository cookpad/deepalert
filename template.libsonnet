{
  build(
    ReviewerLambdaArn='',
    LambdaRoleArn='',
    StepFunctionRoleArn='',
    InspectionDelay=60,
    ReviewDelay=120,
    LogGroupNamePrefix='/DeepAlert/'
  ):: {
    local LambdaRole = (if LambdaRoleArn != '' then LambdaRoleArn else { 'Fn::GetAtt': 'LambdaRole.Arn' }),
    local StepFunctionRole = (if StepFunctionRoleArn != '' then StepFunctionRoleArn else { 'Fn::GetAtt': 'StepFunctionRole.Arn' }),
    local ReviewerLambda = (if ReviewerLambdaArn != '' then ReviewerLambdaArn else { 'Fn::GetAtt': 'DummyReviewer.Arn' }),

    local LambdaRoleTemplate = {
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
              PolicyName: 'DeepAlertLambda',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      'dynamodb:PutItem',
                      'dynamodb:DeleteItem',
                      'dynamodb:GetItem',
                      'dynamodb:Query',
                      'dynamodb:Scan',
                      'dynamodb:UpdateItem',
                    ],
                    Resource: [
                      { 'Fn::GetAtt': 'CacheTable.Arn' },
                      {
                        'Fn::Sub': [
                          '${TableArn}/index/*',
                          { TableArn: { 'Fn::GetAtt': 'CacheTable.Arn' } },
                        ],
                      },
                    ],
                  },
                  {
                    Effect: 'Allow',
                    Action: ['sns:Publish'],
                    Resource: [
                      { Ref: 'ReportNotification' },
                      { Ref: 'TaskNotification' },
                      { Ref: 'ErrorNotification' },
                      { Ref: 'NullNotification' },
                    ],
                  },
                  {
                    Effect: 'Allow',
                    Action: ['states:StartExecution'],
                    Resource: [
                      { 'Fn::Sub': 'arn:aws:states:${AWS::Region}:${AWS::AccountId}:stateMachine:${AWS::StackName}-delay-dispatcher' },
                      { 'Fn::Sub': 'arn:aws:states:${AWS::Region}:${AWS::AccountId}:stateMachine:${AWS::StackName}-review-invoker' },
                    ],
                  },
                ],
              },
            },
            {
              PolicyName: 'LogOutput',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      'logs:DescribeLogStreams',
                      'logs:PutLogEvents',
                    ],
                    Resource: [
                      { 'Fn::Sub': 'arn:aws:logs:${AWS::Region}:*:log-group:${LogStore}' },
                      { 'Fn::Sub': 'arn:aws:logs:${AWS::Region}:*:log-group:${LogStore}:*:*' },
                    ],
                  },
                ],
              },
            },
          ],
        },
      },
    },

    local StepFunctionRoleTemplate = {
      StepFunctionRole: {
        Type: 'AWS::IAM::Role',
        Properties: {
          AssumeRolePolicyDocument: {
            Version: '2012-10-17',
            Statement: [
              {
                Effect: 'Allow',
                Principal: {
                  Service: { 'Fn::Sub': 'states.${AWS::Region}.amazonaws.com' },
                },
                Action: [
                  'sts:AssumeRole',
                ],
              },
            ],
          },
          Path: '/',
          Policies: [
            {
              PolicyName: 'AlertResponderLambdaInvokeReviewer',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      'lambda:InvokeFunction',
                    ],
                    Resource: [
                      ReviewerLambda,
                      { 'Fn::GetAtt': 'DispatchInspection.Arn' },
                      { 'Fn::GetAtt': 'StepFunctionError.Arn' },
                      { 'Fn::GetAtt': 'CompileReport.Arn' },
                      { 'Fn::GetAtt': 'PublishReport.Arn' },
                    ],
                  },
                ],
              },
            },
          ],
        },
      },
    },

    local stateErrorCatch = [
      {
        ErrorEquals: ['States.ALL'],
        ResultPath: '$.error',
        Next: 'ErrorHandler',
      },
    ],

    local inspectorStateMachine = {
      StartAt: 'Wait1',
      States: {
        Wait1: {
          Type: 'Wait',
          Next: 'Inspection',
          Seconds: InspectionDelay,
        },
        Inspection: {
          Type: 'Task',
          Resource: '${dispatcherArn}',
          Catch: stateErrorCatch,
          End: true,
        },
        ErrorHandler: {
          Type: 'Task',
          Resource: '${errorHandlerArn}',
          End: true,
        },
      },
    },

    local reviewerStateMachine = {
      StartAt: 'Wait2',
      States: {
        Wait2: {
          Type: 'Wait',
          Next: 'Compile',
          Seconds: ReviewDelay,
        },
        Compile: {
          Type: 'Task',
          Resource: '${compilerArn}',
          Catch: stateErrorCatch,
          Next: 'Review',
        },
        Review: {
          Type: 'Task',
          Resource: ReviewerLambda,
          Catch: stateErrorCatch,
          ResultPath: '$.result',
          Next: 'Publish',
        },
        ErrorHandler: {
          Type: 'Task',
          Resource: '${errorHandlerArn}',
          End: true,
        },
        Publish: {
          Type: 'Task',
          Resource: '${publisherArn}',
          End: true,
        },
      },
    },

    local stateMachineMapping = {
      dispatcherArn: { 'Fn::GetAtt': 'DispatchInspection.Arn' },
      compilerArn: { 'Fn::GetAtt': 'CompileReport.Arn' },
      publisherArn: { 'Fn::GetAtt': 'PublishReport.Arn' },
      errorHandlerArn: { 'Fn::GetAtt': 'StepFunctionError.Arn' },
    },

    AWSTemplateFormatVersion: '2010-09-09',
    Description: 'DeepAlert https://github.com/m-mizutani/deepalert',
    Transform: 'AWS::Serverless-2016-10-31',
    Globals: {
      Function: {
        Runtime: 'go1.x',
        Timeout: 30,
        MemorySize: 128,
        ReservedConcurrentExecutions: 2,
        DeadLetterQueue: {
          Type: 'SNS',
          TargetArn: {
            Ref: 'ErrorNotification',
          },
        },
        Environment: {
          Variables: {
            TASK_NOTIFICATION: { Ref: 'TaskNotification' },
            REPORT_NOTIFICATION: { Ref: 'ReportNotification' },
            LOG_GROUP: { Ref: 'LogStore' },
            CACHE_TABLE: { Ref: 'CacheTable' },
          },
        },
      },
    },
    Resources: {
      // DynamoDB Tables
      CacheTable: {
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
          ProvisionedThroughput: {
            ReadCapacityUnits: 2,
            WriteCapacityUnits: 2,
          },
          TimeToLiveSpecification: {
            AttributeName: 'expires_at',
            Enabled: true,
          },
        },
      },

      // StepFunctions
      InspectionMachine: {
        Type: 'AWS::StepFunctions::StateMachine',
        Properties: {
          StateMachineName: { 'Fn::Sub': '${AWS::StackName}-inspection-machine' },
          RoleArn: StepFunctionRole,
          DefinitionString: { 'Fn::Sub': [std.toString(inspectorStateMachine), stateMachineMapping] },
        },
      },

      ReviewMachine: {
        Type: 'AWS::StepFunctions::StateMachine',
        Properties: {
          StateMachineName: { 'Fn::Sub': '${AWS::StackName}-review-machine' },
          RoleArn: StepFunctionRole,
          DefinitionString: { 'Fn::Sub': [std.toString(reviewerStateMachine), stateMachineMapping] },
        },
      },

      // Lambda Functions
      ReceptAlert: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'ReceptAlert',
          Role: LambdaRole,
          Environment: {
            Variables: {
              INSPECTOR_MACHINE: { Ref: 'InspectionMachine' },
              REVIEW_MACHINE: { Ref: 'ReviewMachine' },
            },
          },
          Events: {
            NotifyTopic: {
              Type: 'SNS',
              Properties: {
                Topic: { Ref: 'AlertNotification' },
              },
            },
          },
        },
      },
      DispatchInspection: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'DispatchInspection',
          Role: LambdaRole,
        },
      },
      SubmitContent: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'SubmitContent',
          Role: LambdaRole,
          Events: {
            ContentNotification: {
              Type: 'SNS',
              Properties: {
                Topic: { Ref: 'ContentNotification' },
              },
            },
          },
        },
      },
      FeedbackAttribute: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'FeedbackAttribute',
          Role: LambdaRole,
          Events: {
            AttributeNotification: {
              Type: 'SNS',
              Properties: {
                Topic: { Ref: 'AttributeNotification' },
              },
            },
          },
        },
      },
      CompileReport: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'CompileReport',
          Role: LambdaRole,
        },
      },
      PublishReport: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'PublishReport',
          Role: LambdaRole,
        },
      },
      StepFunctionError: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'StepFunctionError',
          Role: LambdaRole,
        },
      },
      DummyReviewer: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'DummyReviewer',
          Role: LambdaRole,
        },
      },
      ErrorHandler: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'ErrorHandler',
          Role: LambdaRole,
          DeadLetterQueue: {
            Type: 'SNS',
            TargetArn: { Ref: 'NullNotification' },
          },
          Events: {
            ErrorNotification: {
              Type: 'SNS',
              Properties: {
                Topic: { Ref: 'ErrorNotification' },
              },
            },
          },
        },
      },

      // SNS topics
      AlertNotification: {
        Type: 'AWS::SNS::Topic',
      },
      TaskNotification: {
        Type: 'AWS::SNS::Topic',
      },
      ContentNotification: {
        Type: 'AWS::SNS::Topic',
      },
      AttributeNotification: {
        Type: 'AWS::SNS::Topic',
      },
      ErrorNotification: {
        Type: 'AWS::SNS::Topic',
      },
      NullNotification: {
        Type: 'AWS::SNS::Topic',
      },
      ReportNotification: {
        Type: 'AWS::SNS::Topic',
      },

      // CloudWatch Logs LogGroup
      LogStore: {
        Type: 'AWS::Logs::LogGroup',
        Properties: {
          LogGroupName: { 'Fn::Sub': LogGroupNamePrefix + '${AWS::StackName}' },
        },
      },
    } + (
      if LambdaRoleArn != '' then {} else LambdaRoleTemplate
    ) + (
      if LambdaRoleArn != '' then {} else StepFunctionRoleTemplate
    ),

    // Output section
    Outputs: {
      AlertTopic: {
        Value: { Ref: 'AlertNotification' },
        Export: { Name: { 'Fn::Sub': '${AWS::StackName}-AlertTopic' } },
      },
      TaskTopic: {
        Value: { Ref: 'TaskNotification' },
        Export: { Name: { 'Fn::Sub': '${AWS::StackName}-TaskTopic' } },
      },
      ContentTopic: {
        Value: { Ref: 'ContentNotification' },
        Export: { Name: { 'Fn::Sub': '${AWS::StackName}-ContentTopic' } },
      },
      AttributeTopic: {
        Value: { Ref: 'AttributeNotification' },
        Export: { Name: { 'Fn::Sub': '${AWS::StackName}-AttributeTopic' } },
      },
      ReportTopic: {
        Value: { Ref: 'ReportNotification' },
        Export: { Name: { 'Fn::Sub': '${AWS::StackName}-ReportTopic' } },
      },
    },
  },
}
