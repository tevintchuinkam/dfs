package metadata

import (
	"fmt"
	"testing"
)

func TestDirFuncs(t *testing.T) {
	root := &fileInfo{
		name: "root",
		subEntries: []*fileInfo{
			{
				name: "sub1",
				subEntries: []*fileInfo{
					{name: "sub1_1",
						isDir: true,
					}, {
						name:  "file1.txt",
						isDir: false,
					},
					{
						name:  "file2.txt",
						isDir: false,
					},
				},
			},
			{
				name: "sub2",
				subEntries: []*fileInfo{{
					name:  "file3.txt",
					isDir: false,
				},
				},
			},
			{
				name:  "rootfile.txt",
				isDir: false,
			},
		},
	}

	dir, err := root.walkTo("sub1/sub1_1")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Found directory:", dir.name)
	}

	fmt.Println("Is 'sub2' a directory in root?", isDir(root, "sub2"))
	fmt.Println("Is 'sub3' a directory in root?", isDir(root, "sub3"))
}
