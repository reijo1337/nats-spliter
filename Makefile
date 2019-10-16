BIN_PATH?=.bin/nats-spliter

build:
	CGO_ENABLE=0 GOOS=linux go build -mod=vendor -o ${BIN_PATH} ./cmd/*.go

deps:
	go test -v ./...
	go mod vendor