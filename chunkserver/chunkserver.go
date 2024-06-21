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
	s := grpc.NewServer()
	// add routes
	if err := s.Serve(lis); err != nil {

	}

	// store the incoming files

	// return the requested files

}
