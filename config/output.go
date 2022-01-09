// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

package config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func printJSON(v interface{}) error {
	bits, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(bits))
	return nil
}

// PrintJSON prints v as JSON to stdout. If v is a slice, the members of the
// slice are printed one by one; otherwise v is printed alone.
func PrintJSON(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := printJSON(val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}
	return printJSON(v)
}
