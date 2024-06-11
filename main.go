package main

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"
)

const CHUNK_SIZE_BYTES = 1024

// implements fs.FS
var _ fs.FS = &TDFS{}

type ChunckServer struct {
	Port string
	ID   string
}

type TDFS struct {
	currentDir *directory
	rootDir    *directory
	server     Server
	chunksize  int
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
		chunk.ID = chunkID(uuid.New().String())
		for i := processed; i-processed < m.ChunkSize(); i++ {
			chunk.data = append(chunk.data, data[i])
		}
		file.chunks = append(file.chunks, chunk)
		// TODO:  send the chunk to one of the file servers
	}
	return nil
}

// ChunkSize returns the chunk size used
// in the file system
func (n *TDFS) ChunkSize() int {
	return n.chunksize
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
	return nil, fmt.Errorf("directory does not exist: %s", base)
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
		currentDir: root,
		rootDir:    root,
		server: Server{
			chunkServers: make(map[chunkID]ChunckServer),
		},
		chunksize: 1024,
	}
}
