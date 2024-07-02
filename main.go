package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/tevintchuinkam/tdfs/client"
	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/metadata"
)

const (
	ADDED_LATENCY             = 0
	CLIENT_PORT               = 4999
	MDS_PORT                  = 5000
	NUM_CHUNK_SERVERS         = 10
	CLIENT_PREFETCH_THRESHOLD = 8
	NUM_FOLDERS               = 2
	NUM_FILE                  = 30
	NUM_ITERATIONS            = 30
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	log.SetFlags(log.Lshortfile)
	// create the metadata server
	mds := metadata.New(MDS_PORT)
	go mds.Start(ADDED_LATENCY)
	slog.Info("mds started", "port", MDS_PORT)

	// create a few files servers
	var fsPorts []int
	for i := range NUM_CHUNK_SERVERS {
		fsPorts = append(fsPorts, MDS_PORT+i+1)
	}
	for _, port := range fsPorts {
		go files.New(port).Start(ADDED_LATENCY)
	}
	time.Sleep(1 * time.Second)
	for _, port := range fsPorts {
		if err := mds.RegisterFileServer(port); err != nil {
			panic(err)
		}
		slog.Info("registered chunk server", "port", port)
	}

	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	// create files
	createFoldersAndFile(c)

	// Open the CSV file
	csvFile, writer := openCSVFile("results/results.csv")
	defer csvFile.Close()

	// do a file traversal (with and without metadata prefetching)
	for _, useCache := range []bool{true, false} {
		for iteration := range NUM_ITERATIONS {
			// wait until cache is empty
			c.ClearCache()
			if err := traverseDirectory(c, ".", useCache, writer, iteration); err != nil {
				log.Fatal(err)
			}
		}
	}
	// do a grep (with and without smart data proximity)

}

// Function to traverse the directory
func traverseDirectory(c *client.Client, dirPath string, useCache bool, writer *csv.Writer, iteration int) error {
	// Open the directory
	dir, err := c.OpenDir(dirPath)
	if err != nil {
		slog.Error("could not open dir")
		return err
	}

	index := 0
	for {
		// Read the directory entry
		start := time.Now()
		entry, err := c.ReadDir(dir, index, useCache)
		took := time.Since(start)
		if err != nil {
			// ugly but works for now
			switch {
			case strings.Contains(err.Error(), (metadata.EndOfDirectoryError{}).Error()):
				return nil
			case errors.Is(err, io.EOF):
				return nil
			}
			slog.Error(err.Error())
			return err
		}

		// Check if the entry is a directory
		if entry.IsDir {
			// If it's a directory, recursively traverse it
			nextPath := filepath.Join(dirPath, entry.Name)
			if err := traverseDirectory(c, nextPath, useCache, writer, iteration); err != nil {
				return err
			}
		} else {
			// Write the file info to the CSV file
			writer.Write([]string{
				fmt.Sprint(time.Now().UnixNano()),
				fmt.Sprint(took.Nanoseconds()),
				fmt.Sprint(iteration),
				fmt.Sprintf("%t", useCache),
			})
			writer.Flush()
		}
		index++
	}
}

// Function to open the CSV file
func openCSVFile(filePath string) (*os.File, *csv.Writer) {
	csvFile, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	writer := csv.NewWriter(csvFile)

	// Write the header if the file is newly created
	fileInfo, err := csvFile.Stat()
	if err != nil {
		return nil, nil
	}
	if fileInfo.Size() == 0 {
		writer.Write([]string{"Timestamp", "Time Taken", "Iteration", "UseCache"})
		writer.Flush()
	}

	return csvFile, writer
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
func createFoldersAndFile(c *client.Client) {
	data := data()
	for dirNum := range NUM_FOLDERS {
		files := []string{}
		for i := range NUM_FILE {
			files = append(files, fmt.Sprintf("dir-%d/file-%d.txt", dirNum+1, i+1))
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
	}

}
