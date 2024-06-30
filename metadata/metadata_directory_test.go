package metadata

import (
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
