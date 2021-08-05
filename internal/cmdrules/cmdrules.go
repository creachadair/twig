// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdrules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/rules"
)

var Command = &command.C{
	Name: "rules",
	Help: "View, add, and delete streaming search rules.",
	Commands: []*command.C{
		cmdGet,
		cmdAdd,
		cmdDelete,
	},
}

var cmdGet = &command.C{
	Name:  "get",
	Usage: "id ...",
	Help: `
Look up search rules by ID.

Each argument must be a rule ID. If no IDs are given, all known 
rules are listed.
`,

	Run: func(env *command.Env, args []string) error {
		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		rsp, err := rules.Get(args...).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			R []rules.Rule `json:"rules"`
		}{R: rsp.Rules})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var cmdDelete = &command.C{
	Name:  "delete",
	Usage: "id ...",
	Help:  `Delete search rules by ID.`,

	Run: func(env *command.Env, args []string) error {
		if len(args) == 0 {
			fmt.Fprintln(env, "Error: no rule IDs were specified")
			return command.FailWithUsage(env, args)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		rsp, err := rules.Update(rules.Deletes(args)).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			M *rules.Meta `json:"meta"`
		}{M: rsp.Meta})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var cmdAdd = &command.C{
	Name:  "add",
	Usage: "(tag=query|query)...",
	Help: `Add search rules.

A rule must at minimum provide a search query.
If a tag= prefix is given, the rule is labelled with that tag.
`,

	Run: func(env *command.Env, args []string) error {
		if len(args) == 0 {
			fmt.Fprintln(env, "Error: no rules were specified")
			return command.FailWithUsage(env, args)
		}

		var adds rules.Adds
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			rule := rules.Add{Query: parts[len(parts)-1]}
			if len(parts) == 2 {
				rule.Tag = parts[0]
			}
			adds = append(adds, rule)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		rsp, err := rules.Update(adds).Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		data, err := json.Marshal(struct {
			R []rules.Rule `json:"rules,omitempty"`
			M *rules.Meta  `json:"meta"`
		}{R: rsp.Rules, M: rsp.Meta})
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}
