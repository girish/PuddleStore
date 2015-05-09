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
	cd("", false, client, t)
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

func mkfile(path string, fail bool, client *Client, t *testing.T) {
	err := client.Mkfile(path)
	if err != nil && !fail {
		t.Errorf("'mkfile %v' threw an error: %v", path, err)
		return
	} else if err != nil && fail {
		return
	}

	if fail {
		t.Errorf("'mkifle %v' did not error out when it should've", path)
	}
}
