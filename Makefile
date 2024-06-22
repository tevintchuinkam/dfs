
proto-chunks:
	cd grpc && protoc --go_out=. --go-grpc_out=. chunks.proto 

proto-metadata:
	cd grpc && protoc --go_out=. --go-grpc_out=. metadata.proto 



