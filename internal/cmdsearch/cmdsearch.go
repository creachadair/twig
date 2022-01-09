// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdsearch

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
)

var Command = &command.C{
	Name:  "search",
	Usage: "[-max n] -query query [field-spec...]",
	Help: `
Search for recent tweets matching the specified query.

A field spec has the form type:field, e.g., "tweet:entities".
As a special case, :field is shorthand for "tweet:field".

If the results span multiple pages, use -page to set the
page token to resume searching from.
`,
	SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
		fs.StringVar(&opts.query, "query", "", "Search query (required)")
		fs.IntVar(&opts.maxResults, "max", 0, "Maximum results to request (0 means all)")
		fs.StringVar(&opts.sinceID, "after", "", "Return tweets (strictly) after this ID")
		fs.StringVar(&opts.untilID, "before", "", "Return tweets (strictly) before this ID")
		fs.Var(timestamp{&opts.since}, "since", "Return tweets no older than this")
		fs.Var(timestamp{&opts.until}, "until", "Return tweets no newer than this")
	},

	Run: func(env *command.Env, args []string) error {
		if opts.query == "" {
			fmt.Fprintln(env, "Error: a search -query must be set")
			return command.FailWithUsage(env, args)
		}
		parsed := config.ParseArgs(args, "tweet")
		if len(parsed.Keys) != 0 {
			fmt.Fprintf(env, "Error: extra arguments after query %v\n", parsed.Keys)
			return command.FailWithUsage(env, args)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		// Choose a page size based on the requested max results.  The service
		// will not accept arbitrary sizes, so we'll paginate based on what the
		// user requested.
		max := opts.maxResults
		if max <= 0 || max > 100 {
			max = 100
		} else if max < 10 {
			max = 10
		}

		q := tweets.SearchRecent(opts.query, &tweets.SearchOpts{
			StartTime:  opts.since,
			EndTime:    opts.until,
			MaxResults: max,
			SinceID:    opts.sinceID,
			UntilID:    opts.untilID,
			Optional:   parsed.Fields,
		})
		var numResults int
		for q.HasMorePages() {
			rsp, err := q.Invoke(context.Background(), cli)
			if err != nil {
				return err
			}
			for _, tw := range rsp.Tweets {
				numResults++
				data, err := json.Marshal(tw)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				if opts.maxResults > 0 && numResults >= opts.maxResults {
					return nil // our work is complete
				}
			}
		}
		return nil
	},
}

var opts struct {
	maxResults int
	sinceID    string
	untilID    string
	since      time.Time
	until      time.Time
	query      string
}

type timestamp struct {
	*time.Time
}

func (ts timestamp) Set(s string) error {
	t, err := time.Parse(types.DateFormat, s)
	if err != nil {
		return err
	}
	*ts.Time = t
	return nil
}

func (ts timestamp) String() string {
	if ts.Time == nil {
		// This value is never shown to the user. It averts a panic in the flag
		// package which constructs a zero value to call its String method.
		return ""
	} else if ts.Time.IsZero() {
		return types.DateFormat
	}
	return ts.Time.Format(types.DateFormat)
}

// Get satisfies flag.Getter, the concrete type is time.Time.
func (ts timestamp) Get() interface{} { return *ts.Time }
