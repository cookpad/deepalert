COMMON=result.go
FUNCTIONS=build/inspector build/emitter

all: build

build/inspector: inspector/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/inspector ./inspector
build/emitter: emitter/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/emitter ./emitter

# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

deploy: build
	make -f ../../Makefile
	cdk deploy "*"

test: $(OUTPUT_FILE)
	go test -count=1 -v test
