all: build

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*/*.go

FUNCTIONS= \
	$(CODE_DIR)/build/dummyReviewer/bootstrap \
	$(CODE_DIR)/build/dispatchInspection/bootstrap \
	$(CODE_DIR)/build/compileReport/bootstrap \
	$(CODE_DIR)/build/receptAlert/bootstrap \
	$(CODE_DIR)/build/submitReport/bootstrap \
	$(CODE_DIR)/build/publishReport/bootstrap \
	$(CODE_DIR)/build/submitFinding/bootstrap \
	$(CODE_DIR)/build/apiHandler/bootstrap \
	$(CODE_DIR)/build/feedbackAttribute/bootstrap

GO_OPT=-ldflags="-s -w" -trimpath

# Functions ------------------------
$(CODE_DIR)/build/dummyReviewer/bootstrap: $(CODE_DIR)/lambda/dummyReviewer/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/dummyReviewer
$(CODE_DIR)/build/dispatchInspection/bootstrap: $(CODE_DIR)/lambda/dispatchInspection/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/dispatchInspection
$(CODE_DIR)/build/compileReport/bootstrap: $(CODE_DIR)/lambda/compileReport/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/compileReport
$(CODE_DIR)/build/receptAlert/bootstrap: $(CODE_DIR)/lambda/receptAlert/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/receptAlert
$(CODE_DIR)/build/publishReport/bootstrap: $(CODE_DIR)/lambda/publishReport/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/publishReport
$(CODE_DIR)/build/submitReport/bootstrap: $(CODE_DIR)/lambda/submitReport/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/submitReport
$(CODE_DIR)/build/submitFinding/bootstrap: $(CODE_DIR)/lambda/submitFinding/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/submitFinding
$(CODE_DIR)/build/apiHandler/bootstrap: $(CODE_DIR)/lambda/apiHandler/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/apiHandler
$(CODE_DIR)/build/feedbackAttribute/bootstrap: $(CODE_DIR)/lambda/feedbackAttribute/*.go $(COMMON)
	mkdir -p $(dir $@)
	env GOARCH=amd64 GOOS=linux go build $(GO_OPT) -o $@ ./lambda/feedbackAttribute


# Base Tasks -------------------------------------
# build: $(FUNCTIONS) $(CDK_STACK)
build: $(FUNCTIONS)

