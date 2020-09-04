package cmdusers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name:  "users",
	Usage: "[options] name/id ...",
	Help:  `Look up the specified user IDs or usernames.`,
	Flags: command.FlagSet("user"),

	Run: func(ctx *command.Context, args []string) error {
		if len(args) == 0 {
			return errors.New("no usernames or IDs specified")
		}
		cli, err := ctx.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		opts := &users.LookupOpts{
			More: args[1:],
		}
		if userFields != "" {
			opts.Optional = append(opts.Optional, types.MiscFields{
				Label_:  "user.fields",
				Values_: strings.Split(userFields, ","),
			})
		}
		if expand != "" {
			opts.Optional = append(opts.Optional, types.MiscFields{
				Label_:  "expansions",
				Values_: strings.Split(expand, ","),
			})
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
	byID       bool
	userFields string
	expand     string
)

func init() {
	fs := Command.Flags
	fs.BoolVar(&byID, "id", false, "Resolve users by ID")
	fs.StringVar(&userFields, "fields", "", "Optional user fields (comma-separated)")
	fs.StringVar(&expand, "expand", "", "Optional expansions (comma-separated)")
}
