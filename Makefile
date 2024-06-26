PROTO_FILES := $(wildcard proto/*.proto)
generate_protos:
	cd proto && \
	for proto_file in $(notdir $(PROTO_FILES)); do \
	    protoc --go_out=.. --go-grpc_out=.. $$proto_file; \
	done





