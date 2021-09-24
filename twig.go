// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/creachadair/command"
	"github.com/creachadair/jhttp"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twig/internal/cmdhelp"
	"github.com/creachadair/twig/internal/cmdlist"
	"github.com/creachadair/twig/internal/cmdlookup"
	"github.com/creachadair/twig/internal/cmdrules"
	"github.com/creachadair/twig/internal/cmdsearch"
	"github.com/creachadair/twig/internal/cmdstream"
	"github.com/creachadair/twig/internal/cmdtimeline"
	"github.com/creachadair/twig/internal/cmdtweet"
	"github.com/creachadair/twig/internal/cmduser"
)

var (
	configFile = "$HOME/.config/twig/config.yml"
	logLevel   int
	authUser   string

	root = &command.C{
		Name:  filepath.Base(os.Args[0]),
		Usage: `<command> [arguments]`,
		Help:  `A command-line client for the Twitter API.`,

		Init: func(env *command.Env) error {
			path := os.ExpandEnv(configFile)
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if logLevel > 0 {
				cfg.Log = func(tag jhttp.LogTag, msg string) {
					log.Printf("DEBUG :: %s | %s", tag, msg)
				}
				cfg.LogMask = jhttp.LogTag(logLevel)
			}
			cfg.AuthUser = authUser
			env.Config = cfg
			return nil
		},

		Commands: append([]*command.C{
			cmdlookup.Command,
			cmdsearch.Command,
			cmduser.Command,
			cmdrules.Command,
			cmdstream.Command,
			cmdtweet.Command,
			cmdtimeline.Command,
			cmdlist.Command,
			command.HelpCommand(nil),
		}, cmdhelp.Topics...),
	}
)

func init() {
	root.Flags.StringVar(&configFile, "config", configFile, "Configuration file path")
	root.Flags.IntVar(&logLevel, "log-level", 0, "Verbose client logging level (log tag mask)")
	root.Flags.StringVar(&authUser, "auth-user", authUser, "Authenticate with user context")
}

func main() {
	if err := command.Execute(root.NewEnv(nil), os.Args[1:]); err != nil {
		if errors.Is(err, command.ErrUsage) {
			os.Exit(2)
		}
		log.Printf("Error: %v", err)
		var jerr *jhttp.Error
		if errors.As(err, &jerr) {
			fmt.Println(string(jerr.Data))
		}
		os.Exit(1)
	}
}
