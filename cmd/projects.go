package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	projectClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:  "projects",
		Usage: "Manage your projects",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new projects",
				Flags:     teamFlags,
				ArgsUsage: "[name]",
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, createProjectCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func createProjectCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	projectName, err := optionalArgLabel(cliCtx, 0, "name")
	if err != nil {
		return err
	}

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	autoSelect := projectName != ""
	projectName, err = prompts.ProjectName(projectName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to name project")
	}

	name := manifold.Name(projectName)

	params := projectClient.NewPostProjectsParamsWithContext(ctx)
	body := &mModels.CreateProjectBody{
		Name:   name,
		Label:  manifold.Label(strings.Replace(strings.ToLower(projectName), " ", "-", -1)),
		UserID: userID,
	}

	if teamID == nil {
		body.UserID = userID
	} else {
		body.TeamID = teamID
	}

	params.SetBody(&mModels.CreateProject{
		Body: body,
	})

	spin := prompts.NewSpinner("Creating new project")
	spin.Start()
	defer spin.Stop()

	_, err = marketplaceClient.Project.PostProjects(params, nil)
	if err != nil {
		switch e := err.(type) {
		case *projectClient.PostProjectsBadRequest:
		case *projectClient.PostProjectsUnauthorized:
		case *projectClient.PostProjectsConflict:
			return e.Payload
		case *projectClient.PostProjectsInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	fmt.Printf("Your project '%s' has been created\n", projectName)
	return nil
}

// loadMarketplaceClient returns an identify client based on the configuration file.
func loadMarketplaceClient() (*client.Marketplace, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	identityClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	return identityClient, nil
}
