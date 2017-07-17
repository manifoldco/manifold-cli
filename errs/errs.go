package errs

import (
	"fmt"

	"github.com/urfave/cli"
)

// ErrMustLogin represents an error when a user must login to continue
var ErrMustLogin = cli.NewExitError("You must login to perform that command.", -1)

// ErrAlreadyLoggedIn represents an error where a user is attempting to login,
// but has an existing session.
var ErrAlreadyLoggedIn = cli.NewExitError("You're already logged in!", -1)

// ErrAlreadyLoggedOut represents an error where a user is attempting to logut,
// but does not have an existing session.
var ErrAlreadyLoggedOut = cli.NewExitError("You're already logged out!", -1)

// ErrNotLoggedIn represents an error where a user must log in to continue
var ErrNotLoggedIn = cli.NewExitError("You are not logged in!", -1)

// ErrInvalidAppName represents an error where a user has provided an invalid
// app name
var ErrInvalidAppName = cli.NewExitError("You've provided an invalid app name!", -1)

// ErrInvalidResourceName represents an error where a user has provided an
// invalid resource name
var ErrInvalidResourceName = cli.NewExitError("You've provided an invalid resource name!", -1)

// ErrInvalidPlanLabel represents an error where a user has provided an invalid
// plan label
var ErrInvalidPlanLabel = cli.NewExitError("You've provided an invalid plan!", -1)

// ErrInvalidRegionLabel represents an error where a user has provided an
// invalid region label
var ErrInvalidRegionLabel = cli.NewExitError("You've provided an invalid region!", -1)

// ErrInvalidProductLabel represents an error where a user has provided an
// invalid product name
var ErrInvalidProductLabel = cli.NewExitError("You've provided an invalid product!", -1)

// ErrProductNotFound represents an error where the provided user's product
// label does not exist.
var ErrProductNotFound = cli.NewExitError("The provided product does not exist!", -1)

// ErrPlanNotFound represents an error where the provided user's plan label
// does not exist.
var ErrPlanNotFound = cli.NewExitError("The provided plan does not exist!", -1)

// ErrRegionNotFound represents an error where the provided user's region label
// does not exist
var ErrRegionNotFound = cli.NewExitError("The provided region does not exist!", -1)

// ErrTooManyArgs represents an error where a user has provided too many
// command line arguments
var ErrTooManyArgs = cli.NewExitError("You've provided too many command line arguments!", -1)

// ErrSomethingWentHorriblyWrong represents an error that is completely out of
// our control and unexpected (such as a 500 from the API)
var ErrSomethingWentHorriblyWrong = cli.NewExitError("Something went horribly wrong; please try again!", -1)

// NewUsageExitError returns a new error that includes the usage string for the
// givne command along with the message from the given error.
func NewUsageExitError(ctx *cli.Context, err error) error {
	usage := usageString(ctx)
	return cli.NewExitError(fmt.Sprintf("%s\n%s", err.Error(), usage), -1)
}

// NewErrorExitError creates an ExitError with an appended error message
func NewErrorExitError(message string, err error) error {
	return cli.NewExitError(message+"\n"+err.Error(), -1)
}

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return fmt.Sprintf("Usage:\n%s%s %s [comand options] %s",
		spacer, ctx.App.HelpName, ctx.Command.Name, ctx.Command.ArgsUsage)
}
