package php2go

import (
	"strings"
)

const (
	// COMPARE_ONLY_VALUE 仅比较值
	COMPARE_ONLY_VALUE LkkArrCompareType = 0
	// COMPARE_ONLY_KEY 仅比较键
	COMPARE_ONLY_KEY LkkArrCompareType = 1
	// COMPARE_BOTH_KEYVALUE 同时比较键和值
	COMPARE_BOTH_KEYVALUE LkkArrCompareType = 2
)

type (
	// LkkArrCompareType 枚举类型,数组比较方式
	LkkArrCompareType uint8
)

// ArrayColumn 返回数组(切片/字典/结构体)中元素指定的一列.
// arr的元素必须是字典;
// columnKey为元素的字段名;
// 该方法效率较低.
func ArrayColumn[T any, V any](arr []T, getter func(T) V) []V {
	res := make([]V, 0, len(arr))
	for _, item := range arr {
		res = append(res, getter(item))
	}
	return res
}

// JoinStrings 使用分隔符delimiter连接字符串切片strs.效率比Implode高.
func JoinStrings(strs []string, delimiter string) string {
	if len(strs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, s := range strs {
		if i > 0 {
			sb.WriteString(delimiter)
		}
		sb.WriteString(s)
	}
	return sb.String()
}

// ArrayDiff 比较两个切片，返回在 a 中但不在 b 中的元素
func ArrayDiff[T comparable](a, b []T) []T {
	m := make(map[T]struct{}, len(b))
	for _, v := range b {
		m[v] = struct{}{}
	}
	var diff []T
	for _, v := range a {
		if _, ok := m[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// ArrayDiffByColumnGeneric 返回在 b 中但不在 a 中的元素，按 getter 指定的字段比较
func ArrayDiffByColumnGeneric[T any, K comparable](a, b []T, getter func(T) K) []T {
	m := make(map[K]struct{}, len(a))
	for _, v := range a {
		m[getter(v)] = struct{}{}
	}
	var diff []T
	for _, v := range b {
		if _, ok := m[getter(v)]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// ArrayIntersect 返回 a 和 b 的交集
func ArrayIntersect[T comparable](a, b []T) []T {
	m := make(map[T]struct{}, len(b))
	for _, v := range b {
		m[v] = struct{}{}
	}
	var inter []T
	for _, v := range a {
		if _, ok := m[v]; ok {
			inter = append(inter, v)
		}
	}
	return inter
}

// ArrayUnique 移除数组(切片/字典)中重复的值,返回字典,保留键名.
func ArrayUnique[T comparable](arr []T) []T {
	m := make(map[T]struct{}, len(arr))
	var res []T
	for _, v := range arr {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}

// InArray 元素needle是否在数组haystack(切片/字典)内.
func InArray[T comparable](needle T, haystack []T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// ArrayKeys 返回数组(切片/字典/结构体)中所有的键名;如果是结构体,只返回公开的字段.
func ArrayKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ArrayValues 返回arr(数组/切片/字典/结构体)中所有的值;如果是结构体,只返回公开字段的值.
// filterZero 是否过滤零值元素(nil,false,0,"",[]),true时排除零值元素,false时保留零值元素.
func ArrayValues[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

// IsEqualArray 判断两个切片内容是否相同（顺序无关，元素类型需可比较）
func IsEqualArray[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	countA := make(map[T]int, len(a))
	countB := make(map[T]int, len(b))
	for _, v := range a {
		countA[v]++
	}
	for _, v := range b {
		countB[v]++
	}
	if len(countA) != len(countB) {
		return false
	}
	for k, v := range countA {
		if countB[k] != v {
			return false
		}
	}
	return true
}

// ArrayFlip 交换数组(切片/字典)中的键和值.
func ArrayFlip[T comparable](arr []T) map[T]int {
	res := make(map[T]int, len(arr))
	for i, v := range arr {
		res[v] = i
	}
	return res
}

// ArrayCombine 将key和value组成对应的map
func ArrayCombine[K comparable, V any](keys []K, values []V) map[K]V {
	if len(keys) != len(values) {
		panic("keys and values must have the same length")
	}
	m := make(map[K]V, len(keys))
	for i, k := range keys {
		m[k] = values[i]
	}
	return m
}

// ArrayMerge 所有的数组选项会合并到一个数组中，具有相同键名的值不会被覆盖
func ArrayMerge[T any](slices ...[]T) []T {
	var total int
	for _, s := range slices {
		total += len(s)
	}
	result := make([]T, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
