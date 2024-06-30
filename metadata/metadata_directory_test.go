package metadata

import (
	"fmt"
	"testing"
)

func TestDirFuncs(t *testing.T) {
	root := &directory{
		name: "root",
		subDirs: []*directory{
			{
				name: "sub1",
				subDirs: []*directory{
					{name: "sub1_1"},
				},
				files: []string{"file1.txt", "file2.txt"},
			},
			{
				name:  "sub2",
				files: []string{"file3.txt"},
			},
		},
		files: []string{"rootfile.txt"},
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
