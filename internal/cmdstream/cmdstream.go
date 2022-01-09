// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdstream

import (
	"context"
	"flag"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
)

var Command = &command.C{
	Name:  "stream",
	Usage: "field-spec...",
	Help: `Stream matches to the current query rules.

See the "rules" command for creating, viewing, and deleting the
streaming search rules.`,
	SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
		fs.IntVar(&opts.maxResults, "max", 0, "Maximum results to fetch (0 means all)")
	},
	Run: func(env *command.Env, args []string) error {
		parsed := config.ParseArgs(args, "tweet")
		if len(parsed.Keys) != 0 {
			fmt.Fprintf(env, "Error: extra arguments after query %v\n", parsed.Keys)
			return command.FailWithUsage(env, args)
		}
		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		return tweets.SearchStream(func(rsp *tweets.Reply) error {
			return config.PrintJSON(rsp.Tweets)
		}, &tweets.StreamOpts{
			MaxResults: opts.maxResults,
			Optional:   parsed.Fields,
		}).Invoke(context.Background(), cli)
	},
}

var opts struct {
	maxResults int
}
