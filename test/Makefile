SETTINGS = settings.json

FUNCTIONS=build/TestInspector build/TestEmitter
TEMPLATE_FILE=template.json
TEMPLATE_JSONNET=template.jsonnet
SAM_FILE=sam.yml
OUTPUT_FILE=output.json

ifdef TAGS
TAGOPT=--tags $(TAGS)
else
TAGOPT=
endif

all: deploy

build/TestInspector: TestInspector/*
	env GOARCH=amd64 GOOS=linux go build -o ./build/TestInspector ./TestInspector
build/TestEmitter: TestEmitter/*
	env GOARCH=amd64 GOOS=linux go build -o ./build/TestEmitter ./TestEmitter

# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(TEMPLATE_FILE): $(SETTINGS) $(TEMPLATE_JSONNET)
	jsonnet $(TEMPLATE_JSONNET) -o $(TEMPLATE_FILE)

$(SAM_FILE): $(TEMPLATE_FILE) $(FUNCTIONS) $(DEPLOY_CONFIG)
	$(eval CodeS3Bucket := $(shell cat $(SETTINGS) | jq .CodeS3Bucket))
	$(eval CodeS3Prefix := $(shell cat $(SETTINGS) | jq .CodeS3Prefix))
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(SAM_FILE)

$(OUTPUT_FILE): $(SAM_FILE)
	$(eval StackName := $(shell cat $(SETTINGS) | jq .StackName))
	$(eval Region := $(shell cat $(SETTINGS) | jq .Region))
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(SAM_FILE) \
		--stack-name $(StackName) \
		--no-fail-on-empty-changeset \
		$(TAGOPT) \
		--capabilities CAPABILITY_IAM $(PARAMETERS)
	aws cloudformation describe-stack-resources --stack-name $(StackName) > $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
