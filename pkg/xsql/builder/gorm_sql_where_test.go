package builder

import (
	"fmt"
	"testing"
)

func Test_GormBuildWhere(t *testing.T) {
	queries := []*Query{
		{Key: "field1", Value: "value1", Operator: Operator_opEq},
		{Key: "field2", Value: []string{"v2-1", "v2-2"}, Operator: Operator_opIn},
	}

	whereStr, values, err := GormBuildWhere(queries)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (field1=? AND field2 LIKE ?)
	fmt.Println(values)   // [value1 value2]
}

func Test_GormBuildWhereMap(t *testing.T) {
	queries := []*Query{
		{Key: "field1", Value: "value1", Operator: Operator_opEq},
		{Key: "field2", Value: []string{"v2-1", "v2-2"}, Operator: Operator_opIn},
	}

	whereMap := GormBuildWhereMap(queries)

	fmt.Println(whereMap) // map[field1 =:value1 field2 in:[v2-1 v2-2]]
}

// go test -v -run Test_GormBuildWhere_like ./
// 自动带上两边都有的 %
func Test_GormBuildWhere_like(t *testing.T) {
	queries := []*Query{
		{Key: "field1", Value: `%ivalue1`, Operator: Operator_opLikePercent},
	}

	whereStr, values, err := GormBuildWhere(queries)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) //  (field1 LIKE ?)
	fmt.Println(values)   // [%value1%]
}
