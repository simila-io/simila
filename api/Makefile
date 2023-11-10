PROTOC = protoc
SOURCES = $(shell find . -name "*.proto")

.PHONY: all
all: clean compile simila.v1

.PHONY: clean
clean:
	@rm -rf gen/*
	@rm -rf genpublic/v1/*

.PHONY: compile
compile: $(SOURCES:.proto=.go)

%.go:%.proto
	@DIR=$(shell dirname $^);RESDIR=gen/;mkdir -p $$RESDIR;\
	FNAME=$(shell basename $^);$(PROTOC) -I $$DIR --go-grpc_out=$$RESDIR --go_out=$$RESDIR $$FNAME

.PHONY: oapi-gen
oapi-gen:
	@go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

.PHONY: simila.v1
simila.v1:
	@echo "generating go stubs for public Simila REST API v1 ..."
	@oapi-codegen --config openapi/v1/simila_gen_config.yaml openapi/v1/simila.yaml
