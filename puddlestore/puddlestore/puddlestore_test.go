package puddlestore

import (
	"testing"
	"time"
)

const EMPTY_FILE_LS = "\t./\t../"

func TestLsAndMkdirFileAndCd(t *testing.T) {
	puddle, err := Start()
	if err != nil {
		return
		t.Errorf("Could not init puddlestore: %v", err)
	}
	time.Sleep(time.Millisecond * 500)

	client, err := CreateClient(puddle.Local)
	if err != nil {
		t.Errorf("Could not init client: %v", err)
		return
	}

	var expected string

	// >> ls
	expected = EMPTY_FILE_LS
	ls("", expected, client, t)

	// >> ls doesnotexist
	ls("doesnotexist", "", client, t)

	// >> mkdir dir
	mkdir("dir", false, client, t)
	// >> mkdir dir (should fail)
	mkdir("dir", true, client, t)

	// >> ls
	expected = EMPTY_FILE_LS + "\tdir/"
	ls("", expected, client, t)

	// >> mkdir dir2
	mkdir("dir2", false, client, t)

	// >> ls
	expected = EMPTY_FILE_LS + "\tdir/\tdir2/"
	ls("", expected, client, t)

	// >> cd dir
	cd("dir", false, client, t)

	// >> ls
	expected = EMPTY_FILE_LS
	ls("", expected, client, t)

	// >> ls /
	expected = EMPTY_FILE_LS + "\tdir/\tdir2/"
	ls("/", expected, client, t)

	// >> mkfile file
	mkfile("file", false, client, t)
	// >> mkfile file (fail)
	mkfile("file", true, client, t)
	// >> cd file (fail)
	cd("file", true, client, t)

	// >> ls
	expected = EMPTY_FILE_LS + "\tfile"
	ls("", expected, client, t)

	// >> mkdir dir2
	mkdir("dir3", false, client, t)

	// >> ls
	expected = EMPTY_FILE_LS + "\tfile\tdir3/"
	ls("", expected, client, t)

	// >> ls /////
	expected = EMPTY_FILE_LS + "\tdir/\tdir2/"
	ls("/////", expected, client, t)

	// >> ls //dir///
	expected = EMPTY_FILE_LS + "\tfile\tdir3/"
	ls("//dir///", expected, client, t)

	// >> cd
	cd("", false, client, t)

	// >> cd doesnotexist
	cd("doesnotexist", true, client, t)

	// >> cd .. (we are at root)
	cd("..", false, client, t)
	// >> ls
	expected = EMPTY_FILE_LS + "\tdir/\tdir2/"
	ls("", expected, client, t)
	// >> cd . (we are at root)
	cd(".", false, client, t)
	// >> ls
	ls("", expected, client, t)

	// rmdir dir (fail)
	rmdir("dir", true, client, t)

	// cd dir
	cd("dir", false, client, t)
	// rmdir dir3
	// rmdir("dir3", false, client, t)
	// rmfile file
	rmfile("file", false, client, t)
	// cd ..
	cd("..", false, client, t)
	// rmdir dir
	// rmdir("dir", false, client, t)

}

func TestWriteReadFile(t *testing.T) {
	puddle, err := Start()
	if err != nil {
		return
		t.Errorf("Could not init puddlestore: %v", err)
	}
	time.Sleep(time.Millisecond * 500)

	client, err := CreateClient(puddle.Local)
	if err != nil {
		t.Errorf("Could not init client: %v", err)
		return
	}

	var sendingBuf, expectingBuf []byte
	var nread, expectedRead uint32

	// >> mkfile file
	mkfile("file", false, client, t)

	// >> writefile file 0 mellamojuan
	writefile("file", 0, []byte("mellamojuan"), 11, false, client, t)

	// >> cat file 0 11
	cat("file", 0, 11, []byte("mellamojuan"), 11, false, client, t)
	// >> cat file 2 9
	cat("file", 2, 9, []byte("llamojuan"), 9, false, client, t)

	// >> cat file 3 9
	expectingBuf = make([]byte, 9)
	FillUpBytes(expectingBuf, "lamojuan")
	cat("file", 3, 9, expectingBuf, 9, false, client, t)

	// >> cat file 11 9
	expectingBuf = make([]byte, 9)
	cat("file", 11, 9, expectingBuf, 9, false, client, t)

	// >> writefile file 4094 "Im in the middle of 2 blocks"
	sendingBuf = []byte("Im in the middle of 2 blocks")
	writefile("file", 4094, sendingBuf, uint32(len(sendingBuf)), false, client, t)

	// >> cat file 4094 n
	expectingBuf = sendingBuf
	nread = uint32(len(sendingBuf))
	expectedRead = uint32(len(sendingBuf))
	cat("file", 4094, nread, expectingBuf, expectedRead, false, client, t)

	// Write in the whole file space
	sendingBuf = FullOfAs(BLOCK_SIZE * FILES_PER_INODE)
	writefile("file", 0, sendingBuf, uint32(len(sendingBuf)), false, client, t)

	expectingBuf = sendingBuf
	nread = uint32(len(sendingBuf))
	expectedRead = uint32(len(sendingBuf))
	cat("file", 0, nread, expectingBuf, expectedRead, false, client, t)

	// >> rmfile file
	rmfile("file", false, client, t)
}

