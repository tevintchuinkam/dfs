package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
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

	"github.com/tevintchuinkam/dfs/client"
	"github.com/tevintchuinkam/dfs/files"
	"github.com/tevintchuinkam/dfs/grep"
	"github.com/tevintchuinkam/dfs/metadata"
)

var (
	MDS_PORT          = 47230
	CLIENT_PORT       = MDS_PORT - 1
	NUM_CHUNK_SERVERS = 8
)

var mds *metadata.MetaDataServer
var fileServers []*files.FileServer

func startAllServers(latency time.Duration) {
	// create the metadata server
	mds = metadata.New(MDS_PORT)
	go mds.Start(latency)
	slog.Info("mds started", "port", MDS_PORT)

	// create a few files servers
	var fsPorts []int
	for i := range NUM_CHUNK_SERVERS {
		fsPorts = append(fsPorts, MDS_PORT+i+1)
	}
	for _, port := range fsPorts {
		s := files.New(port)
		fileServers = append(fileServers, s)
		go s.Start(latency)
	}
	time.Sleep(1 * time.Second)
	for _, port := range fsPorts {
		if err := mds.RegisterFileServer(port); err != nil {
			panic(err)
		}
		slog.Info("registered chunk server", "port", port)
	}
}

func stopAllServers() {
	// create the metadata server
	if mds != nil {
		mds.DeleteAllData(context.Background(), &metadata.DeleteAllDataRequest{})
		mds.Stop()
	}
	for _, s := range fileServers {
		s.Stop()
	}

}

func main() {

	slog.SetLogLoggerLevel(slog.LevelError)
	log.SetFlags(log.Lshortfile)
	// Parse command-line flags for example  -iterations=1 -functions=flat
	iterations := flag.Int("iterations", 5, "number of iterations for the optimization data gathering")
	functions := flag.String("functions", "all", "comma-separated list of functions to run: flat, stealing, grep")

	flag.Parse()
	runFlat := strings.Contains(*functions, "flat") || *functions == "all"
	runStealing := strings.Contains(*functions, "stealing") || *functions == "all"
	runGrep := strings.Contains(*functions, "data_proximity") || *functions == "all"

	// Execute the functions based on the flags
	if runStealing {
		fmt.Printf("gather data for stealing optimization with a redundancy of %d iterations...\n", *iterations)
		gatherWorkStealingOptimisationData(*iterations)
	}
	if runFlat {
		fmt.Printf("gather data for flat optimization with a redundancy of %d iterations...\n", *iterations)
		gatherFlatOptimisationData(*iterations)
	}
	if runGrep {
		fmt.Printf("gather data for data_proximity optimization with a redundancy of %d iterations...\n", *iterations)
		gatherDataProximityOptimizationData(*iterations)
	}
}

func createFilesAndDirs(c *client.Client, dir string, level int, data []byte, filesPerFolder int, foldersPerLevel int, levels int) {

	var LEVELS = levels
	var FOLDER_PER_LEVEL = foldersPerLevel
	var FILES_PER_FOLDER = filesPerFolder
	if level > LEVELS {
		return
	}

	for i := 0; i < FOLDER_PER_LEVEL; i++ {
		subDir := fmt.Sprintf("%s/dir-%d", dir, i+1)
		if err := c.MkDir(subDir); err != nil {
			log.Fatal(err)
		}

		// Create files in the current directory
		for j := 0; j < FILES_PER_FOLDER; j++ {
			filename := fmt.Sprintf("%s/file-%d.txt", subDir, j+1)
			r, err := c.CreateFileWithStream(filename, data)
			if err != nil {
				log.Fatal(err)
			}
			if r != len(data) {
				log.Fatalf("expected to write %d bytes but wrote %d bytes", len(data), r)
			}
		}

		// Recursive call to create subdirectories and files
		createFilesAndDirs(c, subDir, level+1, data, filesPerFolder, foldersPerLevel, levels)
	}
}

func gatherDataProximityOptimizationData(iterations int) {
	const CLIENT_PREFETCH_THRESHOLD = 8
	var NUM_ITERATIONS = iterations
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	c.ClearCache()

	// Open the CSV file
	csvFile, writer := openCSVFile("results/data_proximity.csv", []string{"Iteration", "Time Taken", "FileSizeMB", "DataProximity", "FileCount"})
	defer csvFile.Close()

	useCache := false
	fileSizeMB := 10
	for _, dataProximity := range []bool{true, false} {
		for filesPerFolder := range 16 {
			for i := range NUM_ITERATIONS {
				stopAllServers()
				startAllServers(0)
				createFilesAndDirs(c, ".", 1, generateData(fileSizeMB), filesPerFolder+1, 1, 1)
				c.ClearCache()
				totalCount := new(int) // total count of the searched word
				mu := new(sync.Mutex)
				word := "And"
				var f func(*metadata.FileInfo, *sync.WaitGroup)
				if !dataProximity {
					f = func(file *metadata.FileInfo, wg *sync.WaitGroup) {
						defer wg.Done()
						// fetch the file from the chunk server
						//start := time.Now()
						bytes, err := c.GetFileFromPortWithStream(file.Port, file.FullPath)
						if err != nil {
							log.Fatal(err)
						}
						//took := time.Since(start)
						//start = time.Now()
						total := grep.CountWordOccurrences(bytes, word)
						// fmt.Printf("OFF found %s %d times in file %s of size %dMB fetching_took=%v counting_took%v\n", word, total, file.FullPath, fileSizeMB, took, time.Since(start))
						mu.Lock()
						*totalCount += total
						mu.Unlock()
					}
				} else {
					fmt.Println("using grep on file server func")
					f = func(file *metadata.FileInfo, wg *sync.WaitGroup) {
						defer wg.Done()
						go func() {
							// compute the count on the fileserver and reduce on the client
							// start := time.Now()
							count, err := c.GrepOnFileServer(file.FullPath, word, file.Port)
							if err != nil {
								log.Fatal(err)
							}
							// fmt.Printf("ON found %s %d times in file %s of size %dMB took=%v\n", word, count, file.FullPath, fileSizeMB, time.Since(start))
							mu.Lock()
							*totalCount += count
							mu.Unlock()
						}()
					}
				}
				start := time.Now()
				if err := traverseDirectoryWorkStealing(c, ".", useCache, f); err != nil {
					log.Fatal(err)
				}
				took := time.Since(start)
				if err := writer.Write(
					[]string{
						fmt.Sprint(i),
						took.String(),
						fmt.Sprint(fileSizeMB),
						fmt.Sprint(dataProximity),
						fmt.Sprint(filesPerFolder),
					},
				); err != nil {
					log.Fatal(err)
				}
				writer.Flush()
			}
		}
	}
}

