package puddlestore

import (
	// "fmt"
	"testing"
)

func TestRemoveExcessSlashes(t *testing.T) {
	assertEqual(removeExcessSlashes("/"), "/", t)
	assertEqual(removeExcessSlashes("//"), "/", t)
	assertEqual(removeExcessSlashes("///"), "/", t)

	assertEqual(removeExcessSlashes("/path"), "/path", t)
	assertEqual(removeExcessSlashes("//path"), "/path", t)
	assertEqual(removeExcessSlashes("///path"), "/path", t)
	assertEqual(removeExcessSlashes("///path/"), "/path", t)
	assertEqual(removeExcessSlashes("///path//"), "/path", t)
	assertEqual(removeExcessSlashes("///path///"), "/path", t)

	assertEqual(removeExcessSlashes("/another/path"), "/another/path", t)
	assertEqual(removeExcessSlashes("//another////path///"), "/another/path", t)

	assertEqual(removeExcessSlashes("/ultima/y/nos/vamos"), "/ultima/y/nos/vamos", t)
	assertEqual(removeExcessSlashes("//ultima//y//nos//vamos/"), "/ultima/y/nos/vamos", t)
}

func assertEqual(str1, str2 string, t *testing.T) {
	if str1 != str2 {
		t.Errorf("%v is not equal to %v", str1, str2)
	}
}
