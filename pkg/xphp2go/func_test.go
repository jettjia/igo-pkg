package php2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type user struct {
	ID   int
	Name string
}

func TestInSlice(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	assert.True(t, InSlice(2, nums))
	assert.False(t, InSlice(5, nums))

	strs := []string{"a", "b", "c"}
	assert.True(t, InSlice("b", strs))
	assert.False(t, InSlice("d", strs))
}

func TestUniqueSlice(t *testing.T) {
	nums := []int{1, 2, 2, 3, 1, 4}
	expect := []int{1, 2, 3, 4}
	assert.Equal(t, expect, UniqueSlice(nums))

	strs := []string{"a", "b", "a", "c", "b"}
	expectStr := []string{"a", "b", "c"}
	assert.Equal(t, expectStr, UniqueSlice(strs))
}

func TestSliceValues(t *testing.T) {
	nums := []int{1, 2, 3}
	assert.Equal(t, nums, SliceValues(nums))

	strs := []string{"a", "b"}
	assert.Equal(t, strs, SliceValues(strs))
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	vals := MapValues(m)
	assert.ElementsMatch(t, []int{1, 2, 3}, vals)

	m2 := map[int]string{1: "x", 2: "y"}
	vals2 := MapValues(m2)
	assert.ElementsMatch(t, []string{"x", "y"}, vals2)
}

func TestSliceColumn(t *testing.T) {
	users := []user{
		{ID: 1, Name: "Tom"},
		{ID: 2, Name: "Jerry"},
		{ID: 3, Name: "Spike"},
	}
	ids := SliceColumn(users, func(u user) int { return u.ID })
	names := SliceColumn(users, func(u user) string { return u.Name })
	assert.Equal(t, []int{1, 2, 3}, ids)
	assert.Equal(t, []string{"Tom", "Jerry", "Spike"}, names)
}
