package chunks

import (
	"context"
)

type Server struct {
	UnimplementedChunkServiceServer
}

// type ChunkServiceServer interface {
// 	StoreChunk(context.Context, *StoreChunkRequest) (*StoreChunkResponse, error)
// 	GetChunk(*GetChunkRequest, ChunkService_GetChunkServer) error
// 	mustEmbedUnimplementedChunkServiceServer()
// }

func (s *Server) StoreChunk(ctx context.Context, in *StoreChunkRequest) (*StoreChunkResponse, error) {
	return &StoreChunkResponse{}, nil
}

func (s *Server) GetChunk(in *GetChunkRequest, src ChunkService_GetChunkServer) error {
	return nil
}
