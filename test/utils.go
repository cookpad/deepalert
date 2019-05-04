package test

import (
	"encoding/json"
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
		switch resource.LogicalResourceId {
		case "TestInspector":
			conf.TestInspectorName = resource.PhysicalResourceId
		case "TestPublisher":
			conf.TestPublisherName = resource.PhysicalResourceId
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
