/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Utility functions to help with dealing with ID hashes in Chord. */
/*                                                                           */

package chord

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math/big"
)

/* Hash a string to its appropriate size */
func HashKey(key string) []byte {
	h := sha1.New()
	h.Write([]byte(key))
	v := h.Sum(nil)
	return v[:KEY_LENGTH/8]
}

/* Convert a []byte to a big.Int string, useful for debugging/logging */
func HashStr(keyHash []byte) string {
	keyInt := big.Int{}
	keyInt.SetBytes(keyHash)
	return keyInt.String()
}

func EqualIds(a, b []byte) bool {
	return bytes.Equal(a, b)
}

/* Example of how to do math operations on []byte IDs, you may not need this function. */
func AddIds(a, b []byte) []byte {
	aInt := big.Int{}
	aInt.SetBytes(a)

	bInt := big.Int{}
	bInt.SetBytes(b)

	sum := big.Int{}
	sum.Add(&aInt, &bInt)
	return sum.Bytes()
}

/* On this crude ascii Chord ring, X is between (A : B)
   ___
  /   \-A
 |     |
B-\   /-X
   ---
*/
func Between(nodeX, nodeA, nodeB []byte) bool {

	//TODO students should implement this method
	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = false
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) == -1)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) == -1)
	}

	return result
}

/* Is X between (A : B] */
func BetweenRightIncl(nodeX, nodeA, nodeB []byte) bool {

	//TODO students should implement this method
	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = true
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) <= 0)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) <= 0)
	}

	return result
}

func NodeStr(node *Node) string {
	var succ []byte
	var pred []byte
	if node.Successor != nil {
		succ = node.Successor.Id
	}
	if node.Predecessor != nil {
		pred = node.Predecessor.Id
	}

	return fmt.Sprintf("Node-%v: {succ:%v, pred:%v}", node.Id, succ, pred)
}
