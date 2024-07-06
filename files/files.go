package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path"
	sync "sync"

	"github.com/tevintchuinkam/tdfs/grep"

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
		rootDir: path.Join("./", "dfs-data", fmt.Sprint(port)),
	}
}

type FileServer struct {
	UnimplementedFileServiceServer
	port       int
	rootDir    string
	grpcServer *grpc.Server
	listener   net.Listener
	mu         sync.Mutex
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
	file, err := os.Open(path.Join(s.rootDir, path.Clean(in.Name)))
	if err != nil {
		slog.Error("could not open file", "err", err)
		return nil, err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("could not read bytes from file", "file", in.Name, "err", err)
		return nil, err
	}
	res := new(File)
	res.Data = bytes
	return res, nil
}

// if all goes well, this function will not return
func (s *FileServer) Start() {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s error=%v", addr, err)
	}
	s.listener = lis
	grpcServer := grpc.NewServer()
	s.grpcServer = grpcServer
	RegisterFileServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

// Stop gracefully stops the MetaDataServer.
func (s *FileServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.grpcServer = nil
	}

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

func (s *FileServer) Grep(ctx context.Context, req *GrepRequest) (*GrepResponse, error) {
	file, err := os.Open(path.Join(s.rootDir, path.Clean(req.FileName)))
	if err != nil {
		slog.Error("could not open file", "err", err)
		return nil, err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("could not read bytes from file", "file", req.FileName, "err", err)
		return nil, err
	}
	count := grep.CountWordOccurrences(bytes, req.Word)
	return &GrepResponse{
		Count: int64(count),
	}, nil
}
