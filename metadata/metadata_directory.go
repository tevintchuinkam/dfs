package metadata

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
)

type directory struct {
	name    string
	prev    *directory
	subDirs []*directory
	files   []string
}

func (d *directory) String() string {
	if d == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(d.name)
	for _, file := range d.files {
		sb.WriteString(fmt.Sprintf("\n\t├── %s", file))
	}
	return sb.String()
}

func (d *directory) walkTo(p string) (*directory, error) {
	p = path.Clean(p)
	parts := strings.Split(p, "/")
	currentDir := d

	for _, part := range parts {
		if part == "" {
			continue
		}
		found := false
		for _, subDir := range currentDir.subDirs {
			if subDir.name == part {
				currentDir = subDir
				found = true
				break
			}
		}
		if !found {
			return nil, DirNonExistantError{}
		}
	}
	return currentDir, nil
}

type DirNonExistantError struct {
	Dir string
}

func (e DirNonExistantError) Error() string {
	return fmt.Sprintf("the directory %s does not exist", e.Dir)
}

func isDir(root *directory, name string) bool {
	_, err := root.walkTo(name)
	if errors.Is(err, DirNonExistantError{name}) {
		return false
	}
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	return true
}
