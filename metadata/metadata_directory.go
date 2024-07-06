package metadata

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
)

type fileInfo struct {
	name       string
	size       int64
	mode       fs.FileMode
	modTime    time.Time
	isDir      bool
	subEntries []*fileInfo
	prev       *fileInfo
	sys        any
	port       int
	fullPath   string
}

// GenerateFileTree generates the file tree starting from the given root and returns it as a string.
func GenerateFileTree(root *fileInfo) string {
	indent := "\t"
	var result strings.Builder
	if root.isDir {
		result.WriteString(indent + root.name + "/\n")
		indent += "  "
		for _, entry := range root.subEntries {
			result.WriteString(GenerateFileTree(entry))
		}
	} else {
		result.WriteString(indent + root.name + "\n")
	}
	return result.String()
}

func (d *fileInfo) walkTo(p string) (*fileInfo, error) {
	if d == nil {
		slog.Error("dir is nil")
		return nil, errors.New("dir is nil")
	}
	if !d.isDir {
		err := fmt.Errorf("%s is not a directory", d.name)
		slog.Error(err.Error())
		return nil, err
	}
	p = path.Clean(p)
	if d.name == p {
		return d, nil
	}
	parts := strings.Split(p, string(os.PathSeparator))
	currentDir := d

	for _, part := range parts {
		if part == "" {
			continue
		}
		found := false
		for _, subDir := range currentDir.subEntries {
			if subDir.name == part {
				currentDir = subDir
				found = true
				break
			}
		}
		if !found {
			return nil, DirNonExistantError{part}
		}
	}
	return currentDir, nil
}

type EndOfDirectoryError struct {
}

func (e EndOfDirectoryError) Error() string {
	return "end of directory reached"
}

func getFileInfoAtIndex(m *MetaDataServer, dirName string, index int) (*fileInfo, error) {
	parent, err := m.rootDir.walkTo(path.Clean(dirName))
	if err != nil {
		slog.Error("directory not found", "name", dirName)
		return nil, fmt.Errorf("directory not found: %s", dirName)
	}
	if index < 0 {
		return nil, fmt.Errorf("negative index provided: %d", index)
	}

	if index >= len(parent.subEntries) {
		return nil, EndOfDirectoryError{}
	}

	return parent.subEntries[index], nil
}

type EntryAlreadyExistsError struct {
	Name string
}

func (e EntryAlreadyExistsError) Error() string {
	return fmt.Sprintf("entry with name %s already exists", e.Name)
}
func storeFileInfo(root *fileInfo, dirPath string, fi *fileInfo) error {
	// Find the parent directory
	dirPath = path.Dir(dirPath)
	parentDir, err := root.walkTo(dirPath)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// Check if an entry with the same name already exists
	for _, entry := range parentDir.subEntries {
		if entry.name == fi.name {
			return EntryAlreadyExistsError{entry.name}
		}
	}

	// Append the new fileInfo to the parent's subEntries
	parentDir.subEntries = append(parentDir.subEntries, fi)
	return nil
}

// entryAlreadyExists checks if a file or directory with the specified path already exists.
func entryAlreadyExists(root *fileInfo, p string) bool {
	dirPath := path.Dir(p)
	parentDir, err := root.walkTo(dirPath)
	if err != nil {
		return false
	}
	fileName := path.Base(p)

	// Check if an entry with the same name already exists
	for _, entry := range parentDir.subEntries {
		if entry.name == fileName {
			return true
		}
	}
	return false
}

// getAllEntriesFromDir retrieves all entries from the specified directory.
func getAllEntriesFromDir(rootDir *fileInfo, dir string) ([]*fileInfo, error) {
	// Find the directory
	directory, err := rootDir.walkTo(dir)
	if err != nil {
		return nil, err
	}

	// Check if it's indeed a directory
	if !directory.isDir {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	return directory.subEntries, nil
}

type DirNonExistantError struct {
	Dir string
}

func (e DirNonExistantError) Error() string {
	return fmt.Sprintf("the directory %s does not exist", e.Dir)
}

func isDir(root *fileInfo, name string) bool {
	name = path.Clean(name)
	// slog.Debug("comparing root to path", "root", root.name, "root_isdir", root.isDir, "root.name == name", root.name == name, "path", name)
	if root.isDir && root.name == name {
		return true
	}
	if _, err := root.walkTo(name); err != nil {
		return false
	}
	return true
}

// Name returns the base name of the file.
func (fi fileInfo) Name() string {
	return fi.name
}

// Size returns the length in bytes for regular files; system-dependent for others.
func (fi fileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file mode bits.
func (fi fileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime returns the modification time.
func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the file is a directory.
func (fi fileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns the underlying data source (can return nil).
func (fi fileInfo) Sys() any {
	return fi.sys
}

// Ensure that FileInfo implements the FileInfo interface
var _ os.FileInfo = (*fileInfo)(nil)
