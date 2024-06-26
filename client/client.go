package client

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

// ensures that ChunkServer implements chunkServiceClient
var _ ClientServiceServer = (*Client)(nil)

type Client struct {
	UnimplementedClientServiceServer
	port int
}

func New(port int) *Client {
	return &Client{
		port: port,
	}
}

func (c *Client) Start() {
	// accept connections
	addr := fmt.Sprintf(":%d", c.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer()
	RegisterClientServiceServer(grpcServer, c)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (c *Client) CreateFile(ctx context.Context, req *CreateFileRequest) (*CreateFileResponse, error) {
	return &CreateFileResponse{
		BytesWritten: 0,
	}, nil
}

func (c *Client) GetFile(ctx context.Context, req *GetFileRequest) (*GetFileResponse, error) {
	return &GetFileResponse{
		Name: "example.txt",
		Data: []byte{},
	}, nil
}

func (c *Client) MkDir(ctx context.Context, req *MkDirRequest) (*MkDirResponse, error) {
	return &MkDirResponse{
		Depth: 1,
	}, nil
}

func (c *Client) Grep(ctx context.Context, req *GrepRequest) (*GrepReponse, error) {
	return &GrepReponse{
		Results: []string{},
	}, nil
}

func (c *Client) LS(ctx context.Context, req *LSRequest) (*LSResponse, error) {
	return &LSResponse{
		Files: []string{},
		Dirs:  []string{},
	}, nil
}

func (c *Client) Tree(ctx context.Context, req *TreeRequest) (*TreeResponse, error) {
	return &TreeResponse{
		Tree: "-",
	}, nil
}
