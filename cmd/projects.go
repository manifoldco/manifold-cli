package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
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
			{
				Name:  "list",
				Usage: "List all your projects",
				Flags: teamFlags,
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, listProjectsCmd),
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

	projectName, err := optionalArgName(cliCtx, 0, "name")
	if err != nil {
		return err
	}

	autoSelect := projectName != ""
	projectName, err = prompts.ProjectName(projectName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to name project")
	}

	params := projectClient.NewPostProjectsParamsWithContext(ctx)
	body := &mModels.CreateProjectBody{
		Name:  manifold.Name(projectName),
		Label: generateLabel(projectName),
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

	if err := createProject(params); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not create project: %s", err), -1)
	}

	fmt.Printf("Your project '%s' has been created\n", projectName)
	return nil
}

func listProjectsCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	projects, err := clients.FetchProjects(ctx, teamID, marketplaceClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)

	bold := color.New(color.Bold).SprintFunc()

	fmt.Fprintf(w, "%s\n\n", bold("Project"))

	for _, project := range projects {
		fmt.Fprintf(w, "%s\n", project.Body.Label)
	}
	return w.Flush()
}

func createProject(params *projectClient.PostProjectsParams) error {
	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	_, err = marketplaceClient.Project.PostProjects(params, nil)
	if err != nil {
		switch e := err.(type) {
		case *projectClient.PostProjectsBadRequest:
			return e.Payload
		case *projectClient.PostProjectsUnauthorized:
			return e.Payload
		case *projectClient.PostProjectsConflict:
			return e.Payload
		case *projectClient.PostProjectsInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

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

// generateLabel makes a name lowercase and replace spaces with dashes
func generateLabel(name string) manifold.Label {
	label := strings.Replace(strings.ToLower(name), " ", "-", -1)
	return manifold.Label(label)
}
