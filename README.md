# TDFS

TDFS (Tevin's Distributed File System) is a concurrency-proof distributed filesystem that is compatible with go's fs.FS interface.
It implements ideas from the Lustre Metadata Prefetching Algorithm and the POSH Smart Data Proximity concepts. This was created as part of a seminar thesis about Large Filesystem Traversal Algorithms that I wrote at RWTH Aachen Univeristy.

## Design

TDFS is designed to have one master who keeps track of file chucks locations in memory. The file chunks are stored in chunk servers.

### Master

### Chunk Server

### Chunks

## Optimizations

Two optimizations have been implemented in this distributed file system.

### Data Proximity for Grep Operations

This is inpired by the POSH Smart Data Proximity concept.

## Configuration

You can configure the filesystem by changing the environment variables located in `.env`
