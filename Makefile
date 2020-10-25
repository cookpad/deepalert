all: build

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*/*.go
TEST_FUNCTIONS=$(CODE_DIR)/build/TestPublisher $(CODE_DIR)/build/TestInspector
TEST_UTILS=$(CODE_DIR)/test/*.go

CDK_STACK=$(CODE_DIR)/cdk/deepalert-stack.js

FUNCTIONS= \
	$(CODE_DIR)/build/dummyReviewer \
	$(CODE_DIR)/build/dispatchInspection \
	$(CODE_DIR)/build/compileReport \
	$(CODE_DIR)/build/receptAlert \
	$(CODE_DIR)/build/errorHandler \
	$(CODE_DIR)/build/submitReport \
	$(CODE_DIR)/build/publishReport \
	$(CODE_DIR)/build/submitContent \
	$(CODE_DIR)/build/apiHandler \
	$(CODE_DIR)/build/feedbackAttribute

GO_OPT=-ldflags="-s -w" -trimpath

# Functions ------------------------
$(CODE_DIR)/build/dummyReviewer: $(CODE_DIR)/lambda/dummyReviewer/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/dummyReviewer ./lambda/dummyReviewer && cd $(CWD)
$(CODE_DIR)/build/dispatchInspection: $(CODE_DIR)/lambda/dispatchInspection/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/dispatchInspection ./lambda/dispatchInspection && cd $(CWD)
$(CODE_DIR)/build/compileReport: $(CODE_DIR)/lambda/compileReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/compileReport ./lambda/compileReport && cd $(CWD)
$(CODE_DIR)/build/receptAlert: $(CODE_DIR)/lambda/receptAlert/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/receptAlert ./lambda/receptAlert && cd $(CWD)
$(CODE_DIR)/build/errorHandler: $(CODE_DIR)/lambda/errorHandler/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/errorHandler ./lambda/errorHandler && cd $(CWD)
$(CODE_DIR)/build/publishReport: $(CODE_DIR)/lambda/publishReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/publishReport ./lambda/publishReport && cd $(CWD)
$(CODE_DIR)/build/submitReport: $(CODE_DIR)/lambda/submitReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/submitReport ./lambda/submitReport && cd $(CWD)
$(CODE_DIR)/build/submitContent: $(CODE_DIR)/lambda/submitContent/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/submitContent ./lambda/submitContent && cd $(CWD)
$(CODE_DIR)/build/apiHandler: $(CODE_DIR)/lambda/apiHandler/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/apiHandler ./lambda/apiHandler && cd $(CWD)
$(CODE_DIR)/build/feedbackAttribute: $(CODE_DIR)/lambda/feedbackAttribute/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $(CODE_DIR)/build/feedbackAttribute ./lambda/feedbackAttribute && cd $(CWD)


# Base Tasks -------------------------------------
build: $(FUNCTIONS) $(CDK_STACK)

$(CDK_STACK): $(CODE_DIR)/cdk/*.ts
	cd $(CODE_DIR) && tsc && cd $(CWD)

deploy: $(FUNCTIONS) $(CDK_STACK)
	cdk deploy

clean:
	rm $(FUNCTIONS)
