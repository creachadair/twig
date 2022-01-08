// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdlist

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/olists"
	"github.com/creachadair/twitter/types"
)

var Command = &command.C{
	Name: "list",
	Help: "Commands to list followers and following.",

	SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
		fs.BoolVar(&opts.byID, "id", false, "Resolve user by ID")
		fs.StringVar(&opts.pageToken, "page-token", "", "Page token")
		fs.IntVar(&opts.pageSize, "page-size", 200, "Number of results per page")
	},

	Commands: []*command.C{
		{
			Name:  "followers",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the followers of the specified user.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Followers(id, newFollowOpts())
			}),
		},
		{
			Name:  "following",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the users following the specified user.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Following(id, newFollowOpts())
			}),
		},
		{
			Name:  "members",
			Usage: "list-id [user.fields...]",
			Help:  "Fetch the members of the specified list.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Members(id, newListOpts())
			}),
		},
		{
			Name:  "subscribers",
			Usage: "list-id [user.fields...]",
			Help:  "Fetch the subscribers to the specified list.",
			Run: runWithID(func(id string) olists.Query {
				return olists.Subscribers(id, newListOpts())
			}),
		},
	},
}

var opts struct {
	byID      bool
	pageToken string
	pageSize  int
	fields    types.UserFields
}

func newFollowOpts() *olists.FollowOpts {
	return &olists.FollowOpts{
		ByID:      opts.byID,
		PageToken: opts.pageToken,
		PerPage:   opts.pageSize,
		Optional:  opts.fields,
	}
}

func newListOpts() *olists.ListOpts {
	return &olists.ListOpts{
		PageToken: opts.pageToken,
		PerPage:   opts.pageSize,
		Optional:  opts.fields,
	}
}

func runWithID(newQuery func(id string) olists.Query) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		rest, err := config.ParseParams(args, &opts.fields)
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
