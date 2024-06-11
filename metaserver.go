package main

import (
	"fmt"
)

type Server struct {
	fileMap map[string][]chunkID
	// map from chunk to chunk server address
	chunkServers map[chunkID]ChunckServer
}

func (s Server) GetFile(filename string) (*File, error) {
	var fileChunks []chunkID
	fileChunks, ok := s.fileMap[filename]
	if !ok {
		return nil, fmt.Errorf("file %s does not exist", filename)
	}
	if len(fileChunks) < 1 {
		return nil, fmt.Errorf("file %s contains no chunks", filename)
	}
	file := new(File)
	for _, id := range fileChunks {
		chunk, err := s.GetChunk(id)
		if err != nil {
			return nil, err
		}
		file.chunks = append(file.chunks, &chunk)
	}
	return file, nil
}

func (s Server) GetChunk(id chunkID) (Chunck, error) {
	_ = s.chunkServers[id].Port
	return Chunck{}, nil
}
