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
)

func init() {
	initCmd := cli.Command{
		Name:     "init",
		Usage:    "Initialize the current directory for a specified project",
		Category: "ADMINISTRATIVE",
		Flags: append(teamFlags, []cli.Flag{
			appFlag(),
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing project",
			},
		}...),
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, middleware.LoadTeamPrefs, initDir),
	}

	cmds = append(cmds, initCmd)
}

func initDir(cliCtx *cli.Context) error {
	ctx := context.Background()
	projectLabel := cliCtx.String("project")

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	mYaml, err := config.LoadYaml(false)
	if err != nil {
		return err
	}

	if mYaml != nil && mYaml.Path != "" && !cliCtx.Bool("force") {
		return cli.NewExitError(fmt.Sprintf("This directory is already linked to project `%s`.", mYaml.Project), -1)
	}

	mClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	ps, err := clients.FetchProjects(ctx, mClient, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to load projects: %s", err), -1)
	}
	if len(ps) == 0 {
		return errs.ErrNoProjects
	}

	pIdx, _, err := prompts.SelectProject(ps, projectLabel, true)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select project.")
	}
	if pIdx == -1 {
		projectLabel = ""
	} else {
		projectLabel = string(ps[pIdx].Body.Label)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	oldLabel := mYaml.Project
	mYaml.Project = projectLabel
	mYaml.Path = filepath.Join(cwd, config.YamlFilename)

	err = mYaml.Save()
	if err != nil {
		return err
	}

	if mYaml.Project == "" {
		fmt.Println("\nThis directory and its subdirectories have been unlinked from:")
		w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
		fmt.Fprintf(w, "Project:\t%s\n", oldLabel)
		w.Flush()
	} else {
		// Display the output
		fmt.Println("\nThis directory and its subdirectories have been linked to:")
		w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
		fmt.Fprintf(w, "Project:\t%s\n", projectLabel)
		w.Flush()
	}

	return nil
}
