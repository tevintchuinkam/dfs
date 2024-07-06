package client

import "github.com/tevintchuinkam/dfs/metadata"

type dirContents struct {
	// is this the entire directory or just parts of it ?
	full    bool
	entries []*metadata.FileInfo
}
type dirName string

// this client cache will the used for caching
// metadata as part of the metadata prefetching
// optimization
type ClientCache struct {
	dirs map[dirName]dirContents
}
