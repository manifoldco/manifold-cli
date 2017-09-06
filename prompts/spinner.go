package prompts

import (
	"time"

	"github.com/briandowns/spinner"
)

var globalSpinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond)

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
	if suffix == "" {
		globalSpinner.Suffix = ""
	} else {
		globalSpinner.Suffix = " " + suffix
	}
	globalSpinner.Start()
}

// SpinStop stopts the Spinner animation for the global spinner
func SpinStop() {
	globalSpinner.Stop()
}
