package helpers

import (
	"fmt"
	"log"

	"github.com/tevintchuinkam/dfs/files"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewFileServiceClient(port int32) files.FileServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect. err: %v", err)
	}
	return files.NewFileServiceClient(conn)
}
