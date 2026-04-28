.PHONY: build test lint docker run proto clean gen-proto

BINARY=config-analyzer
IMAGE=config-analyzer
PORT=8080

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/main.go

test:
	go test -v ./...

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

gen-proto:
	protoc --go_out=. --go_opt=module=github.com/lopatich-privet/sec-config-scanner \
		--go-grpc_out=. --go-grpc_opt=module=github.com/lopatich-privet/sec-config-scanner \
		api/analyzer.proto

proto: gen-proto

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
