PROTO_FILES := $(wildcard proto/*.proto)
protos:
	cd proto && \
	for proto_file in $(notdir $(PROTO_FILES)); do \
	    protoc --go_out=.. --go-grpc_out=.. $$proto_file; \
	done

build: 
	go build .



