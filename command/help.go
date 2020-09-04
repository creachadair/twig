package command

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// HelpInfo records synthesized help details for a command.
type HelpInfo struct {
	Name     string
	Synopsis string
	Usage    string
	Help     string
	Flags    string
	Commands []HelpInfo // populated only if requested
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

// WriteUsage writes a usage summary to w.
func (h HelpInfo) WriteUsage(w io.Writer) { fmt.Fprint(w, h.Usage, "\n\n") }

// WriteSynopsis writes a usage summary and command synopsis to w.
// If the command defines flags, the flag summary is also written.
func (h HelpInfo) WriteSynopsis(w io.Writer) {
	h.WriteUsage(w)
	if h.Synopsis == "" {
		fmt.Fprint(w, "(no description available)\n\n")
	} else {
		fmt.Fprint(w, h.Synopsis+"\n\n")
	}
	if h.Flags != "" {
		fmt.Fprint(w, h.Flags, "\n\n")
	}
}

// WriteLong writes a complete help description to w, including a usage
// summary, full help text, flag summary, and subcommands.
func (h HelpInfo) WriteLong(w io.Writer) {
	h.WriteUsage(w)
	if h.Help == "" {
		fmt.Fprint(w, "(no description available)\n\n")
	} else {
		fmt.Fprint(w, h.Help, "\n\n")
	}
	if h.Flags != "" {
		fmt.Fprint(w, h.Flags, "\n\n")
	}
	if len(h.Commands) != 0 {
		base := h.Name + " "
		fmt.Fprintln(w, "Subcommands:")
		tw := tabwriter.NewWriter(w, 4, 8, 1, ' ', 0)
		for _, cmd := range h.Commands {
			syn := cmd.Synopsis
			if syn == "" {
				syn = "(no description available)"
			}
			fmt.Fprint(tw, "  ", base+cmd.Name, "\t:\t", syn, "\n")
		}
		tw.Flush()
		fmt.Fprintln(w)
	}
}
