package main

import "github.com/tevintchuinkam/tdfs/chunks"

const (
	NUM_CHUNK_SERVERS = 10
	NUM_CLIENTS       = 10
)

func main() {
	// create a few chunks servers
	for i := range NUM_CHUNK_SERVERS {
		go chunks.New(5000 + i).Start()
	}

	

	// create the metadata server

	// interact with the different clients to store and retreive some files
}
