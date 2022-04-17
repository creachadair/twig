// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdtweet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/edit"
	"github.com/creachadair/twitter/ostatus"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name: "tweet",
	Help: "Commands to create and manipulate tweets",
	Commands: []*command.C{
		cmdCreate,
		{
			Name:  "delete",
			Usage: "id",
			Help:  "Delete the tweet with the specified ID.",
			Run: runWithID(func(_, tweetID string) edit.Query {
				return edit.DeleteTweet(tweetID)
			}),
		},
		{
			Name:  "like",
			Usage: " id",
			Help:  "Mark tweet id as liked by the authorized user.",
			Run: runWithID(func(userID, tweetID string) edit.Query {
				return edit.Like(userID, tweetID)
			}),
		},
		{
			Name:  "unlike",
			Usage: "id",
			Help:  "Unmark tweet id as liked by the authorized user.",
			Run: runWithID(func(userID, tweetID string) edit.Query {
				return edit.Unlike(userID, tweetID)
			}),
		},
		{
			Name:  "retweet",
			Usage: "id",
			Help:  "Retweet tweet id from the authorized users.",
			Run: runWithID(func(userID, tweetID string) edit.Query {
				return edit.Retweet(userID, tweetID)
			}),
		},
		{
			Name:  "unretweet",
			Usage: "id",
			Help:  "Un-retweet tweet id from the authorized user.",
			Run: runWithID(func(userID, tweetID string) edit.Query {
				return edit.Unretweet(userID, tweetID)
			}),
		},
	},
}

var (
	inReplyTo    string
	autoPopReply bool
)

func init() {
	cmdCreate.Flags.StringVar(&inReplyTo, "reply-to", "",
		"Reply to this tweet ID")
	cmdCreate.Flags.BoolVar(&autoPopReply, "auto-reply", false,
		"Automatically populate reply based on mentions")
}

var cmdCreate = &command.C{
	Name:  "create",
	Usage: "text...",
	Help:  `Create a new tweet from the given text.`,

	Run: func(env *command.Env, args []string) error {
		var opts types.TweetFields
		rest, err := config.ParseParams(args, "tweet", &opts)
		if err != nil {
			return err
		}
		text := strings.TrimSpace(strings.Join(rest, " "))
		if text == "" {
			return errors.New("empty status update")
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := ostatus.Create(text, &ostatus.CreateOpts{
			InReplyTo:         inReplyTo,
			AutoPopulateReply: autoPopReply,
			Optional:          opts,
		}).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		return config.PrintJSON(rsp.Tweets)
	},
}

func runWithID(newQuery func(uid, tid string) edit.Query) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		if len(args) != 1 || args[0] == "" {
			return command.FailWithUsage(env, args)
		}
		cfg := env.Config.(*config.Config)
		cli, err := cfg.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		ctx := context.Background()
		var uid string
		if rsp, err := users.LookupByName(cfg.AuthUser, nil).Invoke(ctx, cli); err != nil {
			return fmt.Errorf("resolving user: %w", err)
		} else {
			uid = rsp.Users[0].ID
		}

		rsp, err := newQuery(uid, args[0]).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		return config.PrintJSON(rsp)
	}
}
