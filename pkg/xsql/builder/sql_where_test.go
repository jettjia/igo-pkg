package builder

import (
	"fmt"
	"testing"
)

func Test_BuildWhere1(t *testing.T) {
	where := map[string]interface{}{}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) //
	fmt.Println(values)   // []interface{}
}

func Test_BuildWhere2(t *testing.T) {
	where := map[string]interface{}{
		"city": []string{"beijing", "shanghai"},
		// The in operator can be omitted by default,
		// which is equivalent to:
		// "city in": []string{"beijing", "shanghai"},
		"score": 5,
		"age >": 35,
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (score=? AND city IN (?,?) AND age>?)
	fmt.Println(values)   // []interface{}{5, "beijing", "shanghai", 35}
}

func Test_BuildWhere3(t *testing.T) {
	where := map[string]interface{}{
		"city": []string{"beijing", "shanghai"},
		// The in operator can be omitted by default,
		// which is equivalent to:
		// "city in": []string{"beijing", "shanghai"},
		"score": 5,
		"age >": 35,
		"_or": []map[string]interface{}{
			{
				"x1":    11,
				"x2 >=": 45,
			},
			{
				"x3":    "234",
				"x4 <>": "tx2",
			},
		},
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (((x1=? AND x2>=?) OR (x3=? AND x4!=?)) AND score=? AND city IN (?,?) AND age>?)
	fmt.Println(values)   // []interface{}{11 45 234 tx2 5 beijing shanghai 35}
}

func Test_BuildWhere4_like(t *testing.T) {
	where := map[string]interface{}{
		"name like": "%123",
		"_lockMode": "exclusive",
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) //   (name LIKE ?)
	fmt.Println(values)   // [%123]
}

func Test_BuildWhere5_or_sonOr(t *testing.T) {
	where := map[string]interface{}{
		"foo":      "bar",
		"qq":       "tt",
		"age in":   []interface{}{1, 3, 5, 7, 9},
		"vx":       []interface{}{1, 3, 5},
		"faith <>": "Muslim",
		"_or": []map[string]interface{}{
			{
				"aa": 11,
				"bb": "xswl",
			},
			{
				"cc":    "234",
				"dd in": []interface{}{7, 8},
				"_or": []map[string]interface{}{
					{
						"neeest_ee <>": "dw42",
						"neeest_ff in": []interface{}{34, 59},
					},
					{
						"neeest_gg":        1259,
						"neeest_hh not in": []interface{}{358, 1245},
					},
				},
			},
		},
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (((aa=? AND bb=?) OR (((neeest_ff IN (?,?) AND neeest_ee!=?) OR (neeest_gg=? AND neeest_hh NOT IN (?,?))) AND cc=? AND dd IN (?,?))) AND foo=? AND qq=? AND age IN (?,?,?,?,?) AND vx IN (?,?,?) AND faith!=?)

	fmt.Println(values) // [11 xswl 34 59 dw42 1259 358 1245 234 7 8 bar tt 1 3 5 7 9 1 3 5 Muslim]
}

func Test_BuildWhere6_or(t *testing.T) {
	where := map[string]interface{}{
		"score": 5,
		"_or": []map[string]interface{}{
			{
				"field1": 11,
			},
			{
				"field1": "234",
			},
			{
				"_or": []map[string]interface{}{
					{
						"field2": "dw42",
					},
					{
						"field2": 1259,
					},
				},
			},
		},
		"age": 5,
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) //   (((field1=?) OR (field1=?) OR (((field2=?) OR (field2=?)))) AND age=? AND score=?)
	fmt.Println(values)   // []interface{}{11 45 234 tx2 5 beijing shanghai 35}
}

func Test_BuildWhere7_or_or(t *testing.T) {
	where := map[string]interface{}{
		// 位置条件（15号线或朝阳区）
		"_or_location": []map[string]interface{}{{
			"subway": "beijing_15",
		}, {
			"district": "Chaoyang",
		}},
		// 类型（有煤气或有电梯）
		"_or_functions": []map[string]interface{}{{
			"has_gas": true,
		}, {
			"has_lift": true,
		}},
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (((subway=?) OR (district=?)) AND ((has_gas=?) OR (has_lift=?)))
	fmt.Println(values)   // [beijing_15 Chaoyang true true]
}

func Test_BuildWhere8_or(t *testing.T) {
	where := map[string]interface{}{
		// 类型（有煤气或有电梯）
		"_or_field1": []map[string]interface{}{
			{"field1": "11"},
			{"field1": "12"},
		},
		// 类型（有煤气或有电梯）
		"_or_field2": []map[string]interface{}{
			{"field2": "21"},
			{"field2": "22"},
		},
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) // (((field2=?) OR (field2=?)) AND ((field1=?) OR (field1=?)))
	fmt.Println(values)   // [21 22 11 12]
}

func Test_expValues(t *testing.T) {
	var values []interface{}
	values = append(values, "v1")
	values = append(values, "v2")
	values = append(values, "v3")

}

// go test -v -run Test_BuildWhere5_like ./
// 自动会绑上两边都有的%
func Test_BuildWhere5_like(t *testing.T) {
	where := map[string]interface{}{
		"name like2": "123",
		"_lockMode":  "exclusive",
	}

	whereStr, values, err := BuildWhere(where)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(whereStr) //   (name LIKE ?)
	fmt.Println(values)   // [%123%]
}
