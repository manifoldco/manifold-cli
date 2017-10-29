package prompts

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/chzyer/readline"
)

// IsInteractive tells whether a terminal support interaction, such as prompts
// and colors.
var IsInteractive = true

var globalSpinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond)

func init() {
	IsInteractive = readline.IsTerminal(int(os.Stdout.Fd()))
}

// NewSpinner returns a customized spinner
func NewSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	if suffix != "" {
		s.Suffix = " " + suffix
	}
	return s
}

// SpinStart starts the spinning animation for the global spinner
func SpinStart(suffix string) {
	if !IsInteractive {
		return
	}

	if suffix == "" {
		globalSpinner.Suffix = ""
	} else {
		globalSpinner.Suffix = " " + suffix
	}
	globalSpinner.Start()
}

// SpinStop stopts the Spinner animation for the global spinner
func SpinStop() {
	if !IsInteractive {
		return
	}

	globalSpinner.Stop()
}
