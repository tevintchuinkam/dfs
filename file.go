package main

import (
	"io/fs"
	"sync"
	"time"
)

type chunkID string

type Chunck struct {
	ID   chunkID
	Mu   sync.Mutex
	data []byte
}

type File struct {
	name   string
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
