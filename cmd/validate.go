package main

import (
	"fmt"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/errs"
)

func validateName(cliCtx *cli.Context, option string) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return val, nil
	}

	name := manifold.Name(val)
	if err := name.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s name!", option), -1,
		))
	}

	return val, nil
}

func validateLabel(cliCtx *cli.Context, option string) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return val, nil
	}

	label := manifold.Label(val)
	if err := label.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s!", option), -1,
		))
	}

	return val, nil
}

func requiredLabel(cliCtx *cli.Context, option string) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("--%s is required", option), -1,
		))
	}

	return validateLabel(cliCtx, option)
}
