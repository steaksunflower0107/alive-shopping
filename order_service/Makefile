.PHONY: all build run gotool clean gen help

BINARY="goods_srv"
PROTO_DIR=proto

all: gotool build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}

run:
	@go run ./cmd/main.go conf/config.yaml

gotool:
	go fmt ./
	go vet ./

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

gen:
	protoc \
	--proto_path=$(PROTO_DIR) \
	--go_out=$(PROTO_DIR) \
	--go_opt=paths=source_relative \
	--go-grpc_out=$(PROTO_DIR) \
	--go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=$(PROTO_DIR) \
	--grpc-gateway_opt=paths=source_relative \
	$(shell find $(PROTO_DIR) -iname "*.proto")

help:
	@echo "make - 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make build - 编译 Go 代码, 生成二进制文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除二进制文件和 vim swap files"
	@echo "make gotool - 运行 Go 工具 'fmt' and 'vet'"
	@echo "make gen - 生成pb及grpc代码"
