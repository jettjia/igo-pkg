package builder

import (
	"errors"
)

func GormBuildWhere(queries []*Query) (whereStr string, values []interface{}, err error) {
	whereReq := make(map[string]interface{})

	if len(queries) == 0 {
		return
	}

	for _, val := range queries {
		if val.Key == "" {
			err = errors.New("the query key is incorrect")
			return
		}
		whereReq[val.Key+" "+OperatorMap[val.Operator]] = val.Value
	}

	return BuildWhere(whereReq)
}

func GormBuildWhereMap(queries []*Query) (whereMap map[string]interface{}) {
	whereMap = make(map[string]interface{})

	if len(queries) == 0 {
		return
	}

	for _, val := range queries {
		whereMap[val.Key+" "+OperatorMap[val.Operator]] = val.Value
	}

	return
}
