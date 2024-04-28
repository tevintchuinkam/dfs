package main

import (
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

// implements fs.FS
var _ fs.FS = &TDFS{}

type Chunck struct {
	Mu     sync.Mutex
	Handle string
	Size   int
}

type ChunckServer struct {
	Port string
	ID   string
}

type File struct {
	chunks []Chunck
}

type FileInfo struct {
}

func (f *FileInfo) Name() string {
	return "file"
}
func (f *FileInfo) Size() int64 {
	return 0
}
func (f *FileInfo) Mode() fs.FileMode {
	return 0
}
func (f *FileInfo) ModTime() time.Time {
	return time.Now()
}
func (f *FileInfo) IsDir() bool {
	return false
}

func (f *FileInfo) Sys() any {
	return struct{}{}
}

// File implements fs.File
func (f *File) Close() error {
	return nil
}
func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{}, nil
}

func (f *File) Read(b []byte) (int, error) {
	return 0, nil
}

type Master struct {
	FileToHandlesMap  map[string]File
	HandleToServerMap map[string]ChunckServer
}

type TDFS struct {
	BaseDir string
	Master  *Master
}

// implements fs.FS
func (t *TDFS) Open(name string) (fs.File, error) {
	return &File{
		chunks: []Chunck{},
	}, nil
}

type Server interface {
}

func newMaster() *Master {
	return &Master{
		FileToHandlesMap:  make(map[string]File),
		HandleToServerMap: make(map[string]ChunckServer),
	}
}

func (m *TDFS) CreateFile(name, dir string, data []byte) error {
	return nil
}
func (m *TDFS) ReadFile(name, dir string) ([]byte, error) {
	return []byte{}, nil
}
func (m *TDFS) DeleteFile(name, dir string) error {
	return nil
}
func (m *TDFS) CreateDir(dir string) error {
	return nil
}
func (m *TDFS) DeleteDir(dir string) error {
	return nil
}
func (m *TDFS) LS(dir string) error {
	return nil
}
func (m *TDFS) Grep(exp string) error {
	return nil
}

// creates a new TDFS filesystem with the given base directory
// panics if the given path is invalid
// if not dir is provided, the current working directory is used
// as the base directory
func New(dir ...string) *TDFS {
	d := "."
	if len(dir) > 0 {
		d = dir[0]
	}
	absPath, err := filepath.Abs(d)
	if err != nil {
		panic(err)
	}
	return &TDFS{
		BaseDir: absPath,
		Master:  newMaster(),
	}
}
