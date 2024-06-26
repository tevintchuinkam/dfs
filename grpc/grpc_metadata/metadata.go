package grpc_metadata

import (
	"context"
	"log/slog"
)

type Server struct {
	UnimplementedMetadataServiceServer
}

func (s *Server) GetMetadata(ctx context.Context, in *MetadataRequest) (*MetadataResponse, error) {
	slog.Info("received chunk data request", "filename", in.Filename, "chunk_index", in.ChunkIndex)
	return &MetadataResponse{
		ChunkHandle: "abcde",
		Url:         "chunk:9000",
	}, nil
}
