package tapestry

import (
	"testing"
)

func CheckGet(err error, result []byte, expected string, t *testing.T) {
	if err != nil {
		t.Errorf("Get errored out. returned: %v", err)
		return
	}

	if string(result) != expected {
		t.Errorf("Get(\"%v\") did not return expected result (%v)",
			string(result), expected)
	}
}

func TestPutAndGet(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
	}

	port = 58000
	id := ID{5, 8, 3, 15}
	node0 := makeTapestry(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{7, 0, 0xf, 5}
	node2 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{7, 0, 0xf, 0xa}
	node3 := makeTapestry(id, node0.local.node.Address, t)

	node0.Store("key", []byte("llave"))
	node1.Store("table", []byte("mesa"))
	node2.Store("chair", []byte("silla"))
	node3.Store("fork", []byte("tenedor"))

	result, err := node0.Get("key")
	CheckGet(err, result, "llave", t)
	result, err = node1.Get("key")
	CheckGet(err, result, "llave", t)
	result, err = node2.Get("key")
	CheckGet(err, result, "llave", t)
	result, err = node3.Get("key")
	CheckGet(err, result, "llave", t)

	result, err = node0.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node1.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node2.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node3.Get("table")
	CheckGet(err, result, "mesa", t)

	result, err = node0.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node1.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node2.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node3.Get("chair")
	CheckGet(err, result, "silla", t)

	result, err = node0.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node1.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node2.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node3.Get("fork")
	CheckGet(err, result, "tenedor", t)

	node0.Leave()
	node1.Leave()
	node2.Leave()
	node3.Leave()
}
