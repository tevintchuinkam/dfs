package chunks

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

// ensures that ChunkServer implements chunkServiceClient
var _ ChunkServiceServer = (*ChunkServer)(nil)

func New(port int) *ChunkServer {
	return &ChunkServer{
		port: port,
	}
}

type ChunkServer struct {
	UnimplementedChunkServiceServer
	port int
}

func (s *ChunkServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Challenge: req.Challenge,
	}, nil
}

// required for grpc
func (s *ChunkServer) StoreChunk(ctx context.Context, in *StoreChunkRequest) (*StoreChunkResponse, error) {
	return &StoreChunkResponse{}, nil
}

// required for grpc
func (s *ChunkServer) GetChunk(ctx context.Context, in *GetChunkRequest) (*GetChunkReponse, error) {
	return nil, nil
}

// if all goes well, this function will not return
func (s *ChunkServer) Start() {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", addr, err)
	}
	grpcServer := grpc.NewServer()
	RegisterChunkServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
