
proto-chunks:
	cd grpc && protoc --go-grpc_out=. chunks.proto 

proto-metadata:
	cd grpc && protoc --go-grpc_out=. metadata.proto 
chat:
	cd grpc && protoc --go-grpc_out=. chat.proto 


