package color

import (
	"bytes"

	"github.com/juju/ansiterm"
)

// Bold returns a string in bold.
func Bold(i interface{}) string {
	ctx := ansiterm.Context{
		Styles: []ansiterm.Style{ansiterm.Bold},
	}
	return printContext(ctx, i)
}

// Color returns a string in the specified color.
func Color(c ansiterm.Color, i interface{}) string {
	ctx := ansiterm.Context{
		Foreground: c,
	}
	return printContext(ctx, i)
}

// Faint returns a string in gray.
func Faint(i interface{}) string {
	return Color(ansiterm.Gray, i)
}

func printContext(c ansiterm.Context, i interface{}) string {
	buf := &bytes.Buffer{}
	w := ansiterm.NewWriter(buf)
	w.SetColorCapable(true)
	c.Fprint(w, i)
	return buf.String()
}
