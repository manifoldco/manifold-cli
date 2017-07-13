package placeholder

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// StringFlag represents a flag which satisfies the github.com/urfave/cli.Flag
// interface while also supporting the concept of a default value and required
// flag.
type StringFlag struct {
	cli.StringFlag
	Placeholder string
	Required    bool
}

// String returns a string for explaining this flag, used by
// github.com/urfave/cli to print out help text.
func (sf StringFlag) String() string {
	flags := prefixedNames(sf.Name, sf.Placeholder)
	def := ""

	if sf.Value != "" {
		def = fmt.Sprintf(" (default: %s)", sf.Value)
	}

	return fmt.Sprintf("%s\t%s%s", flags, sf.Usage, def)
}

// New returns a new StringFlag using the provided information
func New(name, placeholder, usage, value, envvar string, required bool) StringFlag {
	return StringFlag{
		StringFlag: cli.StringFlag{
			Name:   name,
			Usage:  usage,
			Value:  value,
			EnvVar: envvar,
		},
		Placeholder: placeholder,
		Required:    required,
	}
}

// prefixedNames and prefixFor are taken from urfave/cli
func prefixedNames(fullName, placeholder string) string {
	var prefixed string
	parts := strings.Split(fullName, ",")
	for i, name := range parts {
		name = strings.Trim(name, " ")
		prefixed += prefixFor(name) + name
		if placeholder != "" {
			prefixed += " " + placeholder
		}
		if i < len(parts)-1 {
			prefixed += ", "
		}
	}
	return prefixed
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}
