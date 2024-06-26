package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tevintchuinkam/tdfs/client"
	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/metadata"
)

const (
	CLIENT_PORT           = 4999
	META_DATA_SERVER_PORT = 5000
	NUM_CHUNK_SERVERS     = 10
	NUM_CLIENTS           = 10
)

func main() {
	// create the metadata server
	mds := metadata.New(META_DATA_SERVER_PORT)

	// create a few files servers
	var port int
	for i := range NUM_CHUNK_SERVERS {
		port = META_DATA_SERVER_PORT + i + 1
		go files.New(port).Start()
		// wait until the server is up
		time.Sleep(2 * time.Second)
		if err := mds.RegisterFileServer(port); err != nil {
			panic(err)
		}
	}

	var clients []*client.Client
	for range NUM_CLIENTS {
		port++
		c := client.New(port, META_DATA_SERVER_PORT)
		clients = append(clients, c)
		go c.Start()
	}

	// wait for all the client to start
	time.Sleep(2 * time.Second)

	data := []byte(`
		The Road Not Taken

		Two roads diverged in a yellow wood,
		And sorry I could not travel both
		And be one traveler, long I stood
		And looked down one as far as I could
		To where it bent in the undergrowth;

		Then took the other, as just as fair,
		And having perhaps the better claim,
		Because it was grassy and wanted wear;
		Though as for that the passing there
		Had worn them really about the same,

		And both that morning equally lay
		In leaves no step had trodden black.
		Oh, I kept the first for another day!
		Yet knowing how way leads on to way,
		I doubted if I should ever come back.

		I shall be telling this with a sigh
		Somewhere ages and ages hence:
		Two roads diverged in a wood, and I—
		I took the one less traveled by,
		And that has made all the difference.

		- Robert Frost
	`)

	fmt.Print(len(data))

	c := clients[0]
	for i := range 3 {
		dir := fmt.Sprintf("dir-%d", i)
		resp, err := c.MkDir(context.Background(), &client.MkDirRequest{
			Dir: dir,
		})
		if resp.Depth != 1 {
			log.Fatal("depth should be 1")
		}
		if err != nil {
			log.Fatal(err)
		}
		for j := range 3 {
			name := fmt.Sprintf("file-%d-%d", i, j)
			c.CreateFile(context.Background(), &client.CreateFileRequest{
				Name: name,
				Dir:  dir,
				Data: data,
			})
		}
	}

	resp, err := c.LS(context.Background(), &client.LSRequest{
		Dir: "/",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("dirs: %v", resp.Dirs)
	log.Printf("files: %v", resp.Files)

	//  TODO GetFile, grep ...

	// interact with the different clients to store and retreive some files
}
