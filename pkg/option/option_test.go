package option

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	name string
	age  int
}

func WithName(name string) Option[User] {
	return func(u *User) {
		u.name = name
	}
}

func WithAge(age int) Option[User] {
	return func(u *User) {
		u.age = age
	}
}

func WithNameErr(name string) OptionErr[User] {
	return func(u *User) error {
		if name == "" {
			return errors.New("name cannot be empty")
		}
		u.name = name
		return nil
	}
}

func WithAgeErr(age int) OptionErr[User] {
	return func(u *User) error {
		if age < 0 {
			return errors.New("age cannot be less than 0")
		}
		u.age = age
		return nil
	}
}

func TestApply(t *testing.T) {
	u := &User{}
	Apply[User](u, WithName("Tom"), WithAge(18))
	assert.Equal(t, u, &User{name: "Tom", age: 18})
}

func TestApplyErr(t *testing.T) {
	u := &User{}
	err := ApplyErr[User](u, WithNameErr("Tom"), WithAgeErr(18))
	require.NoError(t, err)
	assert.Equal(t, u, &User{name: "Tom", age: 18})

	err = ApplyErr[User](u, WithNameErr(""), WithAgeErr(18))
	assert.Equal(t, errors.New("name cannot be empty"), err)
}

func ExampleApplyErr() {
	u := &User{}
	err := ApplyErr[User](u, WithNameErr("Tom"), WithAgeErr(18))
	fmt.Println(err)
	fmt.Println(u)

	err = ApplyErr[User](u, WithNameErr(""), WithAgeErr(18))
	fmt.Println(err)
	// Output:
	// <nil>
	// &{Tom 18}
	// name cannot be empty
}

func ExampleApply() {
	u := &User{}
	Apply[User](u, WithName("Tom"), WithAge(18))
	fmt.Println(u)
	// Output:
	// &{Tom 18}
}
