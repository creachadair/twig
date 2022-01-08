// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdlist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/olists"
	"github.com/creachadair/twitter/types"
)

var Command = &command.C{
	Name: "list",
	Help: "Commands to list followers and following.",
	Commands: []*command.C{
		{
			Name:  "followers",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the followers of the specified user.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Followers(id, &opts)
			}),
		},
		{
			Name:  "following",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the users following the specified user.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Following(id, &opts)
			}),
		},
	},
}

var opts olists.FollowOpts

func init() {
	Command.Flags.BoolVar(&opts.ByID, "id", false, "Resolve user by ID")
	Command.Flags.StringVar(&opts.PageToken, "page-token", "", "Page token")
	Command.Flags.IntVar(&opts.PerPage, "page-size", 200, "Number of results per page")
}

func runWithID(newQuery func(id string) olists.Query) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		rest, err := config.ParseParams(args, &opts.Optional)
		if err != nil {
			return err
		} else if len(rest) != 1 || rest[0] == "" {
			return command.FailWithUsage(env, rest)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := newQuery(rest[0]).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		type meta struct {
			T string `json:"next_token"`
		}
		out := struct {
			D []*types.User `json:"data"`
			M *meta         `json:"meta,omitempty"`
		}{D: rsp.Users}
		if rsp.NextToken != "" {
			out.M = &meta{T: rsp.NextToken}
		}
		data, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
}
