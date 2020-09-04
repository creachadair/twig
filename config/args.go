package config

import (
	"strings"

	"github.com/creachadair/twitter/types"
)

// ParseArgs decodes an argument list consisting of IDs or names mixed with
// field specifiers and expansions. A field spec has the form "name:value",
// where name is the object type (for example "tweet", "user"), and value is an
// arbitrary string. An expansion has the form "@name".
//
// If dtype != "", a spec of the form ":value" is treated as "dtype:value".
func ParseArgs(args []string, dtype string) ParsedArgs {
	var parsed ParsedArgs
	var expand types.Expansions
	fieldMap := make(map[string][]string)

	for _, arg := range args {
		// @foo is an expansion
		if exp := strings.TrimPrefix(arg, "@"); exp != arg {
			expand = append(expand, exp)
			continue
		}

		// name:field is a field spec; everything else is a query key.
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 1 {
			parsed.Keys = append(parsed.Keys, arg)
			continue
		}
		if parts[0] == "" {
			parts[0] = dtype
		}
		fieldMap[parts[0]] = append(fieldMap[parts[0]], parts[1])
	}
	if len(expand) != 0 {
		parsed.Fields = append(parsed.Fields, expand)
	}
	for key, vals := range fieldMap {
		parsed.Fields = append(parsed.Fields, types.MiscFields{
			Label_:  key + ".fields",
			Values_: vals,
		})
	}
	return parsed
}

// ParsedArgs is the result from a call to ParseArgs.
type ParsedArgs struct {
	Keys   []string // all arguments that are not field specs, in the order given
	Fields []types.Fields
}
