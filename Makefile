.PHONY: build test lint docker run proto clean

BINARY=config-analyzer
IMAGE=config-analyzer
PORT=8080

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/main.go

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

docker:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p $(PORT):$(PORT) $(IMAGE) --server --port $(PORT)

run: build
	./$(BINARY)

run-server: build
	./$(BINARY) --server --port $(PORT)

run-grpc: build
	./$(BINARY) --grpc --port $(PORT)

proto:
	protoc --go_out=. --go_opt=module=config-analyzer \
		--go-grpc_out=. --go-grpc_opt=module=config-analyzer \
		api/analyzer.proto

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
