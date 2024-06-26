package files

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

// ensures that FileServer implements chunkServiceClient
var _ FileServiceServer = (*FileServer)(nil)

func New(port int) *FileServer {
	return &FileServer{
		port: port,
	}
}

type FileServer struct {
	UnimplementedFileServiceServer
	port int
}

func (s *FileServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Challenge: req.Challenge,
	}, nil
}

// required for grpc
func (s *FileServer) StoreFile(ctx context.Context, in *StoreFileRequest) (*StoreFileResponse, error) {
	return &StoreFileResponse{}, nil
}

// required for grpc
func (s *FileServer) GetFile(ctx context.Context, in *GetFileRequest) (*GetFileReponse, error) {
	return nil, nil
}

// if all goes well, this function will not return
func (s *FileServer) Start() {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", addr, err)
	}
	grpcServer := grpc.NewServer()
	RegisterFileServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
