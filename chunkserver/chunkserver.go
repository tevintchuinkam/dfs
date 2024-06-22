package chunksever

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// accept connections
	port := ":9000"
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	if err := grpcServer.Serve(lis); err != nil {
		// https://www.youtube.com/watch?v=BdzYdN_Zd9Q
	}

	// store the incoming files

	// return the requested files

}
