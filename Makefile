.PHONY: build test vet lint run clean

APP := bin/aicodingagentteam
PKG := ./cmd/aicodingagentteam

build:
	go build -o $(APP) $(PKG)

test:
	go test ./... -v -count=1

vet:
	go vet ./...

lint:
	golangci-lint run ./...

run: build
	$(APP) serve

quick: build
	$(APP) quick "$(MSG)"

clean:
	rm -rf bin/ .aicodingagentteam/ output/

all: lint vet test build