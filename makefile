PROTO_DIR			=	proto
BUILD_DIR			=	bin
DEV_DATA_DIR		=	dev-data
DEV_CERTS_DIR		=	$(DEV_DATA_DIR)/certs
SCRIPTS_DIR			=	scripts
CLI_BIN_FILE		=	$(BUILD_DIR)/cli
SERVER_BIN_FILE		=	$(BUILD_DIR)/server 
WIRE_GEN_FILE		= 	internal/bootstrap
SERVICE_MOCKS_DIR	= internal/service/mocks
TEST_DIRS			= 	./internal/service/environment/ \
						./internal/service/variables/ \
 						./internal/service/policy/ \
 						./internal/service/subscription/ \
						./internal/domain/environment/services/access

.PHONY: proto test

install-utils:
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go@latest \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
		github.com/wireinject/wire/cmd/wire@latest \
		github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
		github.com/vektra/mockery/v2@latest

proto:
	protoc \
		--proto_path=. \
		--go_out=pkg \
		--go-grpc_out=pkg \
		$(PROTO_DIR)/*.proto

sqlc:
	sqlc generate 

mocks:
	mockery --config mockery.yaml --log-level error

wire:
	cd $(WIRE_GEN_FILE) && wire gen

generate: install-utils proto sqlc wire mocks

test: mocks
	go test $(TEST_DIRS)

test-cover: 
	go test --cover $(TEST_DIRS)

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
	-rm -rf $(BUILD_DIR) $(DEV_DATA_DIR) 
	-rm -rf $(find -type d -path "*mocks*") || true
