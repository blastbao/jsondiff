package jsondiff

import (
	"fmt"
	"reflect"
	"strings"
)

type Config struct {
	MaxDiff     int
	MaxDeep     int
	SortedSlice bool

	// HasExceptedField bool
	ExceptedFields map[int]map[string]int
}

type Differ struct {
	Conf Config
	diff []string
	buff []string
}

func New() *Differ {
	exceptedFields := make(map[int]map[string]int, 0)
	return &Differ{
		Conf: Config{
			MaxDiff:     10,
			MaxDeep:     10,
			SortedSlice: true,
			// HasExceptedField: false,
			ExceptedFields: exceptedFields,
		},
		diff: []string{},
		buff: []string{},
	}
}

func (d *Differ) Compare(expected, actual map[string]interface{}) []string {
	d.compareMap(expected, actual, 1)
	return d.diff
}

func (d *Differ) compareMap(src, tgt map[string]interface{}, deep int) {

	// 第 1 层排除的 keys
	exceptedKeys := d.Conf.ExceptedFields[deep]
	// 全局排除的 keys
	globalExceptedKeys := d.Conf.ExceptedFields[-1]

	// 遍历 src
	for sk, sv := range src {

		// Ignore
		if _, ok := globalExceptedKeys[sk]; ok {
			continue
		}
		if _, ok := exceptedKeys[sk]; ok {
			continue
		}

		// 检查是否存在于 tgt
		tv := tgt[sk]

		// 存在
		if tv != nil {
			// 路径入栈
			d.push(fmt.Sprintf("map[%s]", sk))
			// 深度优先搜索，下沉一层
			d.compareVal(sv, tv, deep+1)
			// 路径出栈
			d.pop()
		// 不存在
		} else {
			d.saveDiff(sv, "<nil>")
		}

	}

	if len(src) == len(tgt) {
		return
	}

	for k, v := range tgt {
		if _, ok := globalExceptedKeys[k]; ok {
			continue
		}
		if _, ok := exceptedKeys[k]; ok {
			continue
		}

		sv := src[k]
		if sv == nil {
			d.saveDiff("<nil>", v)
		}
	}
}

func (d *Differ) compareArray(expected, actual []interface{}, deep int) {
	expectedLen := len(expected)
	actualLen := len(actual)
	maxLen := expectedLen
	if actualLen > maxLen {
		maxLen = actualLen
	}

	for i := 0; i < maxLen; i++ {
		d.push(fmt.Sprintf("array[%d]", i))
		if i < expectedLen && i < actualLen {
			d.compareVal(expected[i], actual[i], deep+1)

		} else if i < expectedLen {
			d.saveDiff(expected[i], "<nil>")
		} else {
			d.saveDiff("<nil>", actual[i])
		}
		d.pop()
	}
}

func (d *Differ) compareVal(expectedVal, actualVal interface{}, deep int) {
	if deep > d.Conf.MaxDeep || len(d.diff) >= d.Conf.MaxDiff {
		return
	}

	expectedType := reflect.TypeOf(expectedVal)
	actualType := reflect.TypeOf(actualVal)

	// 类型不同，保存差异
	if expectedType != actualType {
		d.saveDiff(expectedType, actualType)
		return
	}

	// 类型相同，分类型处理
	switch expectedVal.(type) {
	case map[string]interface{}:
		if deep == d.Conf.MaxDeep && !reflect.DeepEqual(expectedVal, actualVal) {
			d.saveUnReadDiff("map")
		}
		d.compareMap(expectedVal.(map[string]interface{}), actualVal.(map[string]interface{}), deep)
	case []interface{}:
		if deep == d.Conf.MaxDeep && !reflect.DeepEqual(expectedVal, actualVal) {
			d.saveUnReadDiff("array")
		}
		d.compareArray(expectedVal.([]interface{}), actualVal.([]interface{}), deep)
	default:
		if !reflect.DeepEqual(expectedVal, actualVal) {
			d.saveDiff(expectedVal, actualVal)
		}
	}
}

func (d *Differ) saveDiff(expectedVal, actualVal interface{}) {
	if len(d.diff) >= d.Conf.MaxDiff {
		return
	}
	if len(d.buff) > 0 {
		path := strings.Join(d.buff, ".")
		d.diff = append(d.diff, fmt.Sprintf("%s: %v != %v", path, expectedVal, actualVal))
	} else {
		d.diff = append(d.diff, fmt.Sprintf("%v != %v", expectedVal, actualVal))
	}
}

func (d *Differ) saveUnReadDiff(typeStr string) {
	if len(d.diff) >= d.Conf.MaxDiff {
		return
	}
	if len(d.buff) > 0 {
		path := strings.Join(d.buff, ".")
		d.diff = append(d.diff, fmt.Sprintf("%s: %s different", path, typeStr))
	} else {
		d.diff = append(d.diff, fmt.Sprintf("%s different", typeStr))
	}
}

func (d *Differ) push(str string) {
	d.buff = append(d.buff, str)
}

func (d *Differ) pop() {
	if len(d.buff) > 0 {
		d.buff = d.buff[0 : len(d.buff)-1]
	}
}

func (d *Differ) AddExpectedField(key string, deep int) {
	if deep <= 0 {
		deep = -1
	}
	keysMap := d.Conf.ExceptedFields[deep]
	if keysMap == nil {
		newMap := make(map[string]int, 0)
		newMap[key] = 1
		d.Conf.ExceptedFields[deep] = newMap
	} else {
		keysMap[key] = 1
		d.Conf.ExceptedFields[deep] = keysMap
	}
}

// func (d *Differ) RemoveExectedField(key string, deep int) {
// 	//TODO
// }
