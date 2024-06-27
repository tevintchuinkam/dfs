package main

import (
	"io/fs"

	"github.com/tevintchuinkam/tdfs/client"
	"github.com/tevintchuinkam/tdfs/metadata"
)

var _ fs.FS = (*DFS)(nil)
var _ fs.ReadDirFS = (*DFS)(nil)
var _ fs.StatFS = (*DFS)(nil)

var _ fs.File = (*File)(nil)

var _ fs.DirEntry = (*DirEntry)(nil)

type DFS struct {
	MDS     *metadata.MetaDataServer
	Clients []*client.ClientServiceClient
}
type File struct{}
type DirEntry struct{}

// dfs
func (f *DFS) Open(name string) (fs.File, error) {
	return &File{}, nil
}
func (f *DFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries := []*DirEntry{
		&DirEntry{}, // Add actual DirEntry initialization here
	}

	// Convert []*DirEntry to []fs.DirEntry
	var result []fs.DirEntry
	for _, entry := range entries {
		result = append(result, entry)
	}
	return result, nil
}
func (f *DFS) Stat(name string) (fs.FileInfo, error)

// file
func (f *File) Stat() (fs.FileInfo, error)
func (f *File) Read([]byte) (int, error)
func (f *File) Close() error

// dir entry
func (e *DirEntry) Name() string
func (e *DirEntry) IsDir() bool
func (e *DirEntry) Type() fs.FileMode
func (e *DirEntry) Info() (fs.FileInfo, error)
