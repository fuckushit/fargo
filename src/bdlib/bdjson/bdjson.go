package bdjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Object json object
type Object map[string]interface{}

// Array json array
type Array []interface{}

// Version 返回版本信息
func Version() string {
	return "0.0.1"
}

// JSON json 操作对象.
type JSON struct {
	data interface{}
}

// NewJSON 新建json对象.
func NewJSON(body []byte) (*JSON, error) {
	json := new(JSON)
	err := json.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// New 从data中新建json
func New(data interface{}) *JSON {
	return &JSON{data}
}

// Encode json编码.
func (j *JSON) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

// MarshalJSON json marshal.
func (j *JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// UnmarshalJSON json解码.
func (j *JSON) UnmarshalJSON(body []byte) error {
	return json.Unmarshal(body, &j.data)
}

// Map JSON对象格式化,
// JSON对象转化map[string]interface{}.
func (j *JSON) Map() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// MustPhpMap 将php的array转化为map，如果php的array是空的，就不是obj而是空的array，本函数自动处理这个问题
func (j *JSON) MustPhpMap() map[string]interface{} {
	m, err := j.Map()
	if err != nil {
		// 尝试转换php的array为map
		arr, err := j.Slice()
		if err != nil {
			m = map[string]interface{}{}
			return m
		}
		m = map[string]interface{}{}
		for idx, s := range arr {
			m[fmt.Sprintf("%d", idx)] = s
		}
		return m
	}

	return m
}

// Slice JSON对象转化[]interface{}.
func (j *JSON) Slice() ([]interface{}, error) {
	if s, ok := (j.data).([]interface{}); ok {
		return s, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

//// MustSlice JSON对象转化[]interface{}.
//func (j *JSON) MustSlice() []interface{} {
//	return j.data.([]interface{})
//}

// Bool JSON对象转化bool.
func (j *JSON) Bool() (bool, error) {
	if b, ok := (j.data).(bool); ok {
		return b, nil
	}
	return false, errors.New("type assertion to bool failed")
}

//// MustBool JSON对象转化bool.
//func (j *JSON) MustBool() bool {
//	return j.data.(bool)
//}

// MustPhpBool JSON转化为bool
func (j *JSON) MustPhpBool() bool {
	b, ok := j.data.(bool)
	if !ok {
		str := fmt.Sprintf("%v", j.data)
		if str == "true" || str == "True" {
			return true
		}
		if str == "false" || str == "False" {
			return false
		}
	}
	return b
}

// String JSON对象转化string.
func (j *JSON) String() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

//// MustString  JSON对象转化string
//func (j *JSON) MustString() string {
//	return j.data.(string)
//}

// MustPhpString JSON转化为string，如果不是string，则转化为string
func (j *JSON) MustPhpString() string {
	switch val := j.data.(type) {
	case string:
		return val
	case int, int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%f", val)
	}

	if j.data == nil {
		return ""
	}
	return fmt.Sprintf("%v", j.data)
}

// MustPhpInt convert to int64
func (j *JSON) MustPhpInt() int64 {
	switch val := j.data.(type) {
	case float64:
		return int64(val)
	case string:
		id, _ := strconv.ParseInt(val, 10, 64)
		return id
	case int64:
		return val
	case int:
		return int64(val)
	}
	return 0
}

// MustPhpArray 转换成数组，屏蔽php返回的array有时候是string，有时候是object的问题
func (j *JSON) MustPhpArray() (arr []interface{}) {
	var ok bool
	if arr, ok = j.data.([]interface{}); ok {
		return
	}
	mp := j.MustPhpMap()
	arr = make([]interface{}, 0)
	for _, v := range mp {
		arr = append(arr, v)
	}
	return
}

// MustPhpStringArray 转换成字符数组，屏蔽php返回的array有时候是string，有时候是object的问题
func (j *JSON) MustPhpStringArray() (sarr []string) {
	arr := j.MustPhpArray()
	sarr = make([]string, 0)
	for _, v := range arr {
		if s, ok := v.(string); ok {
			sarr = append(sarr, s)
		}
	}
	return
}

// Float64 JSON对象转化float64.
func (j *JSON) Float64() (float64, error) {
	if f, ok := (j.data).(float64); ok {
		return f, nil
	}
	return -1, errors.New("type assertion to float64 failed")
}

//// MustFloat JSON对象转化float64.
//func (j *JSON) MustFloat() float64 {
//	return j.data.(float64)
//}

// MustPhpFloat convert to float64
func (j *JSON) MustPhpFloat() float64 {
	switch val := j.data.(type) {
	case float32:
		return float64(val)
	case float64:
		return val
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		id, _ := strconv.ParseFloat(val, 64)
		return id
	}
	return 0
}

// Int JSON对象转化int.
func (j *JSON) Int() (int, error) {
	if f, ok := (j.data).(float64); ok {
		return int(f), nil
	}
	return -1, errors.New("type assertion to int failed")
}

// Int64 JSON对象转化int64.
func (j *JSON) Int64() (int64, error) {
	if f, ok := (j.data).(float64); ok {
		return int64(f), nil
	}
	return -1, errors.New("type assertion to int64 failed")
}

//// MustInt64 JSON对象转化int64.
//func (j *JSON) MustInt64() int64 {
//	num := j.data.(float64)
//	return int64(num)
//}

// Bytes JSON对象转化[]byte.
func (j *JSON) Bytes() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// StringSlice JSON对象转化[]string.
func (j *JSON) StringSlice() ([]string, error) {
	slice, err := j.Slice()
	if err != nil {
		return nil, err
	}
	var stringSlice = make([]string, 0)
	for _, v := range slice {
		ss, ok := v.(string)
		if !ok {
			return nil, errors.New("type assertion to []string failed")
		}
		stringSlice = append(stringSlice, ss)
	}
	return stringSlice, nil
}

// Set 获取json内数据.
func (j *JSON) Set(key string, val interface{}) (err error) {
	m, err := j.Map()
	if err != nil {
		return
	}
	m[key] = val
	return
}

// Get 可使用连贯操作,
// js.Get("top_level").Get("dict").Get("value").Int(),
// 获得json内数据.
func (j *JSON) Get(key string) *JSON {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &JSON{val}
		}
		return &JSON{nil}
	}
	return &JSON{nil}
}

// GetIndex 获取 json encode 出来的 slice 的下标对应的值.
func (j *JSON) GetIndex(index int) *JSON {
	s, err := j.Slice()
	if err == nil {
		if len(s) > index {
			return &JSON{s[index]}
		}
		return &JSON{nil}
	}
	return &JSON{nil}
}

// CheckGet check map 的值是否正确.
func (j *JSON) CheckGet(key string) (*JSON, bool) {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &JSON{val}, true
		}
	}
	return nil, false
}

// GetLot 纵向获得层次数据,
// js.GetLot("top_level", "dict").
func (j *JSON) GetLot(key ...string) *JSON {
	jin := j
	for i := range key {
		m, err := jin.Map()
		if err != nil {
			return &JSON{nil}
		}
		if val, ok := m[key[i]]; ok {
			jin = &JSON{val}
		} else {
			return &JSON{nil}
		}
	}
	return jin
}

// Data 返回data
func (j *JSON) Data() interface{} {
	return j.data
}

// SetData 设置data
func (j *JSON) SetData(data interface{}) {
	j.data = data
}
