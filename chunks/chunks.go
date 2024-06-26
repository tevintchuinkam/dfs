package chunks

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

func New(port int) *ChunkServer {
	return &ChunkServer{
		addr: fmt.Sprintf(":%d", port),
	}
}

type ChunkServer struct {
	UnimplementedChunkServiceServer
	addr string
}

// required for grpc
func (s *ChunkServer) StoreChunk(ctx context.Context, in *StoreChunkRequest) (*StoreChunkResponse, error) {
	return &StoreChunkResponse{}, nil
}

// required for grpc
func (s *ChunkServer) GetChunk(in *GetChunkRequest, src ChunkService_GetChunkServer) error {
	return nil
}

// if all goes well, this function will not return
func (s *ChunkServer) Start() {
	// accept connections
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", s.addr, err)
	}
	grpcServer := grpc.NewServer()
	RegisterChunkServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
