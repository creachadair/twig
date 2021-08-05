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
			Usage: "[username/id] [tweet.fields...]",
			Help:  "Fetch the user timeline for the given user.",
			Run: runWithID(func(id string) ostatus.TimelineQuery {
				return ostatus.UserTimeline(id, &opts)
			}),
		},
		{
			Name:  "home",
			Usage: "[username/id] [tweet.fields...]",
			Help:  "Fetch the home timeline for the given user.",
			Run: runWithID(func(id string) ostatus.TimelineQuery {
				return ostatus.HomeTimeline(id, &opts)
			}),
		},
		{
			Name:  "mentions",
			Usage: "[username/id] [tweet.fields...]",
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
	Command.Flags.BoolVar(&opts.IncludeRetweets, "include-retweets", false, "Include retweets")
	Command.Flags.BoolVar(&opts.ExcludeReplies, "exclude-replies", false, "Exclude replies")
}

func runWithID(newQuery func(id string) ostatus.TimelineQuery) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		cfg := env.Config.(*config.Config)
		user := cfg.AuthUser

		rest, err := config.ParseParams(args, &opts.Optional)
		if err != nil {
			return err
		} else if len(rest) > 1 {
			return command.FailWithUsage(env, rest)
		} else if len(rest) == 1 {
			user = rest[0]
		}
		if user == "" {
			return command.FailWithUsage(env, rest)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := newQuery(user).Invoke(context.Background(), cli)
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
