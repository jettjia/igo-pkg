package builder

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	errSplitEmptyKey = errors.New("[builder] couldn't split a empty string")
	// ErrUnsupportedOperator reports there's unsupported operators in where-condition
	ErrUnsupportedOperator     = errors.New("[builder] unsupported operator")
	errOrValueType             = errors.New(`[builder] the value of "_or" must be of slice of map[string]interface{} type`)
	errWhereInterfaceSliceType = `[builder] the value of "xxx %s" must be of []interface{} type`
	errEmptySliceCondition     = `[builder] the value of "%s" must contain at least one element`

	defaultIgnoreKeys = map[string]struct{}{
		"_orderby":  struct{}{},
		"_groupby":  struct{}{},
		"_having":   struct{}{},
		"_limit":    struct{}{},
		"_lockMode": struct{}{},
	}
)

type Raw string

type whereMapSet struct {
	set map[string]map[string]interface{}
}

func (w *whereMapSet) add(op, field string, val interface{}) {
	if nil == w.set {
		w.set = make(map[string]map[string]interface{})
	}
	s, ok := w.set[op]
	if !ok {
		s = make(map[string]interface{})
		w.set[op] = s
	}
	s[field] = val
}

type eleLimit struct {
	begin, step uint
}

const (
	OpEq          = "="
	OpNe1         = "!="
	OpNe2         = "<>"
	OpIn          = "in"
	OpNotIn       = "not in"
	OpGt          = ">"
	OpGte         = ">="
	OpLt          = "<"
	OpLte         = "<="
	OpLike        = "like"
	OpLikePercent = "like2"
	OpNotLike     = "not like"
	OpBetween     = "between"
	OpNotBetween  = "not between"
	// special
	OpNull = "null"
)

var opOrder = []string{OpEq, OpIn, OpNe1, OpNe2, OpNotIn, OpGt, OpGte, OpLt, OpLte, OpLike, OpLikePercent, OpNotLike, OpBetween, OpNotBetween, OpNull}

func BuildWhere(whereReq map[string]interface{}) (whereStr string, values []interface{}, err error) {
	conditions, err := getWhereConditions(whereReq, defaultIgnoreKeys)
	if nil != err {
		return
	}

	if len(conditions) == 0 {
		return
	}

	bd := strings.Builder{}
	where, _ := splitCondition(conditions)
	whereString, vals := whereConnector("AND", where...)
	if "" != whereString {
		bd.WriteString(" ")
		bd.WriteString(whereString)
	}

	return bd.String(), vals, nil
}

func getWhereConditions(where map[string]interface{}, ignoreKeys map[string]struct{}) ([]Comparable, error) {
	if len(where) == 0 {
		return nil, nil
	}
	wms := &whereMapSet{}
	var comparables []Comparable
	var field, operator string
	var err error
	for key, val := range where {
		if _, ok := ignoreKeys[key]; ok {
			continue
		}
		if strings.HasPrefix(key, "_or") {
			var (
				orWheres          []map[string]interface{}
				orWhereComparable []Comparable
				ok                bool
			)
			if orWheres, ok = val.([]map[string]interface{}); !ok {
				return nil, errOrValueType
			}
			for _, orWhere := range orWheres {
				if orWhere == nil {
					continue
				}
				orNestWhere, err := getWhereConditions(orWhere, ignoreKeys)
				if nil != err {
					return nil, err
				}
				orWhereComparable = append(orWhereComparable, NestWhere(orNestWhere))
			}
			comparables = append(comparables, OrWhere(orWhereComparable))
			continue
		}
		field, operator, err = splitKey(key, val)
		if nil != err {
			return nil, err
		}
		operator = strings.ToLower(operator)
		if !isStringInSlice(operator, opOrder) {
			return nil, ErrUnsupportedOperator
		}
		if _, ok := val.(NullType); ok {
			operator = OpNull
		}
		wms.add(operator, field, val)
	}
	whereComparables, err := buildWhereCondition(wms)
	if nil != err {
		return nil, err
	}
	comparables = append(comparables, whereComparables...)
	return comparables, nil
}

