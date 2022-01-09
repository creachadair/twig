// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmduser

import (
	"context"
	"fmt"
	"strings"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name:  "user",
	Usage: "[options] user/id/field ...",
	Help: `
Look up the specified user IDs or usernames.

Each argument is either a username (@name), user ID (12345), or an optional
field specifier.  A field specifier has the form type:field, e.g., user:entities
As a special case, :field is shorthand for "user:field".
`,

	Run: func(env *command.Env, args []string) error {
		parsed := config.ParseArgs(args, "user")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(env, "Error: no usernames or IDs were specified")
			return command.FailWithUsage(env, args)
		}
		var ids, names []string
		for _, key := range parsed.Keys {
			if strings.HasPrefix(key, "@") {
				names = append(names, strings.TrimPrefix(key, "@"))
			} else {
				ids = append(ids, key)
			}
		}

		ctx := context.Background()
		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		if len(ids) != 0 {
			rsp, err := users.Lookup(ids[0], &users.LookupOpts{
				More:     ids[1:],
				Optional: parsed.Fields,
			}).Invoke(ctx, cli)
			if err != nil {
				return err
			} else if err := config.PrintJSON(rsp.Users); err != nil {
				return err
			}
		}
		if len(names) != 0 {
			rsp, err := users.LookupByName(names[0], &users.LookupOpts{
				More:     names[1:],
				Optional: parsed.Fields,
			}).Invoke(ctx, cli)
			if err != nil {
				return err
			} else if err := config.PrintJSON(rsp.Users); err != nil {
				return err
			}
		}
		return nil
	},
}
