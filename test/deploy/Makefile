COMMON=result.go
FUNCTIONS=build/testInspector build/testEmitter

all: build

build/testInspector: testInspector/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/testInspector ./testInspector
build/testEmitter: testEmitter/* $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o ./build/testEmitter ./testEmitter

# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

test: $(OUTPUT_FILE)
	go test -count=1 -v test
