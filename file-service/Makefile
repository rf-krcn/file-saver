generate:
	export PATH=$PATH:$(go env GOPATH)/bin
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=. --go-grpc_out=. ./cmd/api/file_service.proto