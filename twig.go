package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/creachadair/twig/command"
	"github.com/creachadair/twig/config"
	"github.com/creachadair/twig/internal/cmduser"
	"github.com/creachadair/twitter/jhttp"
)

var (
	configFile = "$HOME/.config/twig/config.yml"

	root = &command.C{
		Name:  filepath.Base(os.Args[0]),
		Usage: `<command> [arguments]`,
		Flags: command.FlagSet(os.Args[0]),

		Commands: []*command.C{
			cmduser.Command,
		},
	}
)

func init() {
	root.Flags.StringVar(&configFile, "config", configFile, "Configuration file path")
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
