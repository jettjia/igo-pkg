package builder

import (
	"fmt"
	"testing"
)

func Test_BuildSelect(t *testing.T) {
	selectField := []string{}
	res, err := BuildSelect(selectField)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res) // *

	selectField = []string{"table1.name", "table2.name2", "name3"}
	res, err = BuildSelect(selectField)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res) // * table1.name,table2.name2,name3

}

func Test_BuildSelectVariable(t *testing.T) {
	res := BuildSelectVariable()
	fmt.Println(res) // [*]
}

func Test_BuildSelectVariable2(t *testing.T) {
	selectField := []string{"table1.name", "table2.name2", "name3"}
	res := BuildSelectVariable(selectField)
	fmt.Println(res) // [table1.name table2.name2 name3]
}

func Test_BuildSelectVariable3(t *testing.T) {
	selectField1 := []string{"table1.name"}
	selectField2 := []string{"table2.name"}
	res := BuildSelectVariable(selectField1, selectField2)
	fmt.Println(res) // [table1.name table2.name]
}
