package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/tevintchuinkam/dfs/files"
	"github.com/tevintchuinkam/dfs/helpers"
	"github.com/tevintchuinkam/dfs/metadata"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type accessData struct {
	lastAccess  time.Time
	accessCount int
}

type Client struct {
	mdsPort           int
	prefetchThreshold int

	cache *ClientCache
	// dirname: unique access count
	predictionHistory map[string]*accessData
}

func New(mdsPort int, pefetchThreshold int) *Client {

	dirs := make(map[dirName]dirContents)
	// ping the server
	return &Client{
		mdsPort: mdsPort,
		cache: &ClientCache{
			dirs: dirs,
		},
		predictionHistory: make(map[string]*accessData),
		prefetchThreshold: pefetchThreshold,
	}
}

func (c *Client) DeleteAllData() {
	m := NewMDSClient(c.mdsPort)
	_, err := m.DeleteAllData(context.Background(), &metadata.DeleteAllDataRequest{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Client) ClearCache() {
	dirs := make(map[dirName]dirContents)
	c.cache = &ClientCache{
		dirs: dirs,
	}
	c.predictionHistory = make(map[string]*accessData)
}

func (c *Client) ReadDir(name string, index int, useCache bool) (*metadata.FileInfo, error) {
	if useCache {
		// try finding the dirs
		dir, ok := c.cache.dirs[dirName(name)]
		if ok {
			if dir.full {
				if index < 0 || index >= len(dir.entries) {
					return nil, io.EOF
				}
				return dir.entries[index], nil
			}
		}

		// cache the metadata
		if h, ok := c.predictionHistory[name]; ok {
			if h.lastAccess.Before(time.Now().Add(-1 * time.Second)) {
				h.accessCount = 0
			} else {
				if h.accessCount >= c.prefetchThreshold-1 {
					// this is the trigger to prefetch the entire dir from mds
					// slog.Info("prefetching directory", "dir", name, "trigger_index", index)
					entries, err := _prefetchDir(c, name)
					if err != nil {
						log.Fatalf("could not prefetch dir %s err: %v", name, err)
					}
					c.cache.dirs[dirName(name)] = dirContents{
						full:    true,
						entries: entries,
					}
					if index < 0 || index >= len(entries) {
						return nil, io.EOF
					}
					return entries[index], nil
				}
			}
			h.accessCount++
			h.lastAccess = time.Now()
		} else {
			c.predictionHistory[name] = &accessData{
				lastAccess:  time.Now(),
				accessCount: 0,
			}
		}
	}

	// request the file from mds, no caching
	m := NewMDSClient(c.mdsPort)
	r, err := m.ReadDir(context.Background(), &metadata.ReadDirRequest{
		Name:  name,
		Index: int32(index),
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func _prefetchDir(c *Client, name string) ([]*metadata.FileInfo, error) {
	mds := NewMDSClient(c.mdsPort)
	r, err := mds.ReadDirAll(context.Background(), &metadata.ReadDirRequest{
		Name: name,
	})
	if err != nil {
		log.Fatalf("could not prefetch dir %s err : %v", name, err)
	}
	return r.Entries, nil
}

// open a directory
func (c *Client) OpenDir(name string) (string, error) {
	m := NewMDSClient(c.mdsPort)
	r, err := m.OpenDir(context.Background(), &metadata.OpenDirRequest{
		Name: name,
	})
	if err != nil {
		slog.Error("could not open dir", "err", err.Error())
		return "", err
	}
	return r.Name, nil
}

func (c *Client) CreateFile(name string, data []byte) (int, error) {
	// ask the mds on on what storage server to store the file
	mds := NewMDSClient(c.mdsPort)
	rec, err := mds.RegisterFileCreation(context.Background(), &metadata.RecRequest{
		Name:     name,
		FileSize: int64(len(data)),
	})
	if err != nil {
		slog.Error("getting storage location recommendation failed", "err", err.Error())
		return 0, err
	}
	fs := helpers.NewFileServiceClient(rec.Port)
	fr, err := fs.CreateFile(context.Background(), &files.CreateFileRequest{
		Name: name,
		Data: data,
	})
	if err != nil {
		slog.Error("creating file failed", "err", err.Error())
		return 0, err
	}

	// send a createfilerequest to that server
	return int(fr.BytesWritten), nil
}

func (c *Client) MkDir(name string) error {
	mds := NewMDSClient(c.mdsPort)
	_, err := mds.MkDir(context.Background(), &metadata.MkDirRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (c *Client) GetFile(name string) ([]byte, error) {
	mds := NewMDSClient(c.mdsPort)
	loc, err := mds.GetLocation(context.Background(), &metadata.LocRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	fs := helpers.NewFileServiceClient(loc.Port)
	fr, err := fs.GetFile(context.Background(), &files.GetFileRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return fr.Data, nil
}

func (c *Client) GetFileFromPort(port int32, name string) ([]byte, error) {
	fs := helpers.NewFileServiceClient(port)
	fr, err := fs.GetFile(context.Background(), &files.GetFileRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return fr.Data, nil
}

func (c *Client) GetFileFromPortWithStream(port int32, name string) ([]byte, error) {
	fs := helpers.NewFileServiceClient(port)
	stream, err := fs.GetFileWithStream(context.Background(), &files.GetFileWithStreamRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	var data []byte
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed to receive chunk", "err", err.Error())
			return nil, err
		}

		data = append(data, res.GetChunkData()...)
	}
	return data, nil
}

func (c *Client) GrepOnFileServer(fileName string, word string, port int32) (int, error) {
	fs := helpers.NewFileServiceClient(port)
	r, err := fs.Grep(context.Background(), &files.GrepRequest{
		FileName: fileName,
		Word:     word,
	})
	if err != nil {
		return 0, err
	}
	return int(r.Count), nil
}

func NewMDSClient(port int) metadata.MetadataServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect. err: %v", err)
	}
	return metadata.NewMetadataServiceClient(conn)
}

func (c *Client) CreateFileWithStream(name string, data []byte) (int, error) {
	mds := NewMDSClient(c.mdsPort)
	rec, err := mds.RegisterFileCreation(context.Background(), &metadata.RecRequest{
		Name:     name,
		FileSize: int64(len(data)),
	})
	if err != nil {
		slog.Error("getting storage location recommendation failed", "err", err.Error())
		return 0, err
	}

	fs := helpers.NewFileServiceClient(rec.Port)
	stream, err := fs.CreateFileWithStream(context.Background())
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	err = stream.Send(&files.CreateFileWithStreamRequest{
		Data: &files.CreateFileWithStreamRequest_Info{
			Info: &files.FileInfo{
				Name: name,
			},
		},
	})
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}

	reader := bytes.NewReader(data)
	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error(err.Error())
			return 0, err
		}
		err = stream.Send(
			&files.CreateFileWithStreamRequest{
				Data: &files.CreateFileWithStreamRequest_ChunkData{
					ChunkData: buf[:n],
				},
			},
		)
		if err != nil {
			slog.Error(err.Error())
			return 0, err
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	slog.Info("file uploaded", "size", res.GetBytesWritten())
	return int(res.BytesWritten), nil
}

func (c *Client) GetFileWithStream(name string) ([]byte, error) {
	mds := NewMDSClient(c.mdsPort)
	loc, err := mds.GetLocation(context.Background(), &metadata.LocRequest{
		Name: name,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	data, err := c.GetFileFromPort(loc.Port, name)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	slog.Info("file downloaded", "size", len(data))
	return data, nil
}
