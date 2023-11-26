BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
FILE_BINARY=fileApp
LOGGER_BINARY=loggerApp


## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_broker build_file build_auth build_logger
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo "Building broker binary..."
	cd ../broker-service && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} ./cmd/broker-service
##	chmod +x ../broker-service/${BROKER_BINARY} 
	@echo "Done!"

## build_logger: builds the logger binary as a linux executable
build_file:
	@echo "Building file binary..."
	cd ../file-service && env GOOS=linux CGO_ENABLED=0 go build -o ${FILE_BINARY} ./cmd/file-service
##	chmod +x ../file-service/${FILE_BINARY} 
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	cd ../auth-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/auth-service
	@echo "Done!"

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo "Building logger binary..."
	cd ../logger-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LOGGER_BINARY} ./cmd/logger-service
	@echo "Done!"	

install:
	@echo "Installing broker dependencies..."
	cd ../broker-service && go install ./...
	@echo "Installing auth dependencies..."
	cd ../authentication-service && go install ./...
	@echo "Installing logger dependencies..."
	cd ../logger-service && go install ./...
