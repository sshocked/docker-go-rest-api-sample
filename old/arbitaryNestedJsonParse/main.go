package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var nodes = 0
var summ = 0

func main() {
	// Creating the maps for JSON
	m := map[string]interface{}{}

	// Parsing/Unmarshalling JSON encoding/json
	err := json.Unmarshal([]byte(input), &m)

	if err != nil {
		panic(err)
	}
	parseMap(m)
	fmt.Println("Nodes: ", nodes)
	fmt.Println("Summ: ", summ)
}

func parseMap(aMap map[string]interface{}) {
	for key, val := range aMap {
		nodes++
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			fmt.Println(key)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			fmt.Println(key)
			parseArray(val.([]interface{}))
		default:
			i, err := strconv.Atoi(fmt.Sprintf("%v", concreteVal))
			if err == nil {
				summ += i
			}
			fmt.Println(key, ":", concreteVal)
		}
	}
}

func parseArray(anArray []interface{}) {
	for i, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			fmt.Println("Index:", i)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			fmt.Println("Index:", i)
			parseArray(val.([]interface{}))
		default:
			fmt.Println("Index", i, ":", concreteVal)
		}
	}
}

const input = `
{
    "outterJSON": {
        "innerJSON1": {
            "value1": 10,
            "value2": {
				"name": "John",
				"age": 32
			},
            "value3": 32,
            "InnerInnerArray": [ "test1" , "test2"],
            "InnerInnerJSONArray": [{"fld1" : "val1"} , {"fld2" : "val2"}]
        },
        "InnerJSON2":"NoneValue"
    }
}
`
