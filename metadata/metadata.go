package metadata

import (
	context "context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand"
	"path"
	"sync"
	"time"

	"log"
	"net"

	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/helpers"
	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// ensures that FileServer implements chunkServiceClient
var _ MetadataServiceServer = (*MetaDataServer)(nil)

type fileServer struct {
	client *files.FileServiceClient
	port   int

	// how many bytes are stored in the given
	load   int
	muLoad sync.Mutex
}

type MetaDataServer struct {
	UnimplementedMetadataServiceServer
	port        int
	muFile      sync.Mutex
	fileServers []*fileServer
	muDir       sync.Mutex
	rootDir     *fileInfo
	// map from file to file server address
	fileLocation map[string]*fileServer
}

func New(port int) *MetaDataServer {
	return &MetaDataServer{
		port:        port,
		muFile:      sync.Mutex{},
		fileServers: []*fileServer{},
		muDir:       sync.Mutex{},
		rootDir: &fileInfo{
			name:  ".",
			isDir: true,
		},
		fileLocation: make(map[string]*fileServer),
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
	srv := new(fileServer)
	srv.client = &c
	srv.port = port
	s.fileServers = append(s.fileServers, srv)
	s.muFile.Unlock()

	return nil
}

func (s *MetaDataServer) MkDir(ctx context.Context, in *MkDirRequest) (*MkDirResponse, error) {
	dir := path.Clean(in.Name)

	// Check if the parent directory exists
	if !isDir(s.rootDir, path.Dir(dir)) {
		slog.Error("parent directory does not exist", "dir", dir)
		return nil, fmt.Errorf("parent directory %s does not exist", dir)
	}
	res := new(MkDirResponse)
	res.Name = dir
	// Check if the directory to be created already exists
	if entryAlreadyExists(s.rootDir, dir) {
		return res, nil
	}

	// Lock the directory structure and create the new directory
	s.muDir.Lock()
	defer s.muDir.Unlock()
	newDir := &fileInfo{
		name:     path.Base(dir),
		fullPath: dir,
		isDir:    true,
	}
	err := storeFileInfo(s.rootDir, dir, newDir)
	if err != nil {
		slog.Error("failed to store new directory info", "dir", dir, "error", err)
		return nil, err
	}

	return &MkDirResponse{Name: dir}, nil
}

// assumes the client will indeed write the data to the given client
func (s *MetaDataServer) RegisterFileCreation(ctx context.Context, req *RecRequest) (*RecResponse, error) {
	// first check to see if the dir even exists where the file is supposed to be placed
	p := path.Clean(req.Name)
	dir := path.Dir(p)
	if !isDir(s.rootDir, dir) {
		slog.Error("directory does not exist", "dir", dir, "original_dir", req.Name)
		return nil, fmt.Errorf("directory %s does not exist", dir)
	}
	if entryAlreadyExists(s.rootDir, p) {
		err := fmt.Errorf("file %s already exists", p)
		slog.Error(err.Error())
		return nil, err
	}
	// also check if a file with the given name already exists
	if len(s.fileServers) == 0 {
		return nil, errors.New("no file servers have been registered")
	}
	min := s.fileServers[0]
	minLoad := min.load
	for _, s := range s.fileServers {
		s.muLoad.Lock()
		if s.load < minLoad {
			min = s
		}
		s.muLoad.Unlock()
	}
	min.muLoad.Lock()
	min.load += int(req.FileSize)
	min.muLoad.Unlock()
	s.muDir.Lock()
	if err := storeFileInfo(s.rootDir, p, &fileInfo{
		name:     path.Base(p),
		fullPath: p,
		size:     req.FileSize,
		port:     min.port,
		isDir:    false,
	}); err != nil {
		log.Fatal(err)
	}
	s.fileLocation[p] = min
	s.muDir.Unlock()
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

func (s *MetaDataServer) OpenDir(ctx context.Context, req *OpenDirRequest) (*OpenDirResponse, error) {
	p := path.Clean(req.Name)
	// check if this is a valid directory
	if !isDir(s.rootDir, req.Name) {
		slog.Error("dir does not exist", "name", req.Name)
		return nil, fmt.Errorf("the directory %s doesn't exist", p)
	}
	return &OpenDirResponse{
		Name: p,
	}, nil
}

func (s *MetaDataServer) DeleteAllData(ctx context.Context, req *DeleteAllDataRequest) (*DeleteAllDataReponse, error) {
	s.rootDir = &fileInfo{
		name:  ".",
		isDir: true,
	}
	return &DeleteAllDataReponse{}, nil
}

func (s *MetaDataServer) ReadDir(ctx context.Context, req *ReadDirRequest) (*FileInfo, error) {
	p := path.Clean(req.Name)

	info, err := getFileInfoAtIndex(s, p, int(req.Index))
	if err != nil {
		return nil, err
	}
	return convert(info), nil
}

func convert(f *fileInfo) *FileInfo {
	return &FileInfo{
		Name:     f.name,
		FullPath: f.fullPath,
		Size:     f.size,
		Mode:     int32(fs.ModePerm),
		ModTime:  time.Now().String(),
		IsDir:    f.isDir,
		Port:     int32(f.port),
	}
}

func (s *MetaDataServer) ReadDirAll(ctx context.Context, req *ReadDirRequest) (*ReadDirAllResponse, error) {
	p := path.Clean(req.Name)
	// check if this is a valid directory
	if !isDir(s.rootDir, req.Name) {
		slog.Error("dir does not exist")
		return nil, fmt.Errorf("the directory %s doesn't exist", p)
	}

	entries, err := getAllEntriesFromDir(s.rootDir, p)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	res := new(ReadDirAllResponse)
	for _, e := range entries {
		res.Entries = append(res.Entries, convert(e))
	}

	return res, nil
}
