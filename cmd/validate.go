package main

import (
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/urfave/cli"
)

func validateName(cliCtx *cli.Context, option string, err error) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return val, nil
	}

	name := manifold.Name(val)
	if err := name.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, err)
	}

	return val, nil
}

func validateLabel(cliCtx *cli.Context, option string, err error) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return val, nil
	}

	label := manifold.Label(val)
	if err := label.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, err)
	}

	return val, nil
}
