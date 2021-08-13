package core

import (
	"sort"
	"strconv"
)

func parseNum(value interface{}) int {
	if n, ok := value.(float64); ok {
		return int(n)
	}
	if n, ok := value.(int); ok {
		return n
	}
	return 0
}

func parseString(value interface{}) string {
	if n, ok := value.(float64); ok {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	if n, ok := value.(int); ok {
		return string(n)
	}
	if n, ok := value.(string); ok {
		return "'" + n + "'"
	}
	return ""
}

func parseIntListNum(value interface{}) (inList []int) {
	if intList, ok := value.([]interface{}); ok {
		for _, v := range intList {
			inList = append(inList, parseNum(v))
		}

	}
	return inList
}

func parseIntListString(value interface{}) (inList []string) {
	// var inList []string
	if intList, ok := value.([]interface{}); ok {
		for _, v := range intList {
			inList = append(inList, parseString(v))
		}

	}
	return inList
}

// map ASCII  排序
func sortedMap(m map[string]interface{}, f func(k string, v interface{})) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}
