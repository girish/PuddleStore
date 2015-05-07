package puddlestore

import (
	// "fmt"
	"testing"
)

func TestGobEncoding(t *testing.T) {
	inode := new(Inode)
	inode.name = "Test inode"
	inode.filetype = 1
	inode.size = 666
	inode.indirect = "F666"

	bytes, err := inode.GobEncode()
	if err != nil {
		t.Errorf("Gob encode didn't work.")
	}

	sameInode := new(Inode)
	sameInode.GobDecode(bytes)

	if inode.name != sameInode.name {
		t.Errorf("Name not the same\n\t%v != %v.", inode.name, sameInode.name)
	}

	if inode.filetype != sameInode.filetype {
		t.Errorf("Name not the same\n\t%v != %v.", inode.filetype, sameInode.filetype)
	}

	if inode.size != sameInode.size {
		t.Errorf("Name not the same\n\t%v != %v.", inode.size, sameInode.size)
	}

	if inode.indirect != sameInode.indirect {
		t.Errorf("Name not the same\n\t%v != %v.", inode.indirect, sameInode.indirect)
	}
}
