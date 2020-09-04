package cmdusers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/types"
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
		parsed := parseArgs(args)
		if len(parsed.keys) == 0 {
			fmt.Fprintln(ctx, "Error: no usernames or IDs were specified")
			return command.FailWithUsage(ctx, args)
		}

		cli, err := ctx.Config.(*config.Config).NewBearerClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		opts := &users.LookupOpts{
			More:     parsed.keys[1:],
			Optional: parsed.fields,
		}

		var q users.Query
		if byID {
			q = users.Lookup(args[0], opts)
		} else {
			q = users.LookupByName(args[0], opts)
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

var (
	byID   bool
	expand string
)

func init() {
	fs := Command.Flags
	fs.BoolVar(&byID, "id", false, "Resolve users by ID")
	fs.StringVar(&expand, "expand", "", "Optional expansions (comma-separated)")
}

type parsedArgs struct {
	keys   []string
	fields []types.Fields
}

func parseArgs(args []string) parsedArgs {
	var parsed parsedArgs
	if expand != "" {
		parsed.fields = append(parsed.fields, types.Expansions(strings.Split(expand, ",")))
	}
	fieldMap := make(map[string][]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 1 {
			parsed.keys = append(parsed.keys, arg)
			continue
		}
		if parts[0] == "" {
			parts[0] = "user"
		}
		fieldMap[parts[0]] = append(fieldMap[parts[0]], parts[1])
	}
	for key, vals := range fieldMap {
		parsed.fields = append(parsed.fields, types.MiscFields{
			Label_:  key + ".fields",
			Values_: vals,
		})
	}
	return parsed
}
