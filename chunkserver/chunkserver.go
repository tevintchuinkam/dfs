package main

import (
	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/grpc/chunks"
	"google.golang.org/grpc"
)

func main() {
	// accept connections
	port := ":9000"
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}
	s := chunks.Server{}
	grpcServer := grpc.NewServer()
	chunks.RegisterChunkServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		// https://www.youtube.com/watch?v=BdzYdN_Zd9Q
	}

	// store the incoming files

	// return the requested files

}
