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
	if bytes.Equal(key, differentKey) {
		t.Errorf("Hash keys made by the different strings are equal.")
	}
}

// --------- Between --------------

func TestBetweenSimple(t *testing.T) {
	A := []byte{10}
	B := []byte{15}
	C := []byte{20}

	// B is between A and C...
	if !Between(B, A, C) {
		t.Errorf("Between does not return true when it should. %v < %v < %v",
			A[0], B[0], C[0])
	}
	// ...but it shouldn't be between C and A
	if Between(B, C, A) {
		t.Errorf("Between returns true when it shouldn't. %v < %v < %v",
			C[0], B[0], A[0])
	}
	// Between shouldn't be right inclusive.
	if Between(B, A, B) {
		t.Errorf("Between returns true when it shouldn't. %v < %v < %v",
			A[0], B[0], B[0])
	}

	if !Between(A, C, B) {
		t.Errorf("Between returns true when it shouldn't. %v < %v < %v",
			C[0], A[0], B[0])
	}
	if Between(A, B, C) {
		t.Errorf("Between returns true when it shouldn't. %v < %v < %v",
			B[0], A[0], C[0])
	}

}

func TestBetweenEdge(t *testing.T) {
	A := []byte{230}
	B := []byte{15}
	C := []byte{80}

	// B is between A and C...
	if !Between(B, A, C) {
		t.Errorf("Between does not return true when it should. %v < %v < %v",
			A[0], B[0], C[0])
	}
	// ...but it shouldn't be between C and A
	if Between(B, C, A) {
		t.Errorf("Between returns true when it shouldn't. %v < %v < %v",
			A[0], B[0], C[0])
	}
}

// -------------------------------------

func TestBetweenRightIncl(t *testing.T) {
	A := []byte{10}
	B := []byte{20}

	if !BetweenRightIncl(B, A, B) {
		t.Errorf("Between does not return true when it should. %v < %v <= %v",
			A[0], B[0], B[0])
	}
	if BetweenRightIncl(A, A, B) {
		t.Errorf("Between returns true when it shouldn't. %v < %v <= %v",
			A[0], A[0], B[0])
	}
}
