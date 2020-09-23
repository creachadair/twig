// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdtweet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/ostatus"
	"github.com/creachadair/twitter/types"
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
			Run: runWithID(func(id string) ostatus.Query {
				return ostatus.Delete(id, nil)
			}),
		},
		{
			Name:  "like",
			Usage: "id",
			Help:  "Like the tweet with the specified ID.",
			Run: runWithID(func(id string) ostatus.Query {
				return ostatus.Like(id, nil)
			}),
		},
		{
			Name:  "unlike",
			Usage: "id",
			Help:  "Un-like the tweet with the specified ID.",
			Run: runWithID(func(id string) ostatus.Query {
				return ostatus.Unlike(id, nil)
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

	Run: func(ctx *command.Context, args []string) error {
		var opts types.TweetFields
		rest, err := config.ParseParams(args, &opts)
		if err != nil {
			return err
		}
		text := strings.TrimSpace(strings.Join(rest, " "))
		if text == "" {
			return errors.New("empty status update")
		}

		cli, err := ctx.Config.(*config.Config).NewClient()
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
		data, err := json.Marshal(struct {
			D []*types.Tweet `json:"data"`
		}{D: rsp.Tweets})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func runWithID(newQuery func(id string) ostatus.Query) func(*command.Context, []string) error {
	return func(ctx *command.Context, args []string) error {
		if len(args) != 1 || args[0] == "" {
			return command.FailWithUsage(ctx, args)
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
