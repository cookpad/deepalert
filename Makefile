
DEEPALERT_CONFIG ?= config.json

ifeq (,$(wildcard $(DEEPALERT_CONFIG)))
    $(error $(DEEPALERT_CONFIG) is not found)
endif

TEMPLATE_FILE=template.yml
OUTPUT_FILE=sam.yml
COMMON=functions/*.go
FUNCTIONS=build/DummyReviewer build/DispatchInspection build/CompileReport build/SubmitReport build/ReceptAlert build/ErrorHandler build/StepFunctionError build/PublishReport

StackName    := $(shell cat $(DEEPALERT_CONFIG) | jq '.["StackName"]' -r)
Region       := $(shell cat $(DEEPALERT_CONFIG) | jq '.["Region"]' -r)
CodeS3Bucket := $(shell cat $(DEEPALERT_CONFIG) | jq '.["CodeS3Bucket"]' -r)
CodeS3Prefix := $(shell cat $(DEEPALERT_CONFIG) | jq '.["CodeS3Prefix"]' -r)

functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(OUTPUT_FILE): $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(OUTPUT_FILE) \
		--stack-name $(StackName) \
		--capabilities CAPABILITY_IAM

build/DummyReviewer: ./functions/DummyReviewer/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/DummyReviewer ./functions/DummyReviewer/
build/DispatchInspection: ./functions/DispatchInspection/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/DispatchInspection ./functions/DispatchInspection/
build/CompileReport: ./functions/CompileReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/CompileReport ./functions/CompileReport/
build/SubmitReport: ./functions/SubmitReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/SubmitReport ./functions/SubmitReport/
build/ReceptAlert: ./functions/ReceptAlert/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/ReceptAlert ./functions/ReceptAlert/
build/ErrorHandler: ./functions/ErrorHandler/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/ErrorHandler ./functions/ErrorHandler/
build/StepFunctionError: ./functions/StepFunctionError/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/StepFunctionError ./functions/StepFunctionError/
build/PublishReport: ./functions/PublishReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/PublishReport ./functions/PublishReport/