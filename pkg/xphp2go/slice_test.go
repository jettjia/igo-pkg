package php2go

import (
	"fmt"
	"testing"
)

func Test_IntersectArray(t *testing.T) {
	a := []string{"a", "b", "1", "c", "d"}
	b := []string{"a", "b", "d"}

	fmt.Println(SliceIntersectArray(a, b)) //[a b d]
}

func Test_DiffArray(t *testing.T) {
	a := []string{"a", "b", "1", "c", "d"}
	b := []string{"a", "b", "d"}

	fmt.Println(SliceDiffArray(a, b)) //[1 c]
}

func Test_DiffArray2(t *testing.T) {
	a := []string{"a", "b", "d", "e"}
	b := []string{"a", "b", "1", "c", "d"}

	fmt.Println(SliceDiffArray(a, b)) //[e]
}

func Test_RemoveRepeatedElement(t *testing.T) {
	a := []string{"a", "b", "d", "e", "b"}

	fmt.Println(SliceRemoveRepeatedElement(a)) //[a d e b]
}
