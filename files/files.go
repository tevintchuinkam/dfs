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

	"github.com/tevintchuinkam/dfs/grep"

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

func (s *FileServer) CreateFileWithStream(stream FileService_CreateFileWithStreamServer) error {
	req, err := stream.Recv()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	name := req.GetInfo().Name
	if err := os.MkdirAll(path.Join(s.rootDir, path.Dir(name)), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	p := path.Join(s.rootDir, name)
	file, err := os.Create(p)
	if err != nil {
		log.Fatal(err)
	}

	fileSize := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		chunk := req.GetChunkData()
		n, err := file.Write(chunk)
		if err != nil {
			log.Fatal(err)
		}
		fileSize += n
	}
	res := &CreateFileWithStreamResponse{
		BytesWritten: int64(fileSize),
	}
	err = stream.SendAndClose(res)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (s *FileServer) GetFileWithStream(req *GetFileWithStreamRequest, stream FileService_GetFileWithStreamServer) error {
	// Build the file path
	filePath := path.Join(s.rootDir, req.GetName())

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("failed to open file", "err", err.Error())
		return err
	}
	defer file.Close()

	// Create a buffer to read chunks of the file
	buf := make([]byte, 1024)

	for {
		// Read a chunk from the file
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed to read file", "err", err.Error())
			return err
		}

		// Send the chunk to the client
		err = stream.Send(&GetFileWithStreamResponse{
			ChunkData: buf[:n],
		})
		if err != nil {
			slog.Error("failed to send chunk", "err", err.Error())
			return err
		}
	}

	return nil
}