func gatherWorkStealingOptimisationData(iterations int) {
	const CLIENT_PREFETCH_THRESHOLD = 8
	var NUM_ITERATIONS = iterations
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	c.ClearCache()
	// create files
	data := generateData(1)

	// Open the CSV file
	csvFile, writer := openCSVFile("results/workstealing.csv", []string{"Algo", "Iteration", "Time Taken", "LatencyMS"})
	defer csvFile.Close()

	type TraversalAlgo struct {
		traverse func(*client.Client, string, bool, func(*metadata.FileInfo, *sync.WaitGroup)) error
		name     string
	}

	// do a file traversal (with and without metadata prefetching)
	useCache := false
	foldersPerLevel := 5
	for latency := range 20 {
		for i := range NUM_ITERATIONS {
			stopAllServers()
			startAllServers(time.Duration(latency) * time.Millisecond)
			createFilesAndDirs(c, ".", 1, data, 10, foldersPerLevel, 2)
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
				if err := algo.traverse(c, ".", useCache, func(file *metadata.FileInfo, _ *sync.WaitGroup) { fmt.Println(file.FullPath) }); err != nil {
					log.Fatal(err)
				}
				took := time.Since(start)
				if err := writer.Write(
					[]string{
						algo.name,
						fmt.Sprint(i),
						took.String(),
						fmt.Sprint(latency),
					},
				); err != nil {
					log.Fatal(err)
				}
				writer.Flush()
			}
		}
	}
}

func traverseDirectoryWorkStealing(c *client.Client, dirPath string, useCache bool, f func(*metadata.FileInfo, *sync.WaitGroup)) error {
	var wgWork sync.WaitGroup
	work := make(chan string, 200)
	defer close(work)
	work <- dirPath
	wgWork.Add(1)

	var wgFunc sync.WaitGroup

	// Number of workers
	numWorkers := 8
	for i := range numWorkers {
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
						wgFunc.Add(1)
						f(entry, &wgFunc)
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
	wgFunc.Wait()
	return nil
}

func traverseDirectorySimple(c *client.Client, dirPath string, useCache bool, f func(*metadata.FileInfo, *sync.WaitGroup)) error {
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
			f(entry, new(sync.WaitGroup))
		}
		index++
	}
}

func gatherFlatOptimisationData(iterations int) {
	const CLIENT_PREFETCH_THRESHOLD = 8
	const NUM_FOLDERS = 2
	const NUM_FILE = 30
	var NUM_ITERATIONS = iterations
	c := client.New(MDS_PORT, CLIENT_PREFETCH_THRESHOLD)
	c.ClearCache()
	stopAllServers()
	startAllServers(0)
	// create files
	data := generateData(1)
	createFiles := func() {
		for dirNum := range NUM_FOLDERS {
			files := []string{}
			for i := range NUM_FILE {
				files = append(files, fmt.Sprintf("dir-%d/file-%d.txt", dirNum+1, i+1))
			}
			for _, filename := range files {
				if err := c.MkDir(path.Dir(filename)); err != nil {
					log.Fatal(err)
				}
				r, err := c.CreateFileWithStream(filename, data)
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

	// Open the CSV file
	csvFile, writer := openCSVFile("results/flat.csv", []string{"Timestamp", "Time Taken", "Iteration", "UseCache"})
	defer csvFile.Close()

	// do a file traversal (with and without metadata prefetching)
	for _, useCache := range []bool{true, false} {
		for iteration := range NUM_ITERATIONS {
			// wait until cache is empty
			stopAllServers()
			startAllServers(0)
			c.ClearCache()
			createFiles()
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

func generateData(mb int) []byte {
	// Define the base data
	baseData := []byte(`
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

	// Calculate the total number of bytes needed
	totalBytes := mb * 1024 * 1024

	// Create a buffer to hold the result
	data := make([]byte, 0, totalBytes)

	// Repeat the base data until the required size is reached
	for len(data) < totalBytes {
		if len(data)+len(baseData) > totalBytes {
			data = append(data, baseData[:totalBytes-len(data)]...)
		} else {
			data = append(data, baseData...)
		}
	}

	return data
}
