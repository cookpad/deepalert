DEPLOY_CONFIG ?= deploy.jsonnet
STACK_CONFIG ?= stack.jsonnet
TEST_STACK_CONFIG ?= test.jsonnet
CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

TEMPLATE_JSONNET=$(CODE_DIR)/template.libsonnet
TEMPLATE_FILE=template.json
SAM_FILE=sam.yml
OUTPUT_FILE=$(CWD)/output.json

TEST_TEMPLATE_JSONNET=teststack.libsonnet
TEST_TEMPLATE_FILE=teststack_template.json
TEST_SAM_FILE=teststack_sam.yml
TEST_OUTPUT_FILE=$(CWD)/teststack_output.json

COMMON=$(CODE_DIR)/./task.go $(CODE_DIR)/./inspector.go $(CODE_DIR)/./alert.go $(CODE_DIR)/./report.go $(CODE_DIR)/./emitter.go $(CODE_DIR)/./functions/data_store.go $(CODE_DIR)/./functions/logger.go $(CODE_DIR)/./functions/aws_utils.go $(CODE_DIR)/./functions/init.go
FUNCTIONS=build/DummyReviewer build/DispatchInspection build/CompileReport build/ReceptAlert build/ErrorHandler build/StepFunctionError build/PublishReport build/SubmitContent build/FeedbackAttribute
TEST_FUNCTIONS=build/TestPublisher build/TestInspector
TEST_UTILS=$(CODE_DIR)/test/*.go

StackName=$(shell jsonnet $(DEPLOY_CONFIG) | jq .StackName)
TestStackName=$(shell jsonnet $(DEPLOY_CONFIG) | jq .TestStackName)
Region=$(shell jsonnet $(DEPLOY_CONFIG) | jq .Region)
CodeS3Bucket=$(shell jsonnet $(DEPLOY_CONFIG) | jq .CodeS3Bucket)
CodeS3Prefix=$(shell jsonnet $(DEPLOY_CONFIG) | jq .CodeS3Prefix)

# Functions ------------------------
build/DummyReviewer: $(CODE_DIR)/functions/DummyReviewer/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/DummyReviewer ./functions/DummyReviewer && cd $(CWD)
build/DispatchInspection: $(CODE_DIR)/functions/DispatchInspection/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/DispatchInspection ./functions/DispatchInspection && cd $(CWD)
build/CompileReport: $(CODE_DIR)/functions/CompileReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/CompileReport ./functions/CompileReport && cd $(CWD)
build/ReceptAlert: $(CODE_DIR)/functions/ReceptAlert/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/ReceptAlert ./functions/ReceptAlert && cd $(CWD)
build/ErrorHandler: $(CODE_DIR)/functions/ErrorHandler/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/ErrorHandler ./functions/ErrorHandler && cd $(CWD)
build/StepFunctionError: $(CODE_DIR)/functions/StepFunctionError/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/StepFunctionError ./functions/StepFunctionError && cd $(CWD)
build/PublishReport: $(CODE_DIR)/functions/PublishReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/PublishReport ./functions/PublishReport && cd $(CWD)
build/SubmitContent: $(CODE_DIR)/functions/SubmitContent/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/SubmitContent ./functions/SubmitContent && cd $(CWD)
build/FeedbackAttribute: $(CODE_DIR)/functions/FeedbackAttribute/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/FeedbackAttribute ./functions/FeedbackAttribute && cd $(CWD)


# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(TEMPLATE_FILE): $(STACK_CONFIG) $(TEMPLATE_JSONNET)
	jsonnet -J $(CODE_DIR) $(STACK_CONFIG) -o $(TEMPLATE_FILE)

$(SAM_FILE): $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(SAM_FILE)

$(OUTPUT_FILE): $(SAM_FILE)
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(SAM_FILE) \
		--stack-name $(StackName) \
		--no-fail-on-empty-changeset \
		--capabilities CAPABILITY_IAM $(PARAMETERS)
	aws cloudformation describe-stack-resources --stack-name $(StackName) > $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)

test: $(TEST_OUTPUT_FILE) $(TEST_UTILS)
	cd $(CODE_DIR) && env DEEPALERT_STACK_OUTPUT=$(OUTPUT_FILE) DEEPALERT_TEST_STACK_OUTPUT=$(TEST_OUTPUT_FILE) go test -v -count=1 . ./functions && cd $(CWD)

$(TEST_TEMPLATE_FILE): $(TEST_STACK_CONFIG) $(TEMPLATE_JSONNET)
	jsonnet -J $(CODE_DIR) $(TEST_STACK_CONFIG) -o $(TEST_TEMPLATE_FILE)

$(TEST_SAM_FILE): $(TEST_TEMPLATE_FILE) $(TEST_FUNCTIONS) $(OUTPUT_FILE)
	aws cloudformation package \
		--template-file $(TEST_TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(TEST_SAM_FILE)

$(TEST_OUTPUT_FILE): $(TEST_SAM_FILE)
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(TEST_SAM_FILE) \
		--stack-name $(TestStackName) \
		--capabilities CAPABILITY_IAM \
		--no-fail-on-empty-changeset
	aws cloudformation describe-stack-resources --region $(Region) \
        --stack-name $(TestStackName) > $(TEST_OUTPUT_FILE)

setuptest: $(TEST_OUTPUT_FILE)


# TestStack Functions ------------------------
build/TestPublisher: $(CODE_DIR)/test/TestPublisher/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/TestPublisher ./test/TestPublisher && cd $(CWD)
build/TestInspector: $(CODE_DIR)/test/TestInspector/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/TestInspector ./test/TestInspector && cd $(CWD)
