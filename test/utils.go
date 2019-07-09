package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TestConfig is settings prameters for tests of deepalert/ and deepalert/functions
type TestConfig struct {
	StackName string
	TableName string
	Region    string
	LogGroup  string
	LogStream string

	TestInspectorName string
	TestPublisherName string
	TestInspectorArn  string
	TestPublisherArn  string
}

type cfnStack struct {
	StackResources []struct {
		StackName          string
		StackId            string
		LogicalResourceId  string
		PhysicalResourceId string
	}
}

func loadStackOutput(conf *TestConfig, relPathToRoot string) {
	stackOutputPath := "output.json"
	if envvar := os.Getenv("DEEPALERT_STACK_OUTPUT"); envvar != "" {
		stackOutputPath = envvar
	}

	actualPath := filepath.Join(relPathToRoot, stackOutputPath)
	raw, err := ioutil.ReadFile(actualPath)
	if err != nil {
		log.Fatalf("Fail to read DEEPALERT_STACK_OUTPUT: %s, %s", actualPath, err)
	}

	var stack cfnStack
	if err := json.Unmarshal(raw, &stack); err != nil {
		log.Fatalf("Fail to unmarshal output file: %s, %s", actualPath, err)
	}

	for _, resource := range stack.StackResources {
		if conf.StackName == "" {
			conf.StackName = resource.StackName
		}

		switch resource.LogicalResourceId {
		case "CacheTable":
			conf.TableName = resource.PhysicalResourceId
			conf.Region = strings.Split(resource.StackId, ":")[3]
		case "LogStore":
			conf.LogGroup = resource.PhysicalResourceId
		case "LogStream":
			conf.LogStream = resource.PhysicalResourceId
		}
	}
}

// stackIDtoMeta return region and accountID
func stackIDtoMeta(stackID string) (string, string) {
	arn := strings.Split(stackID, ":")
	return arn[3], arn[4]
}

func loadTestStackOutput(conf *TestConfig, relPathToRoot string) {
	stackOutputPath := "test_output.json"
	if envvar := os.Getenv("DEEPALERT_TEST_STACK_OUTPUT"); envvar != "" {
		stackOutputPath = envvar
	}

	actualPath := filepath.Join(relPathToRoot, stackOutputPath)
	raw, err := ioutil.ReadFile(actualPath)
	if err != nil {
		log.Fatalf("Fail to read DEEPALERT_TEST_STACK_OUTPUT: %s, %s", actualPath, err)
	}

	var stack cfnStack
	if err := json.Unmarshal(raw, &stack); err != nil {
		log.Fatalf("Fail to unmarshal output file: %s, %s", actualPath, err)
	}

	for _, resource := range stack.StackResources {
		region, accountID := stackIDtoMeta(resource.StackId)

		switch resource.LogicalResourceId {
		case "TestInspector":
			conf.TestInspectorName = resource.PhysicalResourceId
			conf.TestInspectorArn = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
				region, accountID, resource.PhysicalResourceId)
		case "TestPublisher":
			conf.TestPublisherName = resource.PhysicalResourceId
			conf.TestPublisherArn = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
				region, accountID, resource.PhysicalResourceId)
		}
	}
}

// LoadTestConfig reads and parses config file.
func LoadTestConfig(relPathToRoot string) TestConfig {
	var conf TestConfig

	loadStackOutput(&conf, relPathToRoot)
	loadTestStackOutput(&conf, relPathToRoot)

	return conf
}
