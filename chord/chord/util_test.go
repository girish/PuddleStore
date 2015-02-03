package chord

import (
	"bytes"
	"testing"
)

func TestHashKey(t *testing.T) {
	key := HashKey("Im a string")
	sameKey := HashKey("Im a string")

	if !bytes.Equal(key, sameKey) {
		t.Errorf("Hash keys made by the same string are not equal.")
	}

	differentKey := HashKey("Im another string, totally different.")
	if !bytes.Equal(key, differentKey) {
		t.Errorf("Hash keys made by the different strings are equal.")
	}
}
