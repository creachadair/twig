package command

import (
	"flag"
	"io/ioutil"
	"strings"
)

// usageLines parses and normalizes usage lines. The command name is stripped
// from the head of each line if it is present.
func (c *C) usageLines() []string {
	var lines []string
	prefix := c.Name + " "
	for _, line := range strings.Split(c.Usage, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} else if line == c.Name {
			lines = append(lines, "")
		} else {
			lines = append(lines, strings.TrimPrefix(line, prefix))
		}
	}
	return lines
}

func (c *C) run(ctx *Context, args []string) error {
	if c.Run == nil {
		return FailWithUsage(ctx, args)
	}
	return c.Run(ctx, args)
}

// indent returns text indented as specified; first is prepended to the first
// line, and prefix to all subsequent lines.
func indent(first, prefix, text string) string {
	return first + strings.ReplaceAll(text, "\n", "\n"+prefix)
}

// FailWithUsage is a run function that logs a usage message for the command
// and returns ErrUsage.
func FailWithUsage(ctx *Context, args []string) error {
	ctx.Self.HelpInfo(false).WriteUsage(ctx)
	return ErrUsage
}

// FlagSet creates a new empty flag set for the given command name.
// This is a shortcut for flag.NewFlagSet(name, flag.ContinueOnError).
func FlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)
	return fs
}
