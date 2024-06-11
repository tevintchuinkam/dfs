package main

import (
	"fmt"
	"strings"
)

type directory struct {
	name    string
	prev    *directory
	subDirs []*directory
	files   map[string]*File
}

func (d *directory) String() string {
	if d == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(d.name)
	for _, file := range d.files {
		sb.WriteString(fmt.Sprintf("\n\t├── %s", file.name))
	}
	return sb.String()
}
