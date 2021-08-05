// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
)

var Command = &command.C{
	Name: "stream",
	Help: "Commands to query the stream APIs.",
	Commands: []*command.C{
		{
			Name:  "search",
			Usage: "field-spec...",
			Help:  "Search for matches to the current query rules",
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
				opts.Optional = parsed.Fields
				return tweets.SearchStream(func(rsp *tweets.Reply) error {
					data, err := json.Marshal(rsp.Reply)
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}, &opts).Invoke(context.Background(), cli)
			},
		},
	},
}

var opts tweets.StreamOpts

func init() {
	Command.Flags.IntVar(&opts.MaxResults, "max-results", 0, "Maximum results to fetch")
}
