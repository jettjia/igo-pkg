package php2go

// SliceIntersectArray 求两个切片的交集
func SliceIntersectArray(a []string, b []string) []string {
	var inter []string
	mp := make(map[string]bool)

	for _, s := range a {
		if _, ok := mp[s]; !ok {
			mp[s] = true
		}
	}
	for _, s := range b {
		if _, ok := mp[s]; ok {
			inter = append(inter, s)
		}
	}

	return inter
}

// SliceDiffArray 求两个切片的差集
func SliceDiffArray(a []string, b []string) []string {
	var diffArray []string
	temp := map[string]struct{}{}

	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}

// SliceRemoveRepeatedElement 切片去重实现
func SliceRemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// SliceArrayUnique 切片去重实现
func SliceArrayUnique(arr []string) []string {
	result := make([]string, 0, len(arr))
	temp := map[string]struct{}{}
	for i := 0; i < len(arr); i++ {
		if _, ok := temp[arr[i]]; ok != true {
			temp[arr[i]] = struct{}{}
			result = append(result, arr[i])
		}
	}
	return result
}

// SliceInArray in array
func SliceInArray(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
