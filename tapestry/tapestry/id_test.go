import (
	"testing"
)

// Tests to make sure that prefix length is working
func TestSharedPrefixLength(t *testing.T) {
	a := []Digit{1,2,3,4,5,5,5,5}
	b := []Digit{1,2,3,4,6,6,6,6}
	count := SharedPrefixLength(a, b)
	if (count != 4) {
		t.Errorf("The SharedPrefixLength does not work")
	}
}