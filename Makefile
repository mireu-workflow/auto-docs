.PHONY: build test run docker-build clean

APP_NAME=auto-docs

build:
	@echo "Building Go binary..."
	go build -ldflags="-s -w" -o bin/$(APP_NAME) main.go

test:
	@echo "Running tests..."
	go test -v ./...

run: build
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)

docker-build:
	@echo "Building Docker container..."
	docker build -t $(APP_NAME):latest .

clean:
	@echo "Cleaning binaries..."
	rm -rf bin/ wiki_output/
