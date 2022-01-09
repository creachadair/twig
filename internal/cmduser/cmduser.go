// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmduser

import (
	"context"
	"fmt"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name:  "user",
	Usage: "[options] user/id/field ...",
	Help: `
Look up the specified user IDs or usernames.

Each argument is either a username, user ID, or field specifier.
A field specifier has the form type:field, e.g., "user:entities".
As a special case, :field is shorthand for "user:field".
`,

	Run: func(env *command.Env, args []string) error {
		parsed := config.ParseArgs(args, "user")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(env, "Error: no usernames or IDs were specified")
			return command.FailWithUsage(env, args)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		opts := &users.LookupOpts{
			More:     parsed.Keys[1:],
			Optional: parsed.Fields,
		}

		var q users.Query
		if byID {
			q = users.Lookup(parsed.Keys[0], opts)
		} else {
			q = users.LookupByName(parsed.Keys[0], opts)
		}
		rsp, err := q.Invoke(context.Background(), cli)
		if err != nil {
			return err
		}
		return config.PrintJSON(rsp.Users)
	},
}

var byID bool

func init() {
	Command.Flags.BoolVar(&byID, "id", false, "Resolve users by ID")
}
