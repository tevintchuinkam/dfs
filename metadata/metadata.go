package metadata

import (
	context "context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"path"
	"strings"
	"sync"

	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/chunks"
	"google.golang.org/grpc"
)

// ensures that ChunkServer implements chunkServiceClient
var _ MetadataServiceServer = (*MetaDataServer)(nil)

type chunkServer struct {
	client *chunks.ChunkServiceClient
	port   int

	// how many bytes are stored in the given
	load int
	mu   sync.Mutex
}

type MetaDataServer struct {
	UnimplementedMetadataServiceServer
	port         int
	muChunk      sync.Mutex
	chunkServers []*chunkServer
	rootDir      directory
	// map from chunk to chunk server address
	fileLocation map[string]chunkServer
}

func New(port int) *MetaDataServer {
	return &MetaDataServer{
		port: port,
	}
}

func (s *MetaDataServer) Start() {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer()
	RegisterMetadataServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (s *MetaDataServer) RegisterChunkServer(port int) error {
	// ping the server
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("could not connect. err: %v", err)
	}
	defer conn.Close()
	c := chunks.NewChunkServiceClient(conn)
	// ping the server
	challenge := rand.Int63()
	resp, err := c.Ping(context.Background(), &chunks.PingRequest{Challenge: challenge})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if challenge != resp.Challenge {
		return fmt.Errorf("chunk server failed challenge: %d !=  %d (expected)", resp.Challenge, challenge)
	}
	s.muChunk.Lock()
	srv := new(chunkServer)
	srv.client = &c
	srv.port = port
	s.chunkServers = append(s.chunkServers, srv)
	s.muChunk.Unlock()
	return nil
}

func (s *MetaDataServer) GetMetadata(ctx context.Context, in *MetadataRequest) (*MetadataResponse, error) {
	slog.Info("received chunk data request", "filename", in.Filename, "chunk_index", in.ChunkIndex)
	return &MetadataResponse{
		ChunkHandle: "abcde",
		Url:         "chunk:9000",
	}, nil
}

func (s *MetaDataServer) GetStorageLocationRecommendation(ctx context.Context, req *LocRequest) (*LocResponse, error) {
	if len(s.chunkServers) == 0 {
		return nil, errors.New("no chunk servers have been registered")
	}
	minLoad := -1
	port := s.chunkServers[0].port
	for _, s := range s.chunkServers {
		if s.load < minLoad {
			port = s.port
		}
	}
	return &LocResponse{
		Port: int32(port),
	}, nil
}

type directory struct {
	name    string
	prev    *directory
	subDirs []*directory
	files   []string
}

func (d *directory) String() string {
	if d == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(d.name)
	for _, file := range d.files {
		sb.WriteString(fmt.Sprintf("\n\t├── %s", file))
	}
	return sb.String()
}

func (d *directory) WalkTo(p string) (*directory, error) {
	base := path.Base(p)
	rest := strings.TrimPrefix(p, base)
	for _, dir := range d.subDirs {
		if base == dir.name {
			return dir.WalkTo(rest)
		}
	}
	return nil, fmt.Errorf("directory does not exist: %s", base)
}
