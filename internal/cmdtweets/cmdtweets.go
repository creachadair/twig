package cmdtweets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/tweets"
)

var Command = &command.C{
	Name: "tweets",
	Help: "Look up or search tweets.",
	Commands: []*command.C{
		{
			Name:  "lookup",
			Usage: "id/field ...",
			Help: `
Look up the specified tweets by ID.

Each argument is either a tweet ID or field specifier.
A field specifier has the form type:field, e.g., "tweet:entities".
As a special case, :field is shorthand for "tweet:field".
`,

			Run: func(ctx *command.Context, args []string) error {
				parsed := config.ParseArgs(args, "tweet")
				if len(parsed.Keys) == 0 {
					fmt.Fprintln(ctx, "Error: no tweet IDs were specified")
					return command.FailWithUsage(ctx, args)
				}

				cli, err := ctx.Config.(*config.Config).NewBearerClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}
				rsp, err := tweets.Lookup(parsed.Keys[0], &tweets.LookupOpts{
					More:     parsed.Keys[1:],
					Optional: parsed.Fields,
				}).Invoke(context.Background(), cli)
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
		},
	},
}
