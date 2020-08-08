all: build

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*.go
TEST_FUNCTIONS=build/deepalert/TestPublisher build/deepalert/TestInspector
TEST_UTILS=$(CODE_DIR)/test/*.go

FUNCTIONS= \
	build/deepalert/dummyReviewer \
	build/deepalert/dispatchInspection \
	build/deepalert/compileReport \
	build/deepalert/receptAlert \
	build/deepalert/errorHandler \
	build/deepalert/stepFunctionError \
	build/deepalert/publishReport \
	build/deepalert/submitContent \
	build/deepalert/feedbackAttribute

# Functions ------------------------
build/deepalert/dummyReviewer: $(CODE_DIR)/lambda/dummyReviewer/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/dummyReviewer ./lambda/dummyReviewer && cd $(CWD)
build/deepalert/dispatchInspection: $(CODE_DIR)/lambda/dispatchInspection/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/dispatchInspection ./lambda/dispatchInspection && cd $(CWD)
build/deepalert/compileReport: $(CODE_DIR)/lambda/compileReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/compileReport ./lambda/compileReport && cd $(CWD)
build/deepalert/receptAlert: $(CODE_DIR)/lambda/receptAlert/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/receptAlert ./lambda/receptAlert && cd $(CWD)
build/deepalert/errorHandler: $(CODE_DIR)/lambda/errorHandler/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/errorHandler ./lambda/errorHandler && cd $(CWD)
build/deepalert/stepFunctionError: $(CODE_DIR)/lambda/stepFunctionError/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/stepFunctionError ./lambda/stepFunctionError && cd $(CWD)
build/deepalert/publishReport: $(CODE_DIR)/lambda/publishReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/publishReport ./lambda/publishReport && cd $(CWD)
build/deepalert/submitContent: $(CODE_DIR)/lambda/submitContent/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/submitContent ./lambda/submitContent && cd $(CWD)
build/deepalert/feedbackAttribute: $(CODE_DIR)/lambda/feedbackAttribute/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/deepalert/feedbackAttribute ./lambda/feedbackAttribute && cd $(CWD)


# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)
