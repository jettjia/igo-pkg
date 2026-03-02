package php2go

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// isInt 变量是否整型数值.
func isInt(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case string:
		str := val.(string)
		if str == "" {
			return false
		}
		_, err := strconv.Atoi(str)
		return err == nil
	}

	return false
}

// toInt 强制将变量转换为整型.
// 数值类型将转为整型;
// 字符串将使用str2Int;
// 布尔型的true为1,false为0;
// 数组、切片、字典、通道类型将取它们的长度;
// 指针、结构体类型为1,其他为0.
func toInt(val interface{}) (res int) {
	switch val.(type) {
	case int:
		res = val.(int)
	case int8:
		res = int(val.(int8))
	case int16:
		res = int(val.(int16))
	case int32:
		res = int(val.(int32))
	case int64:
		res = int(val.(int64))
	case uint:
		res = int(val.(uint))
	case uint8:
		res = int(val.(uint8))
	case uint16:
		res = int(val.(uint16))
	case uint32:
		res = int(val.(uint32))
	case uint64:
		res = int(val.(uint64))
	case float32:
		res = int(val.(float32))
	case float64:
		res = int(val.(float64))
	case string:
		res = str2Int(val.(string))
	case bool:
		res = bool2Int(val.(bool))
	default:
		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
			res = v.Len()
		case reflect.Ptr, reflect.Struct:
			res = 1
		}
	}

	return
}

// str2Int 将字符串转换为int.其中"true", "TRUE", "True"为1;若为浮点字符串,则取整数部分.
func str2Int(val string) (res int) {
	if val == "true" || val == "TRUE" || val == "True" {
		res = 1
		return
	} else if ok := regexp.MustCompile(`^([+-]?\d+)(\.\d+)$`).MatchString(val); ok {
		fl, _ := strconv.ParseFloat(val, 1)
		res = int(fl)
		return
	}

	res, _ = strconv.Atoi(val)
	return
}

// toStr 强制将变量转换为字符串.
func toStr(val interface{}) string {
	//先处理其他类型
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Invalid:
		return ""
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Ptr, reflect.Struct, reflect.Map: //指针、结构体和字典
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return ""
		}
		return string(b)
	}

	//再处理字节切片
	switch val.(type) {
	case []uint8:
		return string(val.([]uint8))
	}

	return fmt.Sprintf("%v", val)
}

// md5Byte 计算字节切片的 MD5 散列值.
func md5Byte(str []byte, length uint8) []byte {
	var res []byte
	h := md5.New()
	_, err := h.Write(str)
	if err == nil {
		hashInBytes := h.Sum(nil)
		dst := make([]byte, hex.EncodedLen(len(hashInBytes)))
		hex.Encode(dst, hashInBytes)
		if length > 0 && length < 32 {
			res = dst[:length]
		} else {
			res = dst
		}
	}

	return res
}

// reflect2Itf 将反射值转为接口(原值)
func reflect2Itf(r reflect.Value) (res interface{}) {
	switch r.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		res = r.Int()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		res = r.Uint()
	case reflect.Float32, reflect.Float64:
		res = r.Float()
	case reflect.String:
		res = r.String()
	case reflect.Bool:
		res = r.Bool()
	default:
		if r.CanInterface() {
			res = r.Interface()
		} else {
			res = r
		}
	}

	return
}

// bool2Int 将布尔值转换为整型.
func bool2Int(val bool) int {
	if val {
		return 1
	}
	return 0
}

// arrayValues 返回arr(数组/切片/字典/结构体)中所有的值.
// filterZero 是否过滤零值元素(nil,false,0,”,[]),true时排除零值元素,false时保留零值元素.
func arrayValues(arr interface{}, filterZero bool) []interface{} {
	var res []interface{}
	var fieldVal reflect.Value
	val := reflect.ValueOf(arr)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			fieldVal = val.Index(i)
			if !filterZero || (filterZero && !fieldVal.IsZero()) {
				res = append(res, fieldVal.Interface())
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			fieldVal = val.MapIndex(k)
			if !filterZero || (filterZero && !fieldVal.IsZero()) {
				res = append(res, fieldVal.Interface())
			}
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			fieldVal = val.Field(i)
			if fieldVal.CanInterface() {
				if !filterZero || (filterZero && !fieldVal.IsZero()) {
					res = append(res, fieldVal.Interface())
				}
			}
		}
	default:
		panic("[arrayValues]`arr type must be php2go|slice|map|struct; but : " + val.Kind().String())
	}

	return res
}

// GetFieldValue 获取(字典/结构体的)字段值;fieldName为字段名,大小写敏感.
func GetFieldValue(arr interface{}, fieldName string) (res interface{}, err error) {
	val := reflect.ValueOf(arr)
	switch val.Kind() {
	case reflect.Map:
		for _, subKey := range val.MapKeys() {
			if fmt.Sprintf("%s", subKey) == fieldName {
				res = val.MapIndex(subKey).Interface()
				break
			}
		}
	case reflect.Struct:
		field := val.FieldByName(fieldName)
		if !field.IsValid() || !field.CanInterface() {
			break
		}
		res = field.Interface()
	default:
		err = errors.New("[GetFieldValue]`arr type must be map|struct; but : " + val.Kind().String())
	}

	return
}

// 泛型：判断元素是否在切片中
func InSlice[T comparable](needle T, haystack []T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// 泛型：去重
func UniqueSlice[T comparable](arr []T) []T {
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

// 泛型版：返回切片所有值
func SliceValues[T any](arr []T) []T {
	return arr
}

// 泛型版：返回 map 所有值
func MapValues[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

func SliceColumn[T any, V any](arr []T, getter func(T) V) []V {
	res := make([]V, 0, len(arr))
	for _, item := range arr {
		res = append(res, getter(item))
	}
	return res
}
