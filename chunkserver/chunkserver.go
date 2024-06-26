package main

import (
	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/grpc/grpc_chunks"
	"google.golang.org/grpc"
)

func main() {
	// accept connections
	port := ":9000"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", port, err)
	}
	s := grpc_chunks.ChunkServer{}
	grpcServer := grpc.NewServer()
	grpc_chunks.RegisterChunkServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		// https://www.youtube.com/watch?v=BdzYdN_Zd9Q
	}

	// store the incoming files

	// return the requested files

}
