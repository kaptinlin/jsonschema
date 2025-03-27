package jsonschema

import "encoding/json"

func GetDefaultCompiler() *Compiler {
	compiler := NewCompiler()

	compiler.AssertFormat = true
	return compiler
}

func AnyToJSONString(any interface{}) string {
	json, err := json.Marshal(any)
	if err != nil {
		return ""
	}
	return string(json)
}
