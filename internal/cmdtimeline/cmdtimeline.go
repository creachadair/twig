// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdtimeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/ostatus"
	"github.com/creachadair/twitter/types"
)

var Command = &command.C{
	Name: "timeline",
	Help: "Commands to browse timelines.",
	Commands: []*command.C{
		{
			Name:  "user",
			Usage: "username/id",
			Help:  "Fetch the user timeline for the given user.",
			Run: runWithID(func(id string) ostatus.TimelineQuery {
				return ostatus.UserTimeline(id, &opts)
			}),
		},
		{
			Name:  "home",
			Usage: "username/id",
			Help:  "Fetch the home timeline for the given user.",
			Run: runWithID(func(id string) ostatus.TimelineQuery {
				return ostatus.HomeTimeline(id, &opts)
			}),
		},
		{
			Name:  "mentions",
			Usage: "username/id",
			Help:  "Fetch the mentions timeline for the given user.",
			Run: runWithID(func(id string) ostatus.TimelineQuery {
				return ostatus.MentionsTimeline(id, &opts)
			}),
		},
	},
}

var opts ostatus.TimelineOpts

func init() {
	Command.Flags.BoolVar(&opts.ByID, "id", false, "Resolve user by ID")
	Command.Flags.IntVar(&opts.MaxResults, "max-results", 0, "Maximum results to fetch")
}

func runWithID(newQuery func(id string) ostatus.TimelineQuery) func(*command.Context, []string) error {
	return func(ctx *command.Context, args []string) error {
		rest, err := config.ParseParams(args, &opts.Optional)
		if err != nil {
			return err
		} else if len(rest) != 1 || rest[0] == "" {
			return command.FailWithUsage(ctx, rest)
		}

		cli, err := ctx.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := newQuery(args[0]).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			D []*types.Tweet `json:"data"`
		}{D: rsp.Tweets})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
}
