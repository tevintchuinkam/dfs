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
	"sync"
	"time"

	"github.com/tevintchuinkam/tdfs/client"
	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/helpers"
	"github.com/tevintchuinkam/tdfs/metadata"
)

const (
	CLIENT_PORT       = 4999
	MDS_PORT          = 5000
	NUM_CHUNK_SERVERS = 1
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelError)
	log.SetFlags(log.Lshortfile)
	// create the metadata server
	mds := metadata.New(MDS_PORT)
	go mds.Start()
	slog.Info("mds started", "port", MDS_PORT)

	// create a few files servers
	var fsPorts []int
	for i := range NUM_CHUNK_SERVERS {
		fsPorts = append(fsPorts, MDS_PORT+i+1)
	}
	for _, port := range fsPorts {
		go files.New(port).Start()
	}
	time.Sleep(1 * time.Second)
	for _, port := range fsPorts {
		if err := mds.RegisterFileServer(port); err != nil {
			panic(err)
		}
		slog.Info("registered chunk server", "port", port)
	}
	/*
		f, err := os.Create("prof.prof")
		if err != nil {

			fmt.Println(err)
			return

		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/
	// flatOptimisation(c)
	// gatherWorkStealingOptimisationData()
	gatherGrepOptimizationData()

	// do a grep (with and without smart data proximity)
}

func createFilesAndDirs(c *client.Client, dir string, level int, data []byte, foldersPerLevel int) {
	const (
		CLIENT_PREFETCH_THRESHOLD = 8
		LEVELS                    = 3 // depth of the folders
		FILES_PER_LEVEL           = 3 // number of files in each folder
	)
	if level > LEVELS {
		return
	}

	for i := 0; i < foldersPerLevel; i++ {
		subDir := fmt.Sprintf("%s/dir-%d", dir, i+1)
		if err := c.MkDir(subDir); err != nil {
			log.Fatal(err)
		}

		// Create files in the current directory
		for j := 0; j < FILES_PER_LEVEL; j++ {
			filename := fmt.Sprintf("%s/file-%d.txt", subDir, j+1)
			r, err := c.CreateFile(filename, data)
			if err != nil {
				log.Fatal(err)
			}
			if r != len(data) {
				log.Fatalf("expected to write %d bytes but wrote %d bytes", len(data), r)
			}
		}

		// Recursive call to create subdirectories and files
		createFilesAndDirs(c, subDir, level+1, data, foldersPerLevel)
	}
}

func gatherGrepOptimizationData() {
	const CLIENT_PREFETCH_THRESHOLD = 8
	const NUM_ITERATIONS = 10
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	// create files
	data := data()

	// Open the CSV file
	csvFile, writer := openCSVFile("results/results_workstealing.csv", []string{"Algo", "Iteration", "Time Taken", "FoldersPerLevel"})
	defer csvFile.Close()

	type TraversalAlgo struct {
		traverse func(*client.Client, string, bool, func(*metadata.FileInfo)) error
		name     string
	}

	// do a file traversal (with and without metadata prefetching)
	useCache := false
	for foldersPerLevel := range 1 {
		c.DeleteAllData()
		createFilesAndDirs(c, ".", 1, data, foldersPerLevel+1)
		for i := range NUM_ITERATIONS {
			for _, algo := range [](TraversalAlgo){
				TraversalAlgo{
					traverse: traverseDirectorySimple,
					name:     "simple",
				},
				TraversalAlgo{
					traverse: traverseDirectoryWorkStealing,
					name:     "workstealing",
				},
			} {
				slog.Info("iteration", "count", i)
				c.ClearCache()
				start := time.Now()
				totalCount := new(int) // total count of the searched word
				mu := new(sync.Mutex)
				if err := algo.traverse(c, ".", useCache, func(file *metadata.FileInfo) {
					// fetch the file from the chunk server
					bytes, err := c.GetFileFromPort(file.Port, file.FullPath)
					if err != nil {
						log.Fatal(err)
					}
					word := "And"
					total := helpers.CountWordOccurrences(bytes, word)
					fmt.Printf("found %s %d times in file %s of size %d\n", word, total, file.FullPath, file.Size)
					mu.Lock()
					*totalCount += total
					mu.Unlock()

				}); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("total count of the word 'road': %d\n", *totalCount)
				took := time.Since(start)
				if err := writer.Write(
					[]string{
						algo.name,
						fmt.Sprint(i),
						took.String(),
						fmt.Sprint(foldersPerLevel),
					},
				); err != nil {
					log.Fatal(err)
				}
				writer.Flush()
			}
		}
	}
}

func gatherWorkStealingOptimisationData() {
	const CLIENT_PREFETCH_THRESHOLD = 8
	const NUM_ITERATIONS = 10
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	// create files
	data := data()

	// Open the CSV file
	csvFile, writer := openCSVFile("results/results_workstealing.csv", []string{"Algo", "Iteration", "Time Taken", "FoldersPerLevel"})
	defer csvFile.Close()

	type TraversalAlgo struct {
		traverse func(*client.Client, string, bool, func(*metadata.FileInfo)) error
		name     string
	}

	// do a file traversal (with and without metadata prefetching)
	useCache := false
	for foldersPerLevel := range 8 {
		c.DeleteAllData()
		createFilesAndDirs(c, ".", 1, data, foldersPerLevel+1)
		for i := range NUM_ITERATIONS {
			for _, algo := range [](TraversalAlgo){
				TraversalAlgo{
					traverse: traverseDirectorySimple,
					name:     "simple",
				},
				TraversalAlgo{
					traverse: traverseDirectoryWorkStealing,
					name:     "workstealing",
				},
			} {
				slog.Info("iteration", "count", i)
				c.ClearCache()
				start := time.Now()
				if err := algo.traverse(c, ".", useCache, func(file *metadata.FileInfo) { fmt.Println(file.FullPath) }); err != nil {
					log.Fatal(err)
				}
				took := time.Since(start)
				if err := writer.Write(
					[]string{
						algo.name,
						fmt.Sprint(i),
						took.String(),
						fmt.Sprint(foldersPerLevel),
					},
				); err != nil {
					log.Fatal(err)
				}
				writer.Flush()
			}
		}
	}
}

func traverseDirectoryWorkStealing(c *client.Client, dirPath string, useCache bool, f func(*metadata.FileInfo)) error {
	var wgWork sync.WaitGroup
	work := make(chan string, 100)
	defer close(work)
	// work <- workRes{[]string{dirPath}}
	work <- dirPath
	wgWork.Add(1)
	done := make(map[int]bool)
	var muDone sync.Mutex

	// Number of workers
	numWorkers := 8
	for i := range numWorkers {
		muDone.Lock()
		done[i] = false
		muDone.Unlock()
		go func(id int) {
			for w := range work {
				// Open the directory
				dir, err := c.OpenDir(w)
				if err != nil {
					slog.Error("could not open dir")
					wgWork.Done()
					continue
				}
				index := 0
				for {
					// Read the directory entry
					entry, err := c.ReadDir(dir, index, useCache)
					if err != nil {
						// Ugly but works for now
						switch {
						case strings.Contains(err.Error(), (metadata.EndOfDirectoryError{}).Error()):
							break
						case errors.Is(err, io.EOF):
							break
						default:
							slog.Error(err.Error())
						}
						break
					}
					if entry.IsDir {
						nextPath := filepath.Join(w, entry.Name)
						// adding more work
						wgWork.Add(1)
						work <- nextPath
					} else {
						f(entry)
					}
					index++
				}
				// we finished one piece of work
				wgWork.Done()
			}
		}(i)
	}
	// Wait for all goroutines to finish
	wgWork.Wait()
	return nil
}

func traverseDirectorySimple(c *client.Client, dirPath string, useCache bool, f func(*metadata.FileInfo)) error {
	// Open the directory
	dir, err := c.OpenDir(dirPath)
	if err != nil {
		slog.Error("could not open dir")
		return err
	}

	index := 0
	for {
		entry, err := c.ReadDir(dir, index, useCache)
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
			if err := traverseDirectorySimple(c, nextPath, useCache, f); err != nil {
				return err
			}
		} else {
			// Write the file info to the CSV file
			f(entry)
		}
		index++
	}
}

func gatherFlatOptimisationData() {
	const CLIENT_PREFETCH_THRESHOLD = 8
	const NUM_FOLDERS = 2
	const NUM_FILE = 30
	const NUM_ITERATIONS = 30
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	// create files
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

	// Open the CSV file
	csvFile, writer := openCSVFile("results/results.csv", []string{"Timestamp", "Time Taken", "Iteration", "UseCache"})
	defer csvFile.Close()

	// do a file traversal (with and without metadata prefetching)
	for _, useCache := range []bool{true, false} {
		for iteration := range NUM_ITERATIONS {
			// wait until cache is empty
			c.ClearCache()
			if err := computeReadDirTime(c, ".", useCache, writer, iteration); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Function to traverse the directory
func computeReadDirTime(c *client.Client, dirPath string, useCache bool, writer *csv.Writer, iteration int) error {
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
			if err := computeReadDirTime(c, nextPath, useCache, writer, iteration); err != nil {
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
func openCSVFile(filePath string, headers []string) (*os.File, *csv.Writer) {
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
		writer.Write(headers)
		writer.Flush()
	}

	return csvFile, writer
}

func data() []byte {
	d := []byte(`
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
	data := []byte{}
	for range 1000 {
		data = append(data, d...)
	}
	return data
}
