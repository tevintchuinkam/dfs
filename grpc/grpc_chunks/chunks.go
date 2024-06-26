package grpc_chunks

import (
	"context"
)

type ChunkServer struct {
	UnimplementedChunkServiceServer
}

// type ChunkServiceServer interface {
// 	StoreChunk(context.Context, *StoreChunkRequest) (*StoreChunkResponse, error)
// 	GetChunk(*GetChunkRequest, ChunkService_GetChunkServer) error
// 	mustEmbedUnimplementedChunkServiceServer()
// }

func (s *ChunkServer) StoreChunk(ctx context.Context, in *StoreChunkRequest) (*StoreChunkResponse, error) {
	return &StoreChunkResponse{}, nil
}

func (s *ChunkServer) GetChunk(in *GetChunkRequest, src ChunkService_GetChunkServer) error {
	return nil
}
