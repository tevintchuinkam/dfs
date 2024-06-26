package main

import (
	"time"

	"github.com/tevintchuinkam/tdfs/chunks"
	"github.com/tevintchuinkam/tdfs/metadata"
)

const (
	META_DATA_SERVER_PORT = 5000
	NUM_CHUNK_SERVERS     = 10
	NUM_CLIENTS           = 10
)

func main() {
	// create the metadata server
	mds := metadata.New(META_DATA_SERVER_PORT)

	// create a few chunks servers
	for i := range NUM_CHUNK_SERVERS {
		port := META_DATA_SERVER_PORT + i + 1
		go chunks.New(port).Start()
		// wait until the server is up
		time.Sleep(1 * time.Second)
		if err := mds.RegisterChunkServer(port); err != nil {
			panic(err)
		}
		

	}

	// interact with the different clients to store and retreive some files
}
