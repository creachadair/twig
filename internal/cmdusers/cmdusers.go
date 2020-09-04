package cmdusers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name:  "users",
	Usage: "[options] user/id/field ...",
	Help: `
Look up the specified user IDs or usernames.

Each argument is either a username, user ID, or field specifier.
A field specifier has the form type:field, e.g., "user:entities".
As a special case, :field is shorthand for "user:field".
`,
	Flags: command.FlagSet("user"),

	Run: func(ctx *command.Context, args []string) error {
		parsed := config.ParseArgs(args, "user")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(ctx, "Error: no usernames or IDs were specified")
			return command.FailWithUsage(ctx, args)
		}

		cli, err := ctx.Config.(*config.Config).NewBearerClient()
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
		data, err := json.Marshal(rsp.Reply)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var byID bool

func init() {
	Command.Flags.BoolVar(&byID, "id", false, "Resolve users by ID")
}
