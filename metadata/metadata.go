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
	"time"

	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/helpers"
	"github.com/tevintchuinkam/tdfs/interceptors"
	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// ensures that FileServer implements chunkServiceClient
var _ MetadataServiceServer = (*MetaDataServer)(nil)

type chunkServer struct {
	client *files.FileServiceClient
	port   int

	// how many bytes are stored in the given
	load int
	mu   sync.Mutex
}

type MetaDataServer struct {
	UnimplementedMetadataServiceServer
	port         int
	muFile       sync.Mutex
	chunkServers []*chunkServer
	rootDir      directory
	// map from file to file server address
	fileLocation map[string]*chunkServer
}

func New(port int) *MetaDataServer {
	return &MetaDataServer{
		port: port,
	}
}

func (s *MetaDataServer) Start(requestDelay time.Duration) {
	// accept connections
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.DelayInterceptor(requestDelay)),
		grpc.StreamInterceptor(interceptors.DelayStreamInterceptor(requestDelay)),
	)
	RegisterMetadataServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (s *MetaDataServer) RegisterFileServer(port int) error {
	c := helpers.NewFileServiceClient(int32(port))
	// ping the server
	challenge := rand.Int63()
	resp, err := c.Ping(context.Background(), &files.PingRequest{Challenge: challenge})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if challenge != resp.Challenge {
		return fmt.Errorf("file server failed challenge: %d !=  %d (expected)", resp.Challenge, challenge)
	}
	s.muFile.Lock()
	srv := new(chunkServer)
	srv.client = &c
	srv.port = port
	s.chunkServers = append(s.chunkServers, srv)
	s.muFile.Unlock()
	return nil
}

// assumes the client will indeed write the data to the given client
func (s *MetaDataServer) GetStorageLocationRecommendation(ctx context.Context, req *RecRequest) (*RecResponse, error) {
	if len(s.chunkServers) == 0 {
		return nil, errors.New("no file servers have been registered")
	}
	min := s.chunkServers[0]
	minLoad := min.load
	for _, s := range s.chunkServers {
		s.mu.Lock()
		if s.load < minLoad {
			min = s
		}
		s.mu.Unlock()
	}
	min.mu.Lock()
	min.load += int(req.FileSize)
	min.mu.Unlock()
	return &RecResponse{
		Port: int32(min.port),
	}, nil
}

func (s *MetaDataServer) GetLocation(ctx context.Context, req *LocRequest) (*LocResponse, error) {
	fullPath := path.Clean(req.Name)
	cs, ok := s.fileLocation[fullPath]
	if !ok {
		return nil, fmt.Errorf("the file %s doesn't exist", fullPath)
	}
	return &LocResponse{
		Port: int32(cs.port),
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
