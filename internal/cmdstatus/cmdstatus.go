// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdstatus

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
	Name: "status",
	Help: "Commands to create and manipulate tweets",
	Commands: []*command.C{
		cmdUpdate,
		cmdDelete,
	},
}

var (
	inReplyTo    string
	autoPopReply bool
)

func init() {
	cmdUpdate.Flags.StringVar(&inReplyTo, "reply-to", "",
		"Reply to this tweet ID")
	cmdUpdate.Flags.BoolVar(&autoPopReply, "auto-reply", false,
		"Automatically populate reply based on mentions")
}

var cmdUpdate = &command.C{
	Name:  "update",
	Usage: "text...",
	Help:  `Create a new tweet from the given text.`,

	Run: func(ctx *command.Context, args []string) error {
		text := strings.TrimSpace(strings.Join(args, " "))
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
		}).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			D *types.Tweet `json:"data"`
		}{D: rsp.Tweet})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var cmdDelete = &command.C{
	Name:  "delete",
	Usage: "id",
	Help:  "Delete the tweet with the specified ID.",

	Run: func(ctx *command.Context, args []string) error {
		if len(args) != 1 {
			return command.FailWithUsage(ctx, args)
		}
		if args[0] == "" {
			return errors.New("empty ID string")
		}

		cli, err := ctx.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := ostatus.Delete(args[0], nil).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			D *types.Tweet `json:"data"`
		}{D: rsp.Tweet})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}
