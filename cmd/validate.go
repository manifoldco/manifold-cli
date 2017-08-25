package main

import (
	"fmt"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/errs"
)

func validateName(cliCtx *cli.Context, option string, typeName ...string) (string, error) {
	val := cliCtx.String(option)
	if val == "" {
		return val, nil
	}

	if len(typeName) > 1 {
		panic("Only one typeName can be provided")
	} else if len(typeName) == 0 {
		typeName = make([]string, 1)
		typeName[0] = option
	}

	name := manifold.Name(val)
	if err := name.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s name!", typeName[0]), -1,
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

func optionalArgLabel(cliCtx *cli.Context, idx int, name string) (string, error) {
	args := cliCtx.Args()

	if len(args) < idx+1 {
		return "", nil
	}

	val := args[idx]
	l := manifold.Label(val)
	if err := l.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s!", name), -1,
		))
	}

	return val, nil
}

func optionalArgName(cliCtx *cli.Context, idx int, name string) (string, error) {
	args := cliCtx.Args()

	if len(args) < idx+1 {
		return "", nil
	}

	val := args[idx]
	n := manifold.Name(val)
	if err := n.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s name", name), -1,
		))
	}

	return val, nil
}

func optionalArgEmail(cliCtx *cli.Context, idx int, name string) (string, error) {
	args := cliCtx.Args()

	if len(args) < idx+1 {
		return "", nil
	}

	val := args[idx]
	n := manifold.Email(val)
	if err := n.Validate(nil); err != nil {
		return "", errs.NewUsageExitError(cliCtx, cli.NewExitError(
			fmt.Sprintf("You've provided an invalid %s email", name), -1,
		))
	}

	return val, nil
}

func maxOptionalArgsLength(cliCtx *cli.Context, size int) error {
	args := cliCtx.Args()

	if len(args) > size {
		return errs.ErrTooManyArgs
	}

	return nil
}