func splitKey(key string, val interface{}) (field string, operator string, err error) {
	key = strings.Trim(key, " ")
	if "" == key {
		err = errSplitEmptyKey
		return
	}
	idx := strings.IndexByte(key, ' ')
	if idx == -1 {
		field = key
		operator = "="
		if reflect.ValueOf(val).Kind() == reflect.Slice {
			operator = "in"
		}
	} else {
		field = key[:idx]
		operator = strings.Trim(key[idx+1:], " ")
		operator = removeInnerSpace(operator)
	}
	return
}

func removeInnerSpace(operator string) string {
	n := len(operator)
	firstSpace := strings.IndexByte(operator, ' ')
	if firstSpace == -1 {
		return operator
	}
	lastSpace := firstSpace
	for i := firstSpace + 1; i < n; i++ {
		if operator[i] == ' ' {
			lastSpace = i
		} else {
			break
		}
	}
	return operator[:firstSpace] + operator[lastSpace:]
}

func isStringInSlice(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

type compareProducer func(m map[string]interface{}) (Comparable, error)

var op2Comparable = map[string]compareProducer{
	OpEq: func(m map[string]interface{}) (Comparable, error) {
		return Eq(m), nil
	},
	OpNe1: func(m map[string]interface{}) (Comparable, error) {
		return Ne(m), nil
	},
	OpNe2: func(m map[string]interface{}) (Comparable, error) {
		return Ne(m), nil
	},
	OpIn: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m, OpIn)
		if nil != err {
			return nil, err
		}
		return In(wp), nil
	},
	OpNotIn: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m, OpNotIn)
		if nil != err {
			return nil, err
		}
		return NotIn(wp), nil
	},
	OpBetween: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m, OpBetween)
		if nil != err {
			return nil, err
		}
		return Between(wp), nil
	},
	OpNotBetween: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m, OpNotBetween)
		if nil != err {
			return nil, err
		}
		return NotBetween(wp), nil
	},
	OpGt: func(m map[string]interface{}) (Comparable, error) {
		return Gt(m), nil
	},
	OpGte: func(m map[string]interface{}) (Comparable, error) {
		return Gte(m), nil
	},
	OpLt: func(m map[string]interface{}) (Comparable, error) {
		return Lt(m), nil
	},
	OpLte: func(m map[string]interface{}) (Comparable, error) {
		return Lte(m), nil
	},
	OpLike: func(m map[string]interface{}) (Comparable, error) {
		return Like(m), nil
	},
	OpLikePercent: func(m map[string]interface{}) (Comparable, error) {
		return LikePercent(m), nil
	},
	OpNotLike: func(m map[string]interface{}) (Comparable, error) {
		return NotLike(m), nil
	},
	OpNull: func(m map[string]interface{}) (Comparable, error) {
		return nullCompareble(m), nil
	},
}

func buildWhereCondition(mapSet *whereMapSet) ([]Comparable, error) {
	var cpArr []Comparable
	for _, operator := range opOrder {
		whereMap, ok := mapSet.set[operator]
		if !ok {
			continue
		}
		f, ok := op2Comparable[operator]
		if !ok {
			return nil, ErrUnsupportedOperator
		}
		cp, err := f(whereMap)
		if nil != err {
			return nil, err
		}
		cpArr = append(cpArr, cp)
	}
	return cpArr, nil
}

func convertWhereMapToWhereMapSlice(where map[string]interface{}, op string) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})
	for key, val := range where {
		vals, ok := convertInterfaceToMap(val)
		if !ok {
			return nil, fmt.Errorf(errWhereInterfaceSliceType, op)
		}
		// 如果为空，不校验错误
		// if 0 == len(vals) {
		// 	return nil, fmt.Errorf(errEmptySliceCondition, op)
		// }
		result[key] = vals
	}
	return result, nil
}

func convertInterfaceToMap(val interface{}) ([]interface{}, bool) {
	s := reflect.ValueOf(val)
	if s.Kind() != reflect.Slice {
		return nil, false
	}
	interfaceSlice := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		interfaceSlice[i] = s.Index(i).Interface()
	}
	return interfaceSlice, true
}
