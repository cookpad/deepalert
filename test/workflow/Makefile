BASE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

COMMON=result.go
ASSETS=build/inspector build/emitter
CDKOUT=cdk.out/cdk.out
DEEPALERT_CODES=../../*.go ../../internal/**/*.go ../../lambda/**/*.go ../../cdk/*.ts

all: build

build/inspector: inspector/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/inspector ./inspector
build/emitter: emitter/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/emitter ./emitter

# Base Tasks -------------------------------------
build: $(ASSETS)

clean:
	rm $(ASSETS)

$(CDKOUT): build $(DEEPALERT_CODES) $(ASSETS)
	cd ../.. && tsc && cd $(BASE_DIR)
	cdk deploy "*"

deploy: $(DEEPALERT_CODES) $(ASSETS)
	cd ../.. && tsc && cd $(BASE_DIR)
	cdk deploy "*"

test: $(CDKOUT)
	go test -count=1 -v .
