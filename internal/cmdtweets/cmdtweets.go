package cmdtweets

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
)

var Command = &command.C{
	Name: "tweets",
	Help: "Look up or search tweets.",
	Commands: []*command.C{
		cmdLookup,
		cmdSearch,
	},
}

var cmdLookup = &command.C{
	Name:  "lookup",
	Usage: "id/field ...",
	Help: `
Look up the specified tweets by ID.

Each argument is either a tweet ID or field specifier.
A field specifier has the form type:field, e.g., "tweet:entities".
As a special case, :field is shorthand for "tweet:field".
`,

	Run: func(ctx *command.Context, args []string) error {
		parsed := config.ParseArgs(args, "tweet")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(ctx, "Error: no tweet IDs were specified")
			return command.FailWithUsage(ctx, args)
		}

		cli, err := ctx.Config.(*config.Config).NewBearerClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		rsp, err := tweets.Lookup(parsed.Keys[0], &tweets.LookupOpts{
			More:     parsed.Keys[1:],
			Optional: parsed.Fields,
		}).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(rsp.Reply)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var cmdSearch = &command.C{
	Name:  "search",
	Usage: "[-page token] -query q [field-spec...]",
	Help: `
Search for recent tweets matching the specified query.

A field spec has the form type:field, e.g., "tweet:entities".
As a special case, :field is shorthand for "tweet:field".

If the results span multiple pages, use -page to set the
page token to resume searching from.
`,

	Run: func(ctx *command.Context, args []string) error {
		if searchQuery == "" {
			fmt.Fprintln(ctx, "Error: a search -query must be set")
			return command.FailWithUsage(ctx, args)
		}
		parsed := config.ParseArgs(args, "tweet")
		if len(parsed.Keys) != 0 {
			fmt.Fprintf(ctx, "Error: extra arguments after query %v\n", parsed.Keys)
			return command.FailWithUsage(ctx, args)
		}

		cli, err := ctx.Config.(*config.Config).NewBearerClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		searchOpts.Optional = parsed.Fields
		rsp, err := tweets.SearchRecent(searchQuery, &searchOpts).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(rsp.Reply)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var (
	searchOpts  tweets.SearchOpts
	searchQuery string
)

func init() {
	fs := &cmdSearch.Flags
	fs.IntVar(&searchOpts.MaxResults, "max-results", 0, "Maximum results to request (10..100)")
	fs.StringVar(&searchOpts.PageToken, "page", "", "Page token to resume search")
	fs.StringVar(&searchOpts.SinceID, "after", "", "Return tweets (strictly) after this ID")
	fs.StringVar(&searchOpts.UntilID, "before", "", "Return tweets (strictly) before this ID")
	fs.StringVar(&searchQuery, "query", "", "Search query (required)")
	fs.Var(timestamp{&searchOpts.StartTime}, "since", "Return tweets no older than this")
	fs.Var(timestamp{&searchOpts.EndTime}, "until", "Return tweets no newer than this")
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
