PROTO_FILES := $(wildcard grpc/*.proto)
generate_protos:
	cd grpc && \
	for proto_file in $(notdir $(PROTO_FILES)); do \
	    protoc --go_out=. --go-grpc_out=. $$proto_file; \
	done





