package php2go

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	IdType int    `json:"id_type"`
	IdData string `json:"id_data"`
}

// go test -v -run TestArrayColumn_Struct_Generic ./
func TestArrayColumn_Struct_Generic(t *testing.T) {
	dbUsers := []User{
		{IdType: 1, IdData: "user1"},
		{IdType: 1, IdData: "user2"},
		{IdType: 1, IdData: "user3"},
	}

	expect := []string{"user1", "user2", "user3"}
	res := ArrayColumn(dbUsers, func(u User) string { return u.IdData })
	assert.Equal(t, expect, res)
	fmt.Println(res) // [user1 user2 user3]
}

// go test -v -run Test_JoinStrings ./
func Test_JoinStrings(t *testing.T) {
	data := JoinStrings([]string{"a", "b", "c"}, ",")
	assert.Equal(t, "a,b,c", data)
}

// go test -v -run Test_ArrayDiff ./
func Test_ArrayDiff(t *testing.T) {
	dbUsers := []User{
		{IdType: 1, IdData: "user1"},
		{IdType: 1, IdData: "user2"},
		{IdType: 1, IdData: "user3"},
	}

	queryUsers := []User{
		{IdType: 1, IdData: "user2"},
		{IdType: 1, IdData: "user4"},
	}
	expect := []User{{IdType: 1, IdData: "user1"}, {IdType: 1, IdData: "user3"}}
	res := ArrayDiff(dbUsers, queryUsers)
	assert.Equal(t, expect, res)
	fmt.Println(res)
}

// go test -v -run Test_ArrayDiffByColumnGeneric ./
func Test_ArrayDiffByColumnGeneric(t *testing.T) {
	dbUsers := []User{
		{IdType: 1, IdData: "user1"},
		{IdType: 1, IdData: "user2"},
	}
	queryUsers := []User{
		{IdType: 1, IdData: "user11"},
		{IdType: 1, IdData: "user2"},
	}
	expect := []User{{IdType: 1, IdData: "user11"}}
	res := ArrayDiffByColumnGeneric(dbUsers, queryUsers, func(u User) string { return u.IdData })
	assert.Equal(t, expect, res)
	fmt.Println(res)
}

// go test -v -run Test_ArrayIntersect ./
func Test_ArrayIntersect(t *testing.T) {
	dbUsers := []User{
		{IdType: 1, IdData: "user1"},
		{IdType: 1, IdData: "user2"},
		{IdType: 1, IdData: "user3"},
	}
	queryUsers := []User{
		{IdType: 1, IdData: "user2"},
		{IdType: 1, IdData: "user4"},
	}
	expect := []User{{IdType: 1, IdData: "user2"}}
	res := ArrayIntersect(dbUsers, queryUsers)
	assert.Equal(t, expect, res)
	fmt.Println(res)
}

// go test -v -run Test_ArrayUnique ./
func Test_ArrayUnique(t *testing.T) {
	dbUsers := []string{"user1", "user2", "user3", "user3", "user1", "user4"}
	expect := []string{"user1", "user2", "user3", "user4"}
	res := ArrayUnique(dbUsers)
	assert.Equal(t, expect, res)
	fmt.Println(res)
}

// go test -v -run Test_InArray ./
func Test_InArray(t *testing.T) {
	naturalArr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	assert.True(t, InArray(9, naturalArr))
	assert.False(t, InArray(10, naturalArr))

	strs := []string{"a", "b", "c"}
	assert.True(t, InArray("b", strs))
	assert.False(t, InArray("d", strs))
}

// go test -v -run Test_ArrayFlip ./
func Test_ArrayFlip(t *testing.T) {
	naturalArr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	res := ArrayFlip(naturalArr)
	expect := map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9}
	assert.Equal(t, expect, res)
	fmt.Println(res)
}

// go test -v -run Test_ArrayKeys ./
func Test_ArrayKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := ArrayKeys(m)
	expect := []string{"a", "b", "c"}
	assert.ElementsMatch(t, expect, keys)
	fmt.Println(keys)
}

// go test -v -run Test_ArrayValues ./
func Test_ArrayValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 0}
	vals := ArrayValues(m)
	expect := []int{1, 2, 3, 0}
	assert.ElementsMatch(t, expect, vals)
	fmt.Println(vals)
}

// go test -v -run Test_ArrayCombine ./
func Test_ArrayCombine(t *testing.T) {
	keys := []string{"a", "b", "c"}
	values := []int{1, 2, 3}
	m := ArrayCombine(keys, values)
	expect := map[string]int{"a": 1, "b": 2, "c": 3}
	assert.Equal(t, expect, m)
	fmt.Println(m)
}

// go test -v -run Test_ArrayMerge ./
func Test_ArrayMerge(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{4, 5, 6}
	c := []int{7, 8, 9}
	d := ArrayMerge(a, b, c)
	expect := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	assert.Equal(t, expect, d)
	fmt.Println(d)
}

// go test -v -run Test_IsEqualArray ./
func Test_IsEqualArray(t *testing.T) {
	a := []int{1, 2, 3, 2}
	b := []int{2, 1, 2, 3}
	c := []int{1, 2, 2, 3}
	d := []int{1, 2, 3}
	assert.True(t, IsEqualArray(a, b))
	assert.True(t, IsEqualArray(a, c))
	assert.False(t, IsEqualArray(a, d))
}
