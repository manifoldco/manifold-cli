package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	initCmd := cli.Command{
		Name:  "init",
		Usage: "Initialize the current directory for a specified app",
		Flags: append(teamFlags, []cli.Flag{
			appFlag(),
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing app",
			},
		}...),
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.LoadTeamPrefs, initDir),
	}

	cmds = append(cmds, initCmd)
}

func initDir(cliCtx *cli.Context) error {
	ctx := context.Background()
	appName := cliCtx.String("app")

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	mYaml, err := config.LoadYaml(false)
	if err != nil {
		return err
	}

	if mYaml != nil && mYaml.Path != "" && !cliCtx.Bool("force") {
		return cli.NewExitError(fmt.Sprintf("This directory is already linked to application `%s`.", mYaml.App), -1)
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}
	if !s.Authenticated() {
		return errs.ErrNotLoggedIn
	}

	mClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Marketplace API client: "+
			err.Error(), -1)
	}

	res, err := clients.FetchResources(ctx, mClient, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch resource list: "+err.Error(), -1)
	}

	appNames := fetchUniqueAppNames(res)
	if len(appNames) == 0 {
		return errs.ErrNoApps
	}

	newA, appName, err := prompts.SelectCreateAppName(appNames, appName, true)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select app.")
	}
	if newA == -1 {
		// TODO: create app name that doesn't exist yet
		// https://github.com/manifoldco/engineering/issues/2614
		return cli.NewExitError("Whoops! A new app cannot be created without a resource", -1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	mYaml.App = appName
	mYaml.Path = filepath.Join(cwd, config.YamlFilename)

	err = mYaml.Save()
	if err != nil {
		return err
	}

	// Display the output
	fmt.Println("\nThis directory and its subdirectories have been linked to:")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "App:\t%s\n", appName)
	w.Flush()

	return nil
}
