// Package command defines plumbing for command dispatch.
package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Context is the environment passed to the Run function of a command.
// A Context implements the io.Writer interface, and should be used as
// the target of any diagnostic output the command wishes to emit.
// Primary command output should be sent to stdout.
type Context struct {
	Self    *C          // the C value that carries the Run function
	Parent  *C          // if this is a subcommand, its parent command (or nil)
	Name    string      // the name by which the command was invoked
	Helping bool        // whether we are handling a help request
	Config  interface{} // configuration data
	Log     io.Writer   // where to write diagnostic output (nil for os.Stderr)
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

	// Execute the action of the command. If nil, calls FailWithUsage.
	Run func(ctx *Context, args []string) error

	// If set, this will be called after flags are parsed (if any) but before
	// any subcommands are processed. If it reports an error, execution stops
	// and that error is returned to the caller.
	Init func(ctx *Context) error

	// Subcommands of this command.
	Commands []*C
}

func (c *C) findDispatchTarget(ctx *Context, name string) *C {
	for _, cmd := range c.Commands {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
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
		if err == flag.ErrHelp {
			return RunShortHelp(ctx, args)
		} else if err != nil {
			return err
		}
		args = cmd.Flags.Args()
	}

	if cmd.Init != nil {
		if err := cmd.Init(ctx); err != nil {
			return fmt.Errorf("initializing %q: %v", cmd.Name, err)
		}
	}

	// Unclaimed (non-flag) arguments may be free arguments for this command, or
	// may belong to a subcommand.
	for len(args) != 0 {
		// If there's a subcommand on this name, that takes precedence.
		if sub := cmd.findDispatchTarget(ctx, args[0]); sub != nil {
			nctx := *ctx
			nctx.Self = sub
			nctx.Parent = cmd
			return Execute(&nctx, args[1:])
		}

		// Otherwise...
		//
		// If we see the word "help", record that we're looking for help and try
		// to move further down the chain. We don't do this at the END of the
		// chain, however, unless the cmd has no runner: Otherwise we might eat
		// an argument for the command itself.
		//
		// But if cmd.Run == nil, then there is nothing to do, and we can safely
		// request help. If a command wants "help" to work from its own argument
		// list it can include commands.LongHelpCommand in its subcommands.

		if args[0] == "help" && (len(args) > 1 || cmd.Run == nil) {
			ctx.Helping = true
			args = args[1:] // discard "help"
			continue
		}

		// The remaining args are free arguments for cmd itself.
		if cmd.Run == nil && !ctx.Helping {
			// This command does not have any action, so the arguments are dead.
			fmt.Fprintf(ctx, "Error: %s command %q not understood\n", cmd.Name, args[0])
		}
		break
	}

	// If the help flag is set, don't actually run the command,
	if ctx.Helping {
		return RunLongHelp(ctx, args)
	} else if cmd.Run == nil {
		return FailWithUsage(ctx, args)
	} else if len(args) == 1 && args[0] == "help" {
		// This is probably not what the user intends, but who knows?
		fmt.Fprintf(ctx, "Warning: \"help\" will be used as an argument to %[1]q (write \"%[1]s -help\" for command help)\n", cmd.Name)
	}
	return cmd.Run(ctx, args)
}
