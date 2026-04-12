package util

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PrintJson print json
func PrintJson(args interface{}) string {
	b, err := json.Marshal(args)
	if err != nil {
		return fmt.Sprintf("%+v", args)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", args)
	}
	return out.String()
}
