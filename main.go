package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"path"
	"path/filepath"
	"time"

	"github.com/tevintchuinkam/tdfs/client"
	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/metadata"
)

const (
	LATENCY           = 3 * time.Microsecond
	CLIENT_PORT       = 4999
	MDS_PORT          = 5000
	NUM_CHUNK_SERVERS = 10
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	log.SetFlags(log.Lshortfile)
	// create the metadata server
	mds := metadata.New(MDS_PORT)
	go mds.Start(LATENCY)
	slog.Info("mds started", "port", MDS_PORT)

	// create a few files servers
	var fsPorts []int
	for i := range NUM_CHUNK_SERVERS {
		fsPorts = append(fsPorts, MDS_PORT+i+1)
	}
	for _, port := range fsPorts {
		go files.New(port).Start(LATENCY)
	}
	time.Sleep(1 * time.Second)
	for _, port := range fsPorts {
		if err := mds.RegisterFileServer(port); err != nil {
			panic(err)
		}
		slog.Info("registered chunk server", "port", port)
	}

	data := data()

	fmt.Printf("data size: %d bytes\n", len(data))
	// interact with the different clients to store and retreive some files

	c := client.New(MDS_PORT)
	// create files
	files := []string{}
	for i := range 20 {
		files = append(files, fmt.Sprintf("somedir/file-%d.txt", i+1))
	}
	for _, filename := range files {
		if err := c.MkDir(path.Dir(filename)); err != nil {
			log.Fatal(err)
		}
		r, err := c.CreateFile(filename, data)
		if err != nil {
			log.Fatal(err)
		}
		if r != len(data) {
			log.Fatalf("expected to write %d bytes but wrote %d bytes", len(data), len(data))
		}
	}

	//  retrieve the files
	for _, filename := range files {
		bytes, err := c.GetFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		if len(data) != len(bytes) {
			log.Fatalf("wrote %d bytes but only retrieved %d", len(data), len(bytes))
		}
	}

	// do a file traversal (with and without metadata prefetching)

	// do a grep (with and without smart data proximity)

}

// Function to traverse the directory
func traverseDirectory(c *client.Client, dirPath string) error {
	// Open the directory
	dir, err := c.OpenDir(dirPath)
	if err != nil {
		return err
	}
	index := 0
	for {
		// Read the directory enntry
		entry, err := c.ReadDir(dir, index, false)
		if err != nil {
			if err == io.EOF {
				break // End of directory
			}
			return err
		}

		// Iterate through the directory entries
		// Print the entry name
		fmt.Println("Found entry:", entry.Name)

		// Check if the entry is a directory
		if entry.IsDir {
			// If it's a directory, recursively traverse it
			err = traverseDirectory(c, filepath.Join(dirPath, entry.Name))
			if err != nil {
				return err
			}
		} else {
			// If it's a file, perform actions on the file (e.g., print file info)
			fmt.Printf("File: %s, Size: %d bytes\n", entry.Name, entry.Size)
		}
		index++
	}
	return nil
}

func data() []byte {
	return []byte(`
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
}
