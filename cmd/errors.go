package main

import (
	"github.com/urfave/cli"
)

var errAlreadyLoggedIn = cli.NewExitError("You're alredy logged in!", -1)
var errAlreadyLoggedOut = cli.NewExitError("You're already logged out!", -1)
var errNotLoggedIn = cli.NewExitError("You are not logged in!", -1)