// --------------- WRAPPERS ---------------------------------------------------

// Executes ls and assert the expected output is correct.
// If 'expected' = "", it expects it to fail.
func ls(path, expected string, client *Client, t *testing.T) {
	elements, err := client.Ls(path)
	if err != nil && expected != "" {
		t.Errorf("'ls %v' threw an error: %v", path, err)
		return
	} else if err != nil && expected == "" {
		return
	}

	if expected == "" {
		t.Errorf("'ls %v' did not error out when it should've", path)
	}

	if elements != expected {
		t.Errorf("'ls %v' did not give the expected result:\n%v != %v", path,
			elements, expected)
	}
}

func cd(path string, fail bool, client *Client, t *testing.T) {
	err := client.Cd(path)
	if err != nil && !fail {
		t.Errorf("'cd %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'cd %v' did not error out when it should've", path)
	}
}

func mkdir(path string, fail bool, client *Client, t *testing.T) {
	err := client.Mkdir(path)
	if err != nil && !fail {
		t.Errorf("'mkdir %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'mkdir %v' did not error out when it should've", path)
	}
}

func rmdir(path string, fail bool, client *Client, t *testing.T) {
	err := client.Rmdir(path)
	if err != nil && !fail {
		t.Errorf("'rmdir %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'rmdir %v' did not error out when it should've", path)
	}
}

func mkfile(path string, fail bool, client *Client, t *testing.T) {
	err := client.Mkfile(path)
	if err != nil && !fail {
		t.Errorf("'mkfile %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'mkfile %v' did not error out when it should've", path)
	}
}

func writefile(path string, location uint32, bytes []byte, expected uint32,
	fail bool, client *Client, t *testing.T) {

	written, err := client.Writefile(path, location, bytes)
	if err != nil && !fail {
		t.Errorf("'writefile %v %v %v' threw an error: %v", path, location, bytes, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'writefile %v %v %v' did not error out when it should've",
			path, location, bytes)
		return
	}

	if written != expected {
		t.Errorf("'writefile %v %v %v' did not write what was expected %v != %v",
			path, location, bytes, written, expected)
	}

}

func cat(path string, location uint32, count uint32, expected []byte,
	expectedRead uint32, fail bool, client *Client, t *testing.T) {

	buffer, read, err := client.Cat(path, location, count)
	if err != nil && !fail {
		t.Errorf("'cat %v %v %v' threw an error: %v", path, location, count, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'cat %v %v %v' did not error out when it should've",
			path, location, count)
		return
	}

	for i := uint32(0); i < read; i++ {
		if buffer[i] != expected[i] {
			t.Errorf("'cat %v %v %v' did not read what was expected \"%v\" != \"%v\"",
				path, location, count, buffer, expected)
		}
	}

	if read != expectedRead {
		t.Errorf("'cat %v %v %v' did not read what was expected %v != %v",
			path, location, count, read, expectedRead)
	}

}

func rmfile(path string, fail bool, client *Client, t *testing.T) {
	err := client.Rmfile(path)
	if err != nil && !fail {
		t.Errorf("'rmfile %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'rmfile %v' did not error out when it should've", path)
	}
}

func FillUpBytes(bytes []byte, str string) {
	for i := 0; i < len(str); i++ {
		bytes[i] = str[i]
	}
}

func FullOfAs(times uint32) []byte {
	bytes := make([]byte, times)
	for i := uint32(0); i < times; i++ {
		bytes[i] = 'a'
	}
	return bytes
}
