PROTO_DIR		=	proto
BUILD_DIR		=	bin
DEV_DATA_DIR	=	dev-data
DEV_CERTS_DIR	=	$(DEV_DATA_DIR)/certs
SCRIPTS_DIR		=	scripts
CLI_BIN_FILE	=	$(BUILD_DIR)/envmn-cli
SERVER_BIN_FILE	=	$(BUILD_DIR)/envmn-server 

.PHONY: proto

utils:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

proto:
	protoc \
		--proto_path=. \
		--go_out=pkg \
		--go-grpc_out=pkg \
		$(PROTO_DIR)/*.proto

sqlc:
	sqlc generate

generate: proto sqlc

install: 
	go mod download

build %:
	go build -o bin/$* cmd/$*

# build: build-server build-cli

clean: 
	go clean
	rm -rf $(BUILD_DIR) $(DEV_DATA_DIR)

certs:
	-mkdir $(DEV_CERTS_DIR)
	$(SCRIPTS_DIR)/dev-certs.sh $(DEV_CERTS_DIR)

up: utils generate install build clean

