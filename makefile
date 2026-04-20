PROTO_DIR		=	proto
BUILD_DIR		=	bin
DEV_DATA_DIR	=	dev-data
DEV_CERTS_DIR	=	$(DEV_DATA_DIR)/certs
SCRIPTS_DIR		=	scripts
CLI_BIN_FILE	=	$(BUILD_DIR)/cli
SERVER_BIN_FILE	=	$(BUILD_DIR)/server 
WIRE_GEN_FILE	= 	internal/bootstrap


.PHONY: proto

utils:
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go@latest \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
		github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
		github.com/google/wire/cmd/wire@latest 

proto:
	protoc \
		--proto_path=. \
		--go_out=pkg \
		--go-grpc_out=pkg \
		$(PROTO_DIR)/*.proto

sqlc:
	sqlc generate 

wire:
	cd $(WIRE_GEN_FILE) && wire gen

generate: utils proto sqlc wire

certs:
	-mkdir $(DEV_CERTS_DIR)
	$(SCRIPTS_DIR)/dev-certs.sh $(DEV_CERTS_DIR)

install: 
	go mod download

build-server:
	go build -o $(SERVER_BIN_FILE) cmd/server

build-cli:
	go build -o $(CLI_BIN_FILE) cmd/cli

build:
	build-server
	build-cli

up: utils generate install build clean

clean: 
	go clean
	rm -rf $(BUILD_DIR) $(DEV_DATA_DIR)
