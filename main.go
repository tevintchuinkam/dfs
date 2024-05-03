package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// implements fs.FS
var _ fs.FS = &TDFS{}

type Chunck struct {
	ID   string
	Mu   sync.Mutex
	data []byte
}

type ChunckServer struct {
	Port string
	ID   string
}

type File struct {
	chunks []*Chunck
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

type directory struct {
	name    string
	prev    *directory
	subDirs []*directory
	files   map[string]*File
}

func (d *directory) String() string {
	if d == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(d.name)
	for fileName, _ := range d.files {
		sb.WriteString(fmt.Sprintf("\n\t├── %s", fileName))
	}
	return sb.String()
}

type TDFS struct {
	chunkSizeBytes int
	currentDir     *directory
	rootDir        *directory
	chunkServers   map[string]ChunckServer
}

// implements fs.FS
func (t *TDFS) Open(name string) (fs.File, error) {
	return &File{
		chunks: []*Chunck{},
	}, nil
}

func (m *TDFS) CreateFile(name, dir string, data []byte) error {
	d, err := m.rootDir.WalkTo(dir)
	if err != nil {
		return err
	}
	if _, ok := d.files[name]; ok {
		return fmt.Errorf("file %s already exists", name)
	}
	file := new(File)
	processed := 0
	for len(data) < processed {
		chunk := new(Chunck)
		chunk.ID = uuid.New().String()
		for i := processed; i-processed < m.chunkSizeBytes; i++ {
			chunk.data = append(chunk.data, data[i])
		}
		file.chunks = append(file.chunks, chunk)
		// TODO:  send the chunk to one of the file servers

	}
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

func (d *directory) WalkTo(p string) (*directory, error) {
	base := path.Base(p)
	rest := strings.TrimPrefix(p, base)
	for _, dir := range d.subDirs {
		if base == dir.name {
			return dir.WalkTo(rest)
		}
	}
	return nil, errors.New(fmt.Sprintf("directory does not exist: %s", base))
}

// creates a new TDFS filesystem with the given base directory
func New() *TDFS {
	root := &directory{
		name:    "",
		prev:    nil,
		subDirs: []*directory{},
		files:   make(map[string]*File),
	}
	return &TDFS{
		chunkSizeBytes: 64,
		currentDir:     root,
		rootDir:        root,
		chunkServers:   make(map[string]ChunckServer),
	}
}
