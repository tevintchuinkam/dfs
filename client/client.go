package client

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/metadata"
	grpc "google.golang.org/grpc"
)

// ensures that ChunkServer implements chunkServiceClient
var _ ClientServiceServer = (*Client)(nil)

type Client struct {
	UnimplementedClientServiceServer
	port    int
	mdsPort int
}

func New(port int, mdsPort int) *Client {
	// ping the server
	return &Client{
		port:    port,
		mdsPort: mdsPort,
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

func (c *Client) newMDSClient() metadata.MetadataServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", c.mdsPort))
	if err != nil {
		log.Fatalf("could not connect. err: %v", err)
	}
	return metadata.NewMetadataServiceClient(conn)
}

func (c *Client) CreateFile(ctx context.Context, req *CreateFileRequest) (*CreateFileResponse, error) {
	// ask the mds on on what storage server to store the file
	client := c.newMDSClient()
	client.GetStorageLocationRecommendation(context.Background(), &client.)
	// send a createfilerequest to that server
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
