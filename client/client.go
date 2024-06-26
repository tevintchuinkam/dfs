package client

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/metadata"
	grpc "google.golang.org/grpc"
)

// ensures that FileServer implements chunkServiceClient
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

func (c *Client) newFileServiceClient(port int32) files.FileServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("could not connect. err: %v", err)
	}
	return files.NewFileServiceClient(conn)
}

func (c *Client) CreateFile(ctx context.Context, req *CreateFileRequest) (*CreateFileResponse, error) {
	// ask the mds on on what storage server to store the file
	mds := c.newMDSClient()
	rec, err := mds.GetStorageLocationRecommendation(context.Background(), &metadata.RecRequest{
		FileSize: int64(len(req.Data)),
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	fs := c.newFileServiceClient(rec.Port)
	fr, err := fs.CreateFile(ctx, &files.CreateFileRequest{
		Name: req.Name,
		Dir:  req.Dir,
		Data: req.Data,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	// send a createfilerequest to that server
	return &CreateFileResponse{
		BytesWritten: fr.BytesWritten,
	}, nil
}

func (c *Client) GetFile(ctx context.Context, req *GetFileRequest) (*GetFileResponse, error) {
	mds := c.newMDSClient()
	loc, err := mds.GetLocation(context.Background(), &metadata.LocRequest{
		Filename: req.Name,
		Dir:      req.Dir,
	})

	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	fs := c.newFileServiceClient(loc.Port)
	fr, err := fs.GetFile(ctx, &files.GetFileRequest{
		Name: req.Name,
		Dir:  req.Dir,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &GetFileResponse{
		Data: fr.Data,
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
