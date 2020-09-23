// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package config

import (
	"fmt"
	"strings"

	"github.com/creachadair/twitter/types"
)

// OptionSetter represents the ability to set named optional Boolean options.
// This interface is satisfied by the option types from the twitter package.
type OptionSetter interface {
	Set(string, bool) bool
}

// ParseParams parsees optional parameters by name into an options struct.
// Parameter names are prefixed with a colon, for example ":entities".
// It returns the unconsumed arguments or reports a descriptive error for an
// unknown parameter name.
func ParseParams(args []string, opt OptionSetter) ([]string, error) {
	var rest []string
	for _, arg := range args {
		trim := strings.TrimPrefix(arg, ":")
		if trim == arg {
			rest = append(rest, arg)
		} else if !opt.Set(trim, true) {
			return nil, fmt.Errorf("unknown parameter: %q", arg)
		}
	}
	return rest, nil
}

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
			if sc, ok := expShortcut[exp]; ok {
				exp = sc
			}
			expand = append(expand, exp)
			continue
		}

		// name:field is a field spec; everything else is a query key.
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 1 {
			parsed.Keys = append(parsed.Keys, arg)
			continue
		}
		switch parts[0] {
		case "":
			parts[0] = dtype
		case "m":
			parts[0] = "media"
		case "u":
			parts[0] = "user"
		case "t":
			parts[0] = "tweet"
		case "l":
			parts[0] = "place"
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

var expShortcut = map[string]string{
	"tweets":        "referenced_tweets.id",
	"ref_tweets":    "referenced_tweets.id",
	"reply_to_user": "in_reply_to_user_id",
	"media_keys":    "attachments.media_keys",
	"poll_ids":      "attachments.poll_ids",
	"place_id":      "geo.place_id",
	"mentions":      "entities.mentions.username",
	"ref_author":    "referenced_tweets.id.author_id",
	"pinned_tweet":  "pinned_tweet_id",
}
