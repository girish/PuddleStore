package chord

import (
	"math"
	"testing"
	"time"
)

// Checks if FingerTable was initialized correctly
// Checks all combinations of node id numbers and potentiall
// start entries on the finger table.
// from 0 to 255, check all 8 rows' start.
func TestInitFingerTable(t *testing.T) {
	var res, expected []byte
	m := int(math.Pow(2, KEY_LENGTH))
	for i := 0; i < m; i++ {
		node, _ := CreateDefinedNode(nil, []byte{byte(i)})
		for j := 0; j < KEY_LENGTH; j++ {
			res = node.FingerTable[j].Start
			expected = []byte{byte(
				(i + int(math.Pow(float64(2), float64(j)))) % m)}
			if !EqualIds(res, expected) {
				t.Errorf("[%v] BAD ENTRY: %v != %v", i, res, expected)
			}
		}
	}

	nodes, _ := CreateNNodes(10)
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		for j := 0; j < KEY_LENGTH; j++ {
			res = nodes[i].FingerTable[j].Start
			expected = []byte{byte(
				(i + int(math.Pow(float64(2), float64(j)))) % m)}
			if !EqualIds(res, expected) {
				t.Errorf("[%v] BAD ENTRY: %v != %v", i, res, expected)
			}
		}
	}
	// ShutdownNode(node)
}

func TestFixNextFinger(t *testing.T) {
	nodes, _ := CreateNNodes(10)
	time.Sleep(time.Second * 3)
	for i := 0; i < 10; i++ {
		PrintFingerTableLegible(nodes[i])
	}
	t.Errorf("Hash keys made by the same string are not equal.")
}

// REDUNDANT; SAME AS TestInitFingerTable!
// Checks all combinations of node id numbers and potentiall
// start entries on the finger table.
// from 0 to 255, check all 8 rows' start.
/*
func TestFingerMath(t *testing.T) {
	var res, expected []byte
	m := int(math.Pow(2, KEY_LENGTH))
	for i := 0; i < m; i++ {
		for j := 0; j < KEY_LENGTH; j++ {
			res = fingerMath([]byte{byte(i)}, j, KEY_LENGTH)
			expected = []byte{byte(
				(i + int(math.Pow(float64(2), float64(j)))) % m)}
			if !EqualIds(res, expected) {
				t.Errorf("[%v] BAD MATH: %v != %v", i, res, expected)
			}
		}
	}
}
*/
