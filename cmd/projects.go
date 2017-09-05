package main

import (
	"context"
	"fmt"
	"strings"

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
				Name:  "create",
				Usage: "Create a new projects",
				Flags: append(teamFlags, []cli.Flag{
					descriptionFlag(),
				}...),
				ArgsUsage: "[name]",
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, createProjectCmd),
			},
			{
				Name:  "update",
				Usage: "Update an existing project",
				Flags: append(teamFlags, []cli.Flag{
					nameFlag(), descriptionFlag(),
				}...),
				ArgsUsage: "[label]",
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadTeamPrefs,
					updateProjectCmd),
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

	projectDescription := cliCtx.String("description")

	autoSelectName := projectName != ""
	projectName, err = prompts.ProjectName(projectName, autoSelectName)
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to name project")
	}

	autoSelectDescription := projectDescription != ""
	projectDescription, err = prompts.ProjectDescription(projectDescription, autoSelectDescription)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not add description to project")
	}

	params := projectClient.NewPostProjectsParamsWithContext(ctx)
	body := &mModels.CreateProjectBody{
		Name:        manifold.Name(projectName),
		Description: projectDescription,
		Label:       generateLabel(projectName),
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

func updateProjectCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	projectLabel, err := optionalArgLabel(cliCtx, 0, "label")
	if err != nil {
		return err
	}

	newProjectName, err := validateName(cliCtx, "name", "project")
	if err != nil {
		return err
	}

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	p, err := selectProject(ctx, projectLabel, teamID, marketplaceClient)
	if err != nil {
		return err
	}

	autoSelect := newProjectName != ""
	newProjectName, err = prompts.ProjectName(newProjectName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select project")
	}

	params := projectClient.NewPatchProjectsIDParamsWithContext(ctx)
	body := &mModels.PublicUpdateProjectBody{
		Name:  manifold.Name(newProjectName),
		Label: generateLabel(newProjectName),
	}

	params.SetID(p.ID.String())
	params.SetBody(&mModels.PublicUpdateProject{
		Body: body,
	})

	spin := prompts.NewSpinner("Updating project")
	spin.Start()
	defer spin.Stop()

	if err := updateProject(params); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not update project: %s", err), -1)
	}

	fmt.Printf("\nYour project \"%s\" has been updated\n", newProjectName)
	return nil
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

func updateProject(params *projectClient.PatchProjectsIDParams) error {
	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	_, err = marketplaceClient.Project.PatchProjectsID(params, nil)
	if err != nil {
		switch e := err.(type) {
		case *projectClient.PatchProjectsIDBadRequest:
			return e.Payload
		case *projectClient.PatchProjectsIDConflict:
			return e.Payload
		case *projectClient.PatchProjectsIDUnauthorized:
			return e.Payload
		case *projectClient.PatchProjectsIDForbidden:
			return e.Payload
		case *projectClient.PatchProjectsIDInternalServerError:
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

// selectProject prompts a user to select a project (if selects the one provided automatically)
func selectProject(ctx context.Context, projectLabel string, teamID *manifold.ID, marketplaceClient *client.Marketplace) (*mModels.Project, error) {
	projects, err := clients.FetchProjects(ctx, marketplaceClient, teamID, false)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	if len(projects) == 0 {
		return nil, errs.ErrNoProjects
	}

	idx, _, err := prompts.SelectProject(projects, projectLabel)
	if err != nil {
		return nil, prompts.HandleSelectError(err, "Could not select project")
	}

	p := projects[idx]
	return p, nil
}

// generateLabel makes a name lowercase and replace spaces with dashes
func generateLabel(name string) manifold.Label {
	label := strings.Replace(strings.ToLower(name), " ", "-", -1)
	return manifold.Label(label)
}
