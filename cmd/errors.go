package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var errMustLogin = cli.NewExitError("You must login to perform that command.", -1)
var errAlreadyLoggedIn = cli.NewExitError("You're alredy logged in!", -1)
var errAlreadyLoggedOut = cli.NewExitError("You're already logged out!", -1)
var errNotLoggedIn = cli.NewExitError("You are not logged in!", -1)
var errTooManyArgs = cli.NewExitError("You've provided too many arguments!", -1)
var errInvalidAppName = cli.NewExitError("You've provided an invalid app name!", -1)

func newUsageExitError(ctx *cli.Context, err error) error {
	usage := usageString(ctx)
	return cli.NewExitError(fmt.Sprintf("%s\n%s", err.Error(), usage), -1)
}

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return fmt.Sprintf("Usage:\n%s%s %s [comand options] %s",
		spacer, ctx.App.HelpName, ctx.Command.Name, ctx.Command.ArgsUsage)
}
