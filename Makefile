all: build

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*.go
TEST_FUNCTIONS=$(CODE_DIR)/build/TestPublisher $(CODE_DIR)/build/TestInspector
TEST_UTILS=$(CODE_DIR)/test/*.go

FUNCTIONS= \
	$(CODE_DIR)/build/dummyReviewer \
	$(CODE_DIR)/build/dispatchInspection \
	$(CODE_DIR)/build/compileReport \
	$(CODE_DIR)/build/receptAlert \
	$(CODE_DIR)/build/errorHandler \
	$(CODE_DIR)/build/stepFunctionError \
	$(CODE_DIR)/build/publishReport \
	$(CODE_DIR)/build/submitContent \
	$(CODE_DIR)/build/feedbackAttribute

# Functions ------------------------
$(CODE_DIR)/build/dummyReviewer: $(CODE_DIR)/lambda/dummyReviewer/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/dummyReviewer ./lambda/dummyReviewer && cd $(CWD)
$(CODE_DIR)/build/dispatchInspection: $(CODE_DIR)/lambda/dispatchInspection/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/dispatchInspection ./lambda/dispatchInspection && cd $(CWD)
$(CODE_DIR)/build/compileReport: $(CODE_DIR)/lambda/compileReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/compileReport ./lambda/compileReport && cd $(CWD)
$(CODE_DIR)/build/receptAlert: $(CODE_DIR)/lambda/receptAlert/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/receptAlert ./lambda/receptAlert && cd $(CWD)
$(CODE_DIR)/build/errorHandler: $(CODE_DIR)/lambda/errorHandler/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/errorHandler ./lambda/errorHandler && cd $(CWD)
$(CODE_DIR)/build/stepFunctionError: $(CODE_DIR)/lambda/stepFunctionError/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/stepFunctionError ./lambda/stepFunctionError && cd $(CWD)
$(CODE_DIR)/build/publishReport: $(CODE_DIR)/lambda/publishReport/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/publishReport ./lambda/publishReport && cd $(CWD)
$(CODE_DIR)/build/submitContent: $(CODE_DIR)/lambda/submitContent/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/submitContent ./lambda/submitContent && cd $(CWD)
$(CODE_DIR)/build/feedbackAttribute: $(CODE_DIR)/lambda/feedbackAttribute/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/feedbackAttribute ./lambda/feedbackAttribute && cd $(CWD)


# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)
