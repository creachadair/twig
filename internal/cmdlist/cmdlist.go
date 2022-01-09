// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package cmdlist

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/creachadair/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twitter/lists"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

var Command = &command.C{
	Name: "list",
	Help: "Commands to interact with user lists.",

	SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
		fs.IntVar(&opts.maxResults, "max", 0, "Maximum results to return (0 means all)")
	},

	Commands: []*command.C{
		{
			Name:  "lookup",
			Usage: "id [fields...]",
			Help:  "Look up information about the specified list id.",
			Run: runList(func(parsed config.ParsedArgs) lists.Query {
				return lists.Lookup(parsed.Keys[0], &lists.ListOpts{
					Optional: parsed.Fields,
				})
			}),
		},
		{
			Name:  "owned-by",
			Usage: "user-id [fields...]",
			Help:  "Fetch information about the lists owned by user-id.",
			Run: runList(func(parsed config.ParsedArgs) lists.Query {
				return lists.OwnedBy(parsed.Keys[0], &lists.ListOpts{
					MaxResults: maxQueryResults(),
					Optional:   parsed.Fields,
				})
			}),
		},
		{
			Name:  "followers-of",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the followers of the specified user.",
			Run: runUsers(func(parsed config.ParsedArgs) users.Query {
				return users.Followers(parsed.Keys[0], &users.ListOpts{
					MaxResults: maxQueryResults(),
					Optional:   parsed.Fields,
				})
			}),
		},
		{
			Name:  "followed-by",
			Usage: "username/id [user.fields...]",
			Help:  "Fetch the users followed by the specified user.",
			Run: runUsers(func(parsed config.ParsedArgs) users.Query {
				return users.Following(parsed.Keys[0], &users.ListOpts{
					MaxResults: maxQueryResults(),
					Optional:   parsed.Fields,
				})
			}),
		},
		{
			Name:  "create",
			Usage: "name [description...]",
			Help:  "Create a new list with the given name and description.",
			SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
				fs.BoolVar(&opts.private, "private", false, "Set list to private")
			},
			Run: func(env *command.Env, args []string) error {
				if len(args) == 0 {
					return command.FailWithUsage(env, args)
				}
				name := args[0]
				desc := strings.Join(args[1:], " ")

				cli, err := env.Config.(*config.Config).NewClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}

				rsp, err := lists.Create(name, desc, opts.private).Invoke(context.Background(), cli)
				if err != nil {
					return err
				}
				return config.PrintJSON(rsp.Lists[0])
			},
		},
		{
			Name:  "delete",
			Usage: "id",
			Help:  "Delete the list with the specified id.",
			Run: func(env *command.Env, args []string) error {
				if len(args) == 0 {
					return command.FailWithUsage(env, args)
				}
				cli, err := env.Config.(*config.Config).NewClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}

				ok, err := lists.Delete(args[0]).Invoke(context.Background(), cli)
				if err != nil {
					return err
				}
				fmt.Printf("deleted: %v\n", ok)
				return nil
			},
		},
		{
			Name:  "update",
			Usage: "id",
			Help:  "Update the list with the specified id.",
			SetFlags: func(_ *command.Env, fs *flag.FlagSet) {
				fs.BoolVar(&opts.private, "private", false, "Whether the list should be private")
				fs.String("name", "", "The new name of the list")
				fs.String("description", "", "The new description of the list")
			},
			Run: func(env *command.Env, args []string) error {
				if len(args) == 0 {
					return command.FailWithUsage(env, args)
				}
				var uopts lists.UpdateOpts
				if name, ok := isSet(env.Command.Flags, "name"); ok {
					uopts.SetName(name)
				}
				if desc, ok := isSet(env.Command.Flags, "description"); ok {
					uopts.SetDescription(desc)
				}
				if _, ok := isSet(env.Command.Flags, "private"); ok {
					uopts.SetPrivate(opts.private)
				}

				cli, err := env.Config.(*config.Config).NewClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}

				ok, err := lists.Update(args[0], uopts).Invoke(context.Background(), cli)
				if err != nil {
					return err
				}
				fmt.Printf("updated: %v\n", ok)
				return nil
			},
		},
		{
			Name:  "add-member",
			Usage: "list-id user-id...",
			Help:  "Add the given user ids to the specified list id.",
			Run: func(env *command.Env, args []string) error {
				if len(args) < 2 {
					return command.FailWithUsage(env, args)
				}

				ctx := context.Background()
				cli, err := env.Config.(*config.Config).NewClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}

				listID := args[0]
				users, err := config.ResolveID(ctx, cli, args[1:])
				if err != nil {
					return err
				}

				for _, userID := range users {
					ok, err := lists.AddMember(listID, userID).Invoke(ctx, cli)
					if err != nil {
						return fmt.Errorf("add user %q: %w", userID, err)
					}
					fmt.Printf("%s: %v\n", userID, ok)
				}
				return nil
			},
		},
		{
			Name:  "remove-member",
			Usage: "list-id user-id...",
			Help:  "Remove the given user ids from the specified list id.",
			Run: func(env *command.Env, args []string) error {
				if len(args) < 2 {
					return command.FailWithUsage(env, args)
				}

				ctx := context.Background()
				cli, err := env.Config.(*config.Config).NewClient()
				if err != nil {
					return fmt.Errorf("creating client: %w", err)
				}

				listID := args[0]
				users, err := config.ResolveID(ctx, cli, args[1:])
				if err != nil {
					return err
				}

				for _, userID := range users {
					ok, err := lists.RemoveMember(listID, userID).Invoke(ctx, cli)
					if err != nil {
						return fmt.Errorf("add user %q: %w", userID, err)
					}
					fmt.Printf("%s: %v\n", userID, !ok)
				}
				return nil
			},
		},
		{
			Name:  "members",
			Usage: "list-id [user.fields...]",
			Help:  "Fetch the members of the specified list.",
			Run: runUsers(func(parsed config.ParsedArgs) users.Query {
				return lists.Members(parsed.Keys[0], &lists.ListOpts{
					MaxResults: maxQueryResults(),
					Optional:   parsed.Fields,
				})
			}),
		},
		{
			Name:  "followers",
			Usage: "list-id [user.fields...]",
			Help:  "Fetch the followers of the specified list.",
			Run: runUsers(func(parsed config.ParsedArgs) users.Query {
				return lists.Followers(parsed.Keys[0], &lists.ListOpts{
					MaxResults: maxQueryResults(),
					Optional:   parsed.Fields,
				})
			}),
		},
	},
}

