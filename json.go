// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import "encoding/json"

// Unmarshal unstructured JSON
//
//
// Example using the following JSON string
// {
// 	"str": "FirstString",
// 	"int": 120,
// 	"bool": true,
// 	"array": ["5", "10"],
// 	"object": {
// 		"str": "SecondString",
// 		"int": 121,
// 		"bool": false,
// 		"array": ["15", "20"],
// 		"object": {
// 			"str": "ThirdString",
// 			"int": 122,
// 			"bool": true,
// 			"array": ["25", "30"]
// 		}
// 	}
// }
//
// j, _ := PrepareJSON([]byte(`{ "str": "FirstString", "int": 120, "bool": true, "array": ["5", "10"], "object": { "str": "SecondString", "int": 121, "bool": false, "array": ["15", "20"], "object": { "str": "ThirdString", "int": 122, "bool": true, "array": ["25", "30"] } } }`))
// fmt.Printf("%v\n", j.Get("object", "str"))
//
// Expected output: SecondString

type (
	// JSON contains a generic map for output.
	JSON map[string]interface{}

	// JSONValue contains values from JSONData.
	JSONValue struct {
		data interface{}
	}

	// JSONData contains the array map to values.
	JSONData struct {
		JSONValue
		m map[string]*JSONValue
	}
)

// String returns the value as string.
func (v *JSONValue) String() string {
	if v == nil {
		return ""
	}

	if _, ok := v.data.(string); ok {
		return v.data.(string)
	}
	return ""
}

// Bool returns the value as string.
func (v *JSONValue) Bool() bool {
	if v == nil {
		return false
	}

	if _, ok := v.data.(bool); ok {
		return v.data.(bool)
	}
	return false
}

// Float64 returns the value as float64.
func (v *JSONValue) Float64() float64 {
	if v == nil {
		return 0
	}

	if _, ok := v.data.(float64); ok {
		return v.data.(float64)
	}
	return 0
}

// Int64 converts and returns the value as int64.
func (v *JSONValue) Int64() int64 {
	if v == nil {
		return 0
	}

	if _, ok := v.data.(float64); ok {
		return int64(v.data.(float64))
	}
	return 0
}

// Array returns the values as an array.
func (v *JSONValue) Array() []interface{} {
	if v == nil {
		return nil
	}

	if _, ok := v.data.([]interface{}); ok {
		return v.data.([]interface{})
	}
	return nil
}

// Map returns the values as a map[string]interface{}.
func (v *JSONValue) Map() map[string]interface{} {
	if v == nil {
		return nil
	}

	if _, ok := v.data.(map[string]interface{}); ok {
		return v.data.(map[string]interface{})
	}
	return nil
}

// Get a value using the specified keys, returns nil if value is not found.
func (d *JSONData) Get(keys ...string) *JSONValue {
	cv := d

	for _, k := range keys {
		// Key does not exists
		if cv.m[k] == nil {
			continue
		}

		switch cv.m[k].data.(type) {
		case *JSONData:
			cv = cv.m[k].data.(*JSONData)
		case bool:
			return &JSONValue{data: cv.m[k].data}
		case float64:
			return &JSONValue{data: cv.m[k].data}
		case string:
			return &JSONValue{data: cv.m[k].data}
		case []interface{}:
			return &JSONValue{data: cv.m[k].data}
		default:
			return nil
		}
	}

	return nil
}

// Set a key with specified value.
func (d *JSONData) Set(key string, value interface{}) {
	d.m[key] = &JSONValue{
		data: value,
	}
}

// PrepareJSON instantiate the JSONData and unmarshal the data json, returns a JSONData struct.
func PrepareJSON(data []byte) (*JSONData, error) {
	var j = &JSONData{
		m: make(map[string]*JSONValue),
	}

	err := json.Unmarshal(data, &j.data)
	if err != nil {
		return nil, err
	}

	for k, v := range j.data.(map[string]interface{}) {
		switch v.(type) {
		case map[string]interface{}:
			j.m[k] = j.PrepareInterface(v)
		default:
			j.m[k] = &JSONValue{data: v}
		}
	}

	return j, nil
}

// PrepareInterface is a recursive function to translate json objects to maps.
func (d *JSONData) PrepareInterface(data interface{}) *JSONValue {
	out := &JSONData{
		m: make(map[string]*JSONValue),
	}

	out.data = data

	for mk, mv := range data.(map[string]interface{}) {
		switch mv.(type) {
		case map[string]interface{}:
			out.m[mk] = d.PrepareInterface(mv)
		default:
			out.m[mk] = &JSONValue{data: mv}
		}
	}

	return &JSONValue{data: out}
}
