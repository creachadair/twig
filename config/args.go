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

// ParseParams parses optional parameters by name into an options struct.
// Parameter names are prefixed with a colon, for example ":entities".
// It returns the unconsumed arguments or reports a descriptive error for an
// unknown parameter name.
func ParseParams(args []string, kind string, opt OptionSetter) ([]string, error) {
	var rest []string
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 1 {
			rest = append(rest, arg)
		} else if parts[0] != "" && parts[0] != kind {
			return nil, fmt.Errorf("wrong parameter type %q (want %q)", parts[0], kind)
		} else if !opt.Set(parts[1], true) {
			return nil, fmt.Errorf("unknown parameter: %q", arg)
		}
	}
	return rest, nil
}

// ParseArgs decodes an argument list consisting of IDs or names mixed with
// field specifiers and expansions. A field spec has the form "name:value",
// where name is the object type (for example "tweet", "user"), and value is an
// arbitrary string. An expansion has the form "+name".
//
// If dtype != "", a spec of the form ":value" is treated as "dtype:value".
func ParseArgs(args []string, dtype string) ParsedArgs {
	var parsed ParsedArgs
	var expand types.Expansions
	fieldMap := make(map[string][]string)

	for _, arg := range args {
		// @foo is an expansion
		if exp := strings.TrimPrefix(arg, "+"); exp != arg {
			if sc, ok := expShortcut[exp]; ok {
				exp = sc
			}
			expand.Set(exp, true)
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
	if expand != (types.Expansions{}) {
		parsed.Fields = append(parsed.Fields, expand)
	}
	for key, vals := range fieldMap {
		parsed.Fields = append(parsed.Fields, miscFields{
			label:  key,
			values: vals,
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
	"tweets":            "referenced_tweets.id",
	"ref_tweets":        "referenced_tweets.id",
	"referenced_tweets": "referenced_tweets.id",
	"reply_to_user":     "in_reply_to_user_id",
	"media_key":         "attachments.media_keys",
	"media_keys":        "attachments.media_keys",
	"poll_id":           "attachments.poll_ids",
	"poll_ids":          "attachments.poll_ids",
	"place_id":          "geo.place_id",
	"place_ids":         "geo.place_id",
	"mention":           "entities.mentions.username",
	"mentions":          "entities.mentions.username",
	"ref_author":        "referenced_tweets.id.author_id",
	"reference_author":  "referenced_tweets.id.author_id",
	"pinned_tweet":      "pinned_tweet_id",
}

type miscFields struct {
	label  string
	values []string
}

func (m miscFields) Label() string    { return m.label + ".fields" }
func (m miscFields) Values() []string { return m.values }
