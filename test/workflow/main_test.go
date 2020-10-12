package workflow_test

/*
var logger = logrus.New()

type testConfig struct {
	StackName          string
	Region             string
	DeepAlertStackName string
	ResultTableArn     string
}

func loadTestConfig(path string) *testConfig {
	var config testConfig
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		logger.WithError(err).Fatalf("Fail to read file data: %v", path)
		return nil
	}

	if err := json.Unmarshal(raw, &config); err != nil {
		logger.WithError(err).Fatalf("Fail to parse config file: %v", path)
	}

	return &config
}

var (
	testStackResources stackResources
	testCfg            *testConfig
)

func init() {
	if "" == os.Getenv("WORKFLOW_TEST") {
		return
	}

	// Setup generalprobe
	generalprobe.SetLoggerTraceLevel()

	// Setup logger
	logger.SetLevel(logrus.DebugLevel)
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-1"
	}

		deepalertStackName := os.Getenv("DEEPALERT_TEST_STACK_NAME")
		if deepalertStackName == "" {
			deepalertStackName = "DeepAlertTestStack"
		}


	workflowStackName := os.Getenv("DEEPALERT_WORKFLOW_STACK_NAME")
	if workflowStackName == "" {
		workflowStackName = "DeepAlertTestWorkflowStack"
	}

	testStack, err := getStackResources(region, workflowStackName)
	if err != nil {
		logger.WithError(err).Fatal("Fail setup")
	}

	testResultTable := testStack.lookup("ResultTable")
	arr := strings.Split(aws.StringValue(testResultTable.StackId), ":")
	testCfg.ResultTableArn = fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s",
		arr[3], arr[4], aws.StringValue(testResultTable.PhysicalResourceId))
}

type stackResources []*cloudformation.StackResource

func (x stackResources) lookup(logicID string) *cloudformation.StackResource {
	for i := 0; i < len(x); i++ {
		if aws.StringValue(x[i].LogicalResourceId) == logicID {
			return x[i]
		}
	}

	return nil
}

func getStackResources(region, stackName string) (stackResources, error) {
	cfn := cloudformation.New(session.New(), aws.NewConfig().WithRegion(region))

	req := cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	}
	resp, err := cfn.DescribeStackResources(&req)
	if err != nil {
		return nil, err
	}

	logger.WithField("resp", resp).Debug("result")

	return resp.StackResources, nil
}
*/
