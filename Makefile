# BUILD
MODULE=github.com/simila-io/simila
EXEC_NAME=simila
BUILD_DIR=./build
TEST_DIR=$(BUILD_DIR)/utests
BUILD_SRC=./cmd/$(EXEC_NAME)
BUILD_OUT=$(BUILD_DIR)/$(EXEC_NAME)
# Versioning/build metadata
VERSION?=$(shell curl -s https://raw.githubusercontent.com/acquirecloud/appversion/main/version.sh | bash -s -- -s)
REV=$(shell git rev-parse HEAD)
NOW=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# TEST_FLAGS ?= -race
LDFLAGS="-X '$(MODULE)/pkg/version.Version=$(VERSION)' \
		 -X '$(MODULE)/pkg/version.GitCommit=$(REV)' \
		 -X '$(MODULE)/pkg/version.BuildDate=$(NOW)' \
		 -X '$(MODULE)/pkg/version.GoVersion=$(shell go version)' "

#Later
REGISTRY?=<where we host it>.dkr.ecr.us-east-2.amazonaws.com
IMAGE_NAME?=$(EXEC_NAME)
IMAGE_TAG?=$(shell git rev-parse --short HEAD)
IMAGE=${IMAGE_NAME}:${IMAGE_TAG}

# Help by default
default: help

# Build/Run/Clean
.PHONY: fmt
fmt: ## apply fmt to the source code
	@go fmt ./...

.PHONY: builddeps
builddeps: ## install dependency tools
	@go install \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
	google.golang.org/protobuf/cmd/protoc-gen-go \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: fmt-check
fmt-check: ## check formatting of the source code
ifneq (,$(shell gofmt -l .))
	@echo 'please consider reformat the following files:'
	@gofmt -l .
	@echo "or just run: 'make fmt'"
	@exit 1
endif

.PHONY: lint-go
lint-go: ## run golang linter against the source code
	@golangci-lint run ./...

.PHONY: test
test: ## run unit-tests
	mkdir -p ${TEST_DIR}; CGO_ENABLED=1; go test $(TEST_FLAGS) -v -coverprofile=${TEST_DIR}/c.out ./...
	go tool cover -html=${TEST_DIR}/c.out -o ${TEST_DIR}/coverage.html

.PHONY: build
build: fmt-check ## builds the executable and places it to ./build/simila
	go build -ldflags=$(LDFLAGS) -o ${BUILD_OUT} ${BUILD_SRC}

.PHONY: run
run: build ## builds and runs the server locally: `./build/simila start`
	@${BUILD_OUT} start

clean: ## clean up, removes the ./build directory
	@rm -rf ${BUILD_DIR}

all: clean builddeps build

# docker
.PHONY: docker-build
docker-build: ## builds the docker image
	DOCKER_BUILDKIT=1 docker build --no-cache --ssh default -f Dockerfile -t ${IMAGE} .
	docker tag ${IMAGE} ${IMAGE_NAME}:latest

.PHONY: docker-push
docker-push: docker-build ## pushes docker image into registry
	docker tag ${IMAGE} ${REGISTRY}/${IMAGE}
	docker tag ${IMAGE} ${REGISTRY}/${IMAGE_NAME}:latest
	docker push ${REGISTRY}/${IMAGE}
	docker push ${REGISTRY}/${IMAGE_NAME}:latest

.PHONY: db-start
db-start: ## start the postgres server locally in docker
	@mkdir -p testdata/postgres
	@docker run --rm --name postgres-db -v $(shell pwd)/testdata:/testdata \
		-v $(shell pwd)/testdata/postgres:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=simila -d -p 5432:5432 groonga/pgroonga:latest-debian-16
	@echo "Database started at port 5423"

.PHONY: db-stop
db-stop: ## stop the database server
	@docker stop postgres-db
	@echo "Database stopped"

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'