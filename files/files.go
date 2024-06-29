package files

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/tevintchuinkam/tdfs/interceptors"
	grpc "google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// ensures that FileServer implements chunkServiceClient
var _ FileServiceServer = (*FileServer)(nil)

func New(port int) *FileServer {
	return &FileServer{
		port:    port,
		rootDir: path.Join("./", "dfs", fmt.Sprint(port)),
	}
}

type FileServer struct {
	UnimplementedFileServiceServer
	port    int
	rootDir string
}

func (s *FileServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Challenge: req.Challenge,
	}, nil
}

// required for grpc
func (s *FileServer) CreateFile(ctx context.Context, in *CreateFileRequest) (*CreateFileResponse, error) {
	// file server uses his own port as dir name
	if err := os.MkdirAll(path.Join(s.rootDir, path.Dir(in.Name)), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	p := path.Join(s.rootDir, in.Name)
	file, err := os.Create(p)
	if err != nil {
		log.Fatal(err)
	}
	written, err := file.Write(in.Data)
	if err != nil {
		log.Fatal(err)
	}
	return &CreateFileResponse{BytesWritten: int64(written)}, nil
}

// required for grpc
func (s *FileServer) GetFile(ctx context.Context, in *GetFileRequest) (*File, error) {
	return nil, nil
}

// if all goes well, this function will not return
func (s *FileServer) Start(requestDelay time.Duration) {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", addr, err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.DelayInterceptor(requestDelay)),
		grpc.StreamInterceptor(interceptors.DelayStreamInterceptor(requestDelay)),
	)
	RegisterFileServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
