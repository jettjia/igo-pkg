package php2go

import (
	"fmt"
	"reflect"
	"testing"
)

// Expected to be equal.
func equal(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// Expected to be unequal.
func unequal(t *testing.T, expected, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

func Test_Strpos(t *testing.T) {
	equal(t, 6, Strpos("hello wworld", "w", -6))
}

func Test_Explode(t *testing.T) {
	rsp := Explode(",", "1,2,3")
	fmt.Println(rsp) // [1 2 3]
}

func Test_Implode(t *testing.T) {
	rsp := Implode(",", []string{"a", "b", "d"})
	fmt.Println(rsp) // a,b,d
}

func Test_Addslashes(t *testing.T) {
	tAddslashes := Addslashes("f'oo b\"ar")
	equal(t, `f\'oo b\"ar`, tAddslashes)
	equal(t, `f'oo b"ar\a\\\`, Stripslashes("f\\'oo b\\\"ar\\\\a\\\\\\\\\\\\"))
}
