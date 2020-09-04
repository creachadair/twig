package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twig/internal/cmdusers"
	"github.com/creachadair/twitter/jhttp"
)

var (
	configFile = "$HOME/.config/twig/config.yml"
	logLevel   = 0

	root = &command.C{
		Name:  filepath.Base(os.Args[0]),
		Usage: `<command> [arguments]`,
		Flags: command.FlagSet(os.Args[0]),
		Help:  `A command-line client for the Twitter API.`,

		Init: func(ctx *command.Context) error {
			if logLevel > 0 {
				ctx.Config.(*config.Config).Log = func(tag, msg string) {
					if wantTag(tag) {
						log.Printf("DEBUG :: %s | %s", tag, msg)
					}
				}
			}
			return nil
		},

		Commands: []*command.C{
			cmdusers.Command,
		},
	}
)

func init() {
	root.Flags.StringVar(&configFile, "config", configFile, "Configuration file path")
	root.Flags.IntVar(&logLevel, "log-level", 0, "Verbose client logging level (1=http|2=auth|4=body)")
}

func main() {
	path := os.ExpandEnv(configFile)
	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("Loading config file: %v", err)
	}

	if err := command.Execute(&command.Context{
		Self:   root,
		Name:   root.Name,
		Config: cfg,
	}, os.Args[1:]); errors.Is(err, command.ErrUsage) {
		os.Exit(2)
	} else if err != nil {
		log.Printf("Error: %v", err)
		var jerr *jhttp.Error
		if errors.As(err, &jerr) {
			fmt.Println(string(jerr.Data))
		}
		os.Exit(1)
	}
}

func wantTag(tag string) bool {
	switch tag {
	case "RequestURL", "HTTPStatus":
		return logLevel&1 != 0
	case "Authorization":
		return logLevel&2 != 0
	case "ResponseBody", "StreamBody":
		return logLevel&4 != 0
	}
	return logLevel != 0
}
