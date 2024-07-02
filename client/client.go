package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/tevintchuinkam/tdfs/files"
	"github.com/tevintchuinkam/tdfs/helpers"
	"github.com/tevintchuinkam/tdfs/metadata"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Client struct {
	mdsPort int

	cache *ClientCache
	// dirname: unique access count
	predictionHistory map[string]*struct {
		lastAccess                       time.Time
		accessCountInLast200MilliSeconds int
	}
}

func New(mdsPort int) *Client {
	// ping the server
	return &Client{
		mdsPort: mdsPort,
	}
}

func (c *Client) ReadDir(name string, index int, useCache bool) (*metadata.FileInfo, error) {
	if useCache {
		// try finding the dirs
		dir, ok := c.cache.dirs[dirName(name)]
		if !ok {
			slog.Warn("directory not found in cache", "dir_name", name)
		} else {
			if dir.full {
				if index < 0 || index > len(dir.entries) {
					return nil, io.EOF
				}
				return dir.entries[index], nil
			}
		}
	}

	// if use cache is set, cache the metadata
	if useCache {
		if h, ok := c.predictionHistory[name]; ok {
			if h.lastAccess.Before(time.Now().Add(-200 * time.Millisecond)) {
				h.accessCountInLast200MilliSeconds = 0
			} else {
				if h.accessCountInLast200MilliSeconds > 10 {
					// 10 files requested from this dir in the last 200ms
					// this is the trigger to prefetch the entire dir from mds
					slog.Info("prefetching directory", "dir", name)
					entries, err := _prefetchDir(c, name)
					if err != nil {
						log.Fatalf("cound not prefetch dir %s err: %v", name, err)
					}
					c.cache.dirs[dirName(name)] = dirContents{
						full:    true,
						entries: entries,
					}
					return entries[index], nil
				}
			}
			h.accessCountInLast200MilliSeconds++
			h.lastAccess = time.Now()
		} else {
			c.predictionHistory[name] = &struct {
				lastAccess                       time.Time
				accessCountInLast200MilliSeconds int
			}{
				lastAccess:                       time.Now(),
				accessCountInLast200MilliSeconds: 0,
			}
		}
	}
	// request the file from mds, no caching
	m := newMDSClient(c.mdsPort)
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
	mds := newMDSClient(c.mdsPort)
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
	m := newMDSClient(c.mdsPort)
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
	mds := newMDSClient(c.mdsPort)
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
	mds := newMDSClient(c.mdsPort)
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
	mds := newMDSClient(c.mdsPort)
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

func newMDSClient(port int) metadata.MetadataServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect. err: %v", err)
	}
	return metadata.NewMetadataServiceClient(conn)
}