var opts struct {
	maxResults int
	fields     types.UserFields
	private    bool
}

func maxQueryResults() int {
	if opts.maxResults <= 0 || opts.maxResults > 100 {
		return 100
	} else if opts.maxResults < 10 {
		return 10
	}
	return opts.maxResults
}

func isSet(fs flag.FlagSet, name string) (s string, ok bool) {
	fs.Visit(func(f *flag.Flag) {
		if !ok && f.Name == name {
			s = f.Value.String()
			ok = true
		}
	})
	return
}

func runList(newQuery func(config.ParsedArgs) lists.Query) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		parsed := config.ParseArgs(args, "list")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(env, "Error: missing required id argument")
			return command.FailWithUsage(env, args)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		ctx := context.Background()
		if ids, err := config.ResolveID(ctx, cli, parsed.Keys); err != nil {
			return fmt.Errorf("resolving users: %w", err)
		} else {
			parsed.Keys = ids
		}

		q := newQuery(parsed)
		var numResults int
		for q.HasMorePages() {
			rsp, err := q.Invoke(ctx, cli)
			if err != nil {
				return err
			}
			lst := rsp.Lists
			numResults += len(lst)
			if opts.maxResults > 0 && numResults > opts.maxResults {
				lst = lst[:len(lst)-(numResults-opts.maxResults)]
			}
			if err := config.PrintJSON(lst); err != nil {
				return err
			}
			if opts.maxResults > 0 && numResults >= opts.maxResults {
				return nil // nothing more to do
			}
		}
		return nil
	}
}

func runUsers(newQuery func(config.ParsedArgs) users.Query) func(*command.Env, []string) error {
	return func(env *command.Env, args []string) error {
		parsed := config.ParseArgs(args, "user")
		if len(parsed.Keys) == 0 {
			fmt.Fprintln(env, "Error: missing required id argument")
			return command.FailWithUsage(env, args)
		}

		cli, err := env.Config.(*config.Config).NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		ctx := context.Background()
		if ids, err := config.ResolveID(ctx, cli, parsed.Keys); err != nil {
			return fmt.Errorf("resolving users: %w", err)
		} else {
			parsed.Keys = ids
		}

		q := newQuery(parsed)
		var numResults int
		for q.HasMorePages() {
			rsp, err := q.Invoke(ctx, cli)
			if err != nil {
				return err
			}
			users := rsp.Users
			numResults += len(users)
			if opts.maxResults > 0 && numResults > opts.maxResults {
				users = users[:len(users)-(numResults-opts.maxResults)]
			}
			if err := config.PrintJSON(users); err != nil {
				return err
			}
			if opts.maxResults > 0 && numResults >= opts.maxResults {
				return nil // nothing more to do
			}
		}
		return nil
	}
}
