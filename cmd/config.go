package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/middleware"

	"github.com/manifoldco/manifold-cli/generated/marketplace/client/credential"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

var configKeyRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,1000}$`)

func init() {
	cmd := cli.Command{
		Name:  "config",
		Usage: "View and modify resource configuration",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				ArgsUsage: "<key=value...>",
				Usage:     "Set one or more config values on a custom resource.",
				Flags: []cli.Flag{
					resourceFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, configSetCmd),
			},
			{
				Name:      "unset",
				ArgsUsage: "<key...>",
				Usage:     "Unset one or more config values on a custom resource.",
				Flags: []cli.Flag{
					resourceFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, configUnsetCmd),
			},
		},
	}

	cmds = append(cmds, cmd)
}

func patchConfig(cliCtx *cli.Context, req map[string]*string) error {
	ctx := context.Background()

	for k := range req {
		if !configKeyRegexp.MatchString(k) {
			return cli.NewExitError(fmt.Sprintf("Bad config key `%s`", k), -1)
		}
	}

	label, err := requiredLabel(cliCtx, "resource")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	marketplace, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError("Could not create marketplace client: "+err.Error(), -1)
	}

	resources, err := clients.FetchResources(ctx, marketplace)
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}

	// XXX just get a single resource
	var resource *models.Resource
	for _, r := range resources {
		if string(r.Body.Label) == label {
			resource = r
			break
		}
	}

	if resource == nil {
		return cli.NewExitError("No resource found with that label", -1)
	}
	if *resource.Body.Source != "custom" {
		return cli.NewExitError("Config can only be set on custom resources", -1)
	}

	_, err = marketplace.Credential.PatchResourcesIDConfig(&credential.PatchResourcesIDConfigParams{
		ID:      resource.ID.String(),
		Body:    req,
		Context: ctx,
	}, nil)
	if err != nil {
		return cli.NewExitError("Error updating config: "+err.Error(), -1)
	}

	return nil

}

func configSetCmd(cliCtx *cli.Context) error {
	req := make(map[string]*string)
	for _, arg := range cliCtx.Args() {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return cli.NewExitError("Config must be of the form KEY=VALUE", -1)
		}
		req[parts[0]] = &parts[1]
	}
	return patchConfig(cliCtx, req)
}

func configUnsetCmd(cliCtx *cli.Context) error {
	req := make(map[string]*string)
	for _, arg := range cliCtx.Args() {
		req[arg] = nil
	}
	return patchConfig(cliCtx, req)
}
