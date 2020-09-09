package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twig/internal/cmdhelp"
	"github.com/creachadair/twig/internal/cmdtweets"
	"github.com/creachadair/twig/internal/cmdusers"
)

var (
	configFile = "$HOME/.config/twig/config.yml"
	logLevel   = 0

	root = &command.C{
		Name:  filepath.Base(os.Args[0]),
		Usage: `<command> [arguments]`,
		Help:  `A command-line client for the Twitter API.`,

		Init: func(ctx *command.Context) error {
			if logLevel > 0 {
				cfg := ctx.Config.(*config.Config)
				cfg.Log = func(tag jhttp.LogTag, msg string) {
					log.Printf("DEBUG :: %s | %s", tag, msg)
				}
				cfg.LogMask = jhttp.LogTag(logLevel)
			}
			return nil
		},

		Commands: append([]*command.C{
			cmdusers.Command,
			cmdtweets.Command,
			cmdhelp.Command,
		}, cmdhelp.Topics...),
	}
)

func init() {
	root.Flags.StringVar(&configFile, "config", configFile, "Configuration file path")
	root.Flags.IntVar(&logLevel, "log-level", 0, "Verbose client logging level (log tag mask)")
}

func main() {
	path := os.ExpandEnv(configFile)
	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("Loading config file: %v", err)
	}

	if err := command.Execute(root.NewContext(cfg), os.Args[1:]); err != nil {
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
