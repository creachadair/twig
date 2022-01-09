// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdstream

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
)

var Command = &command.C{
	Name:  "stream",
	Usage: "field-spec...",
	Help:  "Stream matches to the current query rules.",
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
			for _, tw := range rsp.Tweets {
				data, err := json.Marshal(tw)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			}
			return nil
		}, &tweets.StreamOpts{
			MaxResults: opts.maxResults,
			Optional:   parsed.Fields,
		}).Invoke(context.Background(), cli)
	},
}

var opts struct {
	maxResults int
}
