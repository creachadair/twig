// Package command defines plumbing for command dispatch.
package command

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Context is the environment passed to the Run function of a command.
// A Context implements the io.Writer interface, and should be used as
// the target of any diagnostic output the command wishes to emit.
// Primary command output should be sent to stdout.
type Context struct {
	Self   *C          // the C value that carries the Run function
	Parent *C          // if this is a subcommand, its parent command (or nil)
	Name   string      // the name by which the command was invoked
	Config interface{} // configuration data
	Log    io.Writer   // where to write diagnostic output (nil for os.Stderr)
}

// output returns the log writer for c.
func (c *Context) output() io.Writer {
	if c.Log != nil {
		return c.Log
	}
	return os.Stderr
}

// Write implements the io.Writer interface. Writing to a context writes to its
// designated output stream, allowing the context to be sent diagnostic output.
func (c *Context) Write(data []byte) (int, error) {
	return c.output().Write(data)
}

// C carries the description and invocation function for a command.
type C struct {
	// The name of the command, preferably one word.
	Name string

	// A terse usage summary for the command. Multiple lines are allowed, but
	// each line should be self-contained for a particular usage sense.
	Usage string

	// A detailed description of the command. Multiple lines are allowed.
	// The first non-blank line of this text is used as a synopsis.
	Help string

	// If set, the flags parsed from the arguments.
	// If nil, Run is responsible for parsing its own flags.
	Flags *flag.FlagSet

	// Execute the action of the command.
	Run func(ctx *Context, args []string) error

	// Subcommands of this command.
	Commands []*C
}

// Dispatch searches for the given subcommand name under c. If it is found, it
// is passed to Execute with args. Otherwise, if name == "help", write a long
// help to ctx and report ErrUsage.
func (c *C) Dispatch(ctx *Context, name string, args []string) error {
	for _, cmd := range c.Commands {
		if cmd.Name == name {
			return Execute(&Context{
				Name:   cmd.Name,
				Self:   cmd,
				Parent: c,
				Config: ctx.Config,
				Log:    ctx.Log,
			}, args)
		}
	}

	// If there wasn't an explicitly-defined "help" command, simulate one.
	if name == "help" {
		c.HelpInfo(true).WriteLong(ctx.output())
		return ErrUsage
	}
	return fmt.Errorf("%s: subcommand %q not found", c.Name, name)
}

// HelpInfo returns help details for c. If includeCommands is true and c has
// subcommands, their help is also generated.
func (c *C) HelpInfo(includeCommands bool) HelpInfo {
	help := strings.TrimSpace(c.Help)
	prefix := "  " + c.Name + " "
	h := HelpInfo{
		Name:     c.Name,
		Synopsis: strings.SplitN(help, "\n", 2)[0],
		Usage:    "Usage:\n\n" + indent(prefix, prefix, strings.Join(c.usageLines(), "\n")),
		Help:     help,
	}
	if c.Flags != nil {
		var buf bytes.Buffer
		fmt.Fprintln(&buf, "\nOptions:")
		c.Flags.SetOutput(&buf)
		c.Flags.PrintDefaults()
		h.Flags = strings.TrimSpace(buf.String())
	}
	if includeCommands {
		for _, cmd := range c.Commands {
			h.Commands = append(h.Commands, cmd.HelpInfo(false)) // don't recur
		}
	}
	return h
}

// ErrUsage is returned from Execute if the user requested help.
var ErrUsage = errors.New("help requested")

// Execute runs the command given unprocessed arguments. If the command has
// flags they are parsed and errors are handled before invoking the handler.
//
// Execute writes usage information to ctx and returns ErrUsage if the
// command-line usage was incorrect or the user requested -help via flags.
func Execute(ctx *Context, rawArgs []string) error {
	cmd := ctx.Self
	args := rawArgs

	// If this command has a flag set, parse the arguments and check for errors
	// before passing control to the handler.
	if cmd.Flags != nil {
		err := cmd.Flags.Parse(rawArgs)
		if err != nil {
			if err == flag.ErrHelp {
				cmd.HelpInfo(true).WriteSynopsis(ctx.output())
				return ErrUsage
			}
			return err
		}
		args = cmd.Flags.Args()
	}

	// If there are unclaimed arguments and subcommands to consume them, try to
	// dispatch on the first in line. As a special case, "help" alone is sent
	// for dispatch.
	if len(args) != 0 {
		if len(cmd.Commands) != 0 || (len(args) == 1 && args[0] == "help") {
			return cmd.Dispatch(ctx, args[0], args[1:])
		}
	}
	return cmd.run(ctx, args)
}
