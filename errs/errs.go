package errs

import (
	"encoding/json"
	"errors"
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

// ErrInvalidResourceName represents an error where a user has provided an
// invalid resource name
var ErrInvalidResourceName = cli.NewExitError("You've provided an invalid resource name!", -1)

// ErrProductNotFound represents an error where the provided user's product
// label does not exist.
var ErrProductNotFound = cli.NewExitError("The provided product does not exist!", -1)

// ErrPlanNotFound represents an error where the provided user's plan label
// does not exist.
var ErrPlanNotFound = cli.NewExitError("The provided plan does not exist!", -1)

// ErrResourceNotFound represents an error where the provided user's resource label
// does not exist
var ErrResourceNotFound = cli.NewExitError("The provided resource does not exist!", -1)

// ErrTeamNotFound represents an error where the provided user's team label does not exist
var ErrTeamNotFound = cli.NewExitError("The provided team does not exist!", -1)

// ErrProjectNotFound represents an error where the provided project does not exist
var ErrProjectNotFound = cli.NewExitError("The provided project does not exist!", -1)

// ErrRegionNotFound represents an error where the provided user's region label
// does not exist
var ErrRegionNotFound = cli.NewExitError("The provided region does not exist!", -1)

// ErrNoApps represents an error where the action requires a resource app but none exist
var ErrNoApps = cli.NewExitError("There are no resources with apps", -1)

// ErrNoTeams represents an error where at least one team is required for an action, but none are
// available
var ErrNoTeams = cli.NewExitError("No teams found", -1)

// ErrNoProjects represents an error when there are no projects found
var ErrNoProjects = cli.NewExitError("No projects found", -1)

// ErrNoResources represents an error where no resources exist to preform some action on
var ErrNoResources = cli.NewExitError("No resources found", -1)

// ErrNoPlans represents an error where no plans exist for the resource being created or updated
var ErrNoPlans = cli.NewExitError("No plans found for this product", -1)

// ErrNoProviders represents an error where no providers exist
var ErrNoProviders = cli.NewExitError("No providers found", -1)

// ErrNoProducts represents an error where no products exist
var ErrNoProducts = cli.NewExitError("No products found", -1)

// ErrTooManyArgs represents an error where a user has provided too many
// command line arguments
var ErrTooManyArgs = cli.NewExitError("You've provided too many command line arguments!", -1)

// ErrSomethingWentHorriblyWrong represents an error that is completely out of
// our control and unexpected (such as a 500 from the API)
var ErrSomethingWentHorriblyWrong = cli.NewExitError("Something went horribly wrong; please try again!", -1)

// NewUsageExitError returns a new error that includes the usage string for the
// givne command along with the message from the given error.
func NewUsageExitError(ctx *cli.Context, err error) error {
	fmt.Printf("Invalid usage: %s\n\n", err.Error())
	cli.ShowCommandHelpAndExit(ctx, ctx.Command.Name, -1)
	return nil
}

// NewErrorExitError creates an ExitError with an appended error message
func NewErrorExitError(message string, err error) error {
	return cli.NewExitError(message+"\n"+err.Error(), -1)
}

type stripeError struct {
	Message string `json:"message"`
}

func (s stripeError) Error() error {
	return errors.New(s.Message)
}

// NewStripeError marshals the stripe json error to a human error
func NewStripeError(err error) error {
	var sErr stripeError
	jErr := json.Unmarshal([]byte(err.Error()), &sErr)
	if jErr != nil {
		return err
	}
	return sErr.Error()
}
