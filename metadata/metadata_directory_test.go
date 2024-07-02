package metadata

import (
	"errors"
	"fmt"
	"testing"
)

func TestDirFuncs(t *testing.T) {
	root := &fileInfo{
		name:  ".",
		isDir: true,
		subEntries: []*fileInfo{
			{
				name:  "sub1",
				isDir: true,
				subEntries: []*fileInfo{
					{name: "sub1_1", isDir: true},
					{name: "file1.txt", isDir: false},
					{name: "file2.txt", isDir: false},
				},
			},
			{
				name:  "sub2",
				isDir: true,
				subEntries: []*fileInfo{
					{name: "file3.txt", isDir: false},
				},
			},
			{name: "rootfile.txt", isDir: false},
		},
	}

	tests := []struct {
		path        string
		shouldExist bool
	}{
		{"sub1/sub1_1", true},
		{"sub3", false},
		{"sub1/file1.txt", true},
		{"sub1/sub1_2", false},
	}

	for _, tt := range tests {
		dir, err := root.walkTo(tt.path)
		if tt.shouldExist {
			if err != nil {
				t.Errorf("Expected %s to exist, but got error: %v", tt.path, err)
			} else {
				t.Logf("Found directory: %s", dir.name)
			}
		} else {
			if err == nil {
				t.Errorf("Expected %s to not exist, but it was found", tt.path)
			} else {
				t.Logf("Correctly did not find directory: %s", tt.path)
			}
		}
	}

	newDirPath := "sub1/sub1_2"
	err := storeFileInfo(root, "sub1", &fileInfo{
		name:  "sub1_2",
		isDir: true,
	})
	if err != nil {
		t.Errorf("Failed to create new directory: %v", err)
	} else {
		t.Logf("Created new directory: %s", newDirPath)
	}

	entries, err := getAllEntriesFromDir(root, "sub1")
	if err != nil {
		t.Errorf("Failed to get all entries from directory: %v", err)
	} else {
		t.Logf("Entries in sub1: %v", entries)
	}
}
func TestGetFileInfoAtIndex(t *testing.T) {
	root := createTestFileTree()
	m := &MetaDataServer{rootDir: root}

	tests := []struct {
		dirName string
		index   int
		want    *fileInfo
		wantErr error
	}{
		{".", 0, &fileInfo{name: "file1.txt"}, nil},
		{".", 1, &fileInfo{name: "dir1", isDir: true}, nil},
		{"./dir1", 0, &fileInfo{name: "file2.txt"}, nil},
		{"./dir1", 1, &fileInfo{name: "file3.txt"}, nil},
		{".", -1, nil, fmt.Errorf("index -1 is out of range. len_entries=2 dirName=.")},
		{".", 2, nil, fmt.Errorf("index 2 is out of range. len_entries=2 dirName=.")},
		{"invalid", 0, nil, errors.New("directory not found: invalid")},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s[%d]", tt.dirName, tt.index), func(t *testing.T) {
			got, err := getFileInfoAtIndex(m, tt.dirName, tt.index)
			if (err != nil) && (tt.wantErr == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("getFileInfoAtIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && (got == nil || got.name != tt.want.name || got.isDir != tt.want.isDir) {
				t.Errorf("getFileInfoAtIndex() = %v, want %v", got, tt.want)
			}
		})
	}

	d, err := root.walkTo("dir1")
	if err != nil {
		t.Error(err)
	}
	t.Log(*d)
}

func createTestFileTree() *fileInfo {
	// Create a sample file tree
	return &fileInfo{
		name:  ".",
		isDir: true,
		subEntries: []*fileInfo{
			{
				name:  "file1.txt",
				isDir: false,
			},
			{
				name:  "dir1",
				isDir: true,
				subEntries: []*fileInfo{
					{
						name:  "file2.txt",
						isDir: false,
					},
					{
						name:  "file3.txt",
						isDir: false,
					},
				},
			},
		},
	}
}
