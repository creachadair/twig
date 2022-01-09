// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/users"
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

// ResolveID checks a slice of user specifications and attempts to resolve any
// that begin with "@" with their user ID.
func ResolveID(ctx context.Context, cli *twitter.Client, specs []string) ([]string, error) {
	var ids, names []string
	for _, spec := range specs {
		if strings.HasPrefix(spec, "@") {
			names = append(names, strings.TrimPrefix(spec, "@"))
		} else {
			ids = append(ids, spec)
		}
	}
	if len(names) != 0 {
		rsp, err := users.LookupByName(names[0], &users.LookupOpts{
			More: names[1:],
		}).Invoke(ctx, cli)
		if err != nil {
			return nil, err
		}
		for _, u := range rsp.Users {
			ids = append(ids, u.ID)
		}
	}
	return ids, nil
}
