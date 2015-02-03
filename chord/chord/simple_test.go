package chord

import (
	"testing"
)

func TestSimple(t *testing.T) {
	_, err := CreateNode(nil)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}
}

/*

func TestPaPendejiar(t *testing.T) {
	h := HashKey("unstring")
	a := big.Int{}
	a.SetInt64(7)
	b := big.Int{}
	b.SetInt64(5)
	t.Errorf("%v\n", a)
	t.Errorf("%v\n", b)
	b.Add(&b, &a)
	t.Errorf("%v\n", b)
	fmt.Println(h[0])
}

*/
