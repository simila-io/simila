PROTOC = protoc
SOURCES = $(shell find . -name "*.proto")

.PHONY: all
all: clean compile

.PHONY: clean
clean:
	@rm -rf protogen/*

.PHONY: compile
compile: $(SOURCES:.proto=.go)

%.go:%.proto
	@DIR=$(shell dirname $^);RESDIR=gen/;mkdir -p $$RESDIR;\
	FNAME=$(shell basename $^);$(PROTOC) -I $$DIR -I include/ --grpc-gateway_out $$RESDIR $$FNAME
