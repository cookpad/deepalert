all: build

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*/*.go

FUNCTIONS= \
	$(CODE_DIR)/build/dummyReviewer \
	$(CODE_DIR)/build/dispatchInspection \
	$(CODE_DIR)/build/compileReport \
	$(CODE_DIR)/build/receptAlert \
	$(CODE_DIR)/build/submitReport \
	$(CODE_DIR)/build/publishReport \
	$(CODE_DIR)/build/submitFinding \
	$(CODE_DIR)/build/apiHandler \
	$(CODE_DIR)/build/feedbackAttribute

GO_OPT=-ldflags="-s -w" -trimpath

# Functions ------------------------
$(CODE_DIR)/build/dummyReviewer: $(CODE_DIR)/lambda/dummyReviewer/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/dummyReviewer ./lambda/dummyReviewer
$(CODE_DIR)/build/dispatchInspection: $(CODE_DIR)/lambda/dispatchInspection/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/dispatchInspection ./lambda/dispatchInspection
$(CODE_DIR)/build/compileReport: $(CODE_DIR)/lambda/compileReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/compileReport ./lambda/compileReport
$(CODE_DIR)/build/receptAlert: $(CODE_DIR)/lambda/receptAlert/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/receptAlert ./lambda/receptAlert
$(CODE_DIR)/build/publishReport: $(CODE_DIR)/lambda/publishReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/publishReport ./lambda/publishReport
$(CODE_DIR)/build/submitReport: $(CODE_DIR)/lambda/submitReport/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/submitReport ./lambda/submitReport
$(CODE_DIR)/build/submitFinding: $(CODE_DIR)/lambda/submitFinding/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/submitFinding ./lambda/submitFinding
$(CODE_DIR)/build/apiHandler: $(CODE_DIR)/lambda/apiHandler/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/apiHandler ./lambda/apiHandler
$(CODE_DIR)/build/feedbackAttribute: $(CODE_DIR)/lambda/feedbackAttribute/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/feedbackAttribute ./lambda/feedbackAttribute


# Base Tasks -------------------------------------
# build: $(FUNCTIONS) $(CDK_STACK)
build: $(FUNCTIONS)

# Use in cdk deploy (bundling of lambda.Code.fromAsset)
asset: build
	cp $(CODE_DIR)/build/* /asset-output

