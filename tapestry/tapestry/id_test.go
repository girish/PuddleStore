package tapestry

import (
	"testing"
)

// Tests to make sure that prefix length is working
func TestSharedPrefixLength(t *testing.T) {
	a := ID{1,2,3,4,5,6,7,8,9,6,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count := SharedPrefixLength(a, b)
	if (count != 9) {
		t.Errorf("The SharedPrefixLength does not work")
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{2,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count = SharedPrefixLength(a, b)
	if (count != 0) {
		t.Errorf("The SharedPrefixLength does not work")
	}

	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count = SharedPrefixLength(a, b)
	if (count != 40) {
		t.Errorf("The SharedPrefixLength does not work")
	}

}

func TestBetterChoice(t *testing.T) {
	a := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice := id.BetterChoice(a, b)
	if (choice) {//choice should be false since they are the same
		t.Errorf("The BetterChoice does not work")
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,8,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (!choice) {//choice should be true for the prefix
		t.Errorf("The BetterChoice does not work")
	}
	a = ID{1,2,3,4,5,6,7,6,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,7,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (choice) {//choice should be false (b is better) because it is the closes when incrementing by 1
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}

	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,7,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,6,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (!choice) {//choice should be true because it is the closes when incrementing by 1 % base
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}
}

func TestCloser(t *testing.T) {

}