BIN_PATH?=.bin/nats-spliter

build:
	CGO_ENABLE=0 go build -mod=vendor -o ${BIN_PATH} ./cmd/*.go

run:build
	SRC_NATS_URL="nats://10.99.100.156:4222" \
	SRC_STAN_CLUSTER="test_cluster_name" \
    SRC_SUB_SUBJECT="producer" \
    STAN_CLIENT="test_compose" \
    SEPARATOR_NAME="sep" \
    DST_FILE_LOC="./config/dsts.json" \
	${BIN_PATH}; rm ${BIN_PATH}

deps:
	go test -v ./...
	go mod vendor