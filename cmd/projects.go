package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	projectClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:     "projects",
		Usage:    "Manage your projects",
		Category: "RESOURCES",
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
				Name:  "update",
				Usage: "Update an existing project",
				Flags: append(teamFlags, []cli.Flag{
					nameFlag(), descriptionFlag(),
				}...),
				ArgsUsage: "[label]",
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadTeamPrefs,
					updateProjectCmd),
			},
			{
				Name:  "add",
				Usage: "Adds or moves a resource to a project",
				Flags: append(teamFlags, []cli.Flag{
					skipFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, addProjectCmd),
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

func updateProjectCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	projectLabel, err := optionalArgLabel(cliCtx, 0, "project")
	if err != nil {
		return err
	}

	newProjectName, err := validateName(cliCtx, "name", "project")
	if err != nil {
		return err
	}

	projectDescription := cliCtx.String("description")

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	p, err := selectProject(ctx, projectLabel, teamID, marketplaceClient)
	if err != nil {
		return err
	}

	autoSelectName := newProjectName != ""
	newProjectName, err = prompts.ProjectName(newProjectName, autoSelectName)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select project")
	}

	autoSelectDescription := projectDescription != ""
	projectDescription, err = prompts.ProjectDescription(projectDescription, autoSelectDescription)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not add description to project")
	}

	params := projectClient.NewPatchProjectsIDParamsWithContext(ctx)
	body := &mModels.PublicUpdateProjectBody{
		Name:  manifold.Name(newProjectName),
		Label: generateLabel(newProjectName),
	}

	if projectDescription != "" {
		body.Description = &projectDescription
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

func addProjectCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 2); err != nil {
		return err
	}

	projectLabel, err := optionalArgLabel(cliCtx, 0, "project")
	if err != nil {
		return err
	}

	resourceLabel, err := optionalArgLabel(cliCtx, 1, "resource")
	if err != nil {
		return err
	}

	dontWait := cliCtx.Bool("no-wait")

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	provisioningClient, err := loadProvisioningClient()
	if err != nil {
		return err
	}

	ps, err := clients.FetchProjects(ctx, marketplaceClient, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch projects list: %s", err), -1)
	}
	projectIdx, _, err := prompts.SelectProject(ps, projectLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select Project")
	}
	p := ps[projectIdx]

	res, err := clients.FetchResources(ctx, marketplaceClient, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of provisioned resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}
	resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select Resource")
	}
	r := res[resourceIdx]

	if err := addProject(ctx, userID, teamID, r, p, provisioningClient, dontWait); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not add resource to project: %s", err), -1)
	}

	if r.Body.ProjectID == nil {
		fmt.Printf("Adding %s to %s\n", resourceLabel, projectLabel)
	} else {
		fmt.Printf("Moving %s to %s\n", resourceLabel, projectLabel)
	}

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

// addProject adds a resource to an existing project
func addProject(ctx context.Context, uid, tid *manifold.ID, r *mModels.Resource,
	p *mModels.Project, provisioningClient *pClient.Provisioning, dontWait bool,
) error {
	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return err
	}

	typeStr := "operation"
	version := int64(1)
	state := "move"
	curTime := strfmt.DateTime(time.Now())
	opBody := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Move{
			ProjectID: &p.ID,
			State:     &state,
		},
	}

	opBody.Body.SetCreatedAt(&curTime)
	opBody.Body.SetUpdatedAt(&curTime)
	opBody.Body.SetResourceID(r.ID)

	if tid == nil {
		opBody.Body.SetUserID(uid)
	} else {
		opBody.Body.SetTeamID(tid)
	}

	op := operation.NewPutOperationsIDParamsWithContext(ctx)
	op.SetBody(opBody)
	op.SetID(ID.String())

	res, err := provisioningClient.Operation.PutOperationsID(op, nil)
	if err != nil {
		switch e := err.(type) {
		case *operation.PutOperationsIDBadRequest:
			return e.Payload
		case *operation.PutOperationsIDUnauthorized:
			return e.Payload
		case *operation.PutOperationsIDNotFound:
			return e.Payload
		case *operation.PutOperationsIDConflict:
			return e.Payload
		case *operation.PutOperationsIDInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	if dontWait {
		return nil
	}

	_, err = waitForOp(ctx, provisioningClient, res.Payload)
	return err
}

// loadMarketplaceClient returns an identify client based on the configuration file.
func loadMarketplaceClient() (*mClient.Marketplace, error) {
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

// loadProvisioningClient returns a provisioning client based on the configuration file.
func loadProvisioningClient() (*pClient.Provisioning, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	provisioningClient, err := clients.NewProvisioning(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Provisioning client: %s", err), -1)
	}

	return provisioningClient, nil
}

// selectProject prompts a user to select a project (if selects the one provided automatically)
func selectProject(ctx context.Context, projectLabel string, teamID *manifold.ID, marketplaceClient *mClient.Marketplace) (*mModels.Project, error) {
	projects, err := clients.FetchProjects(ctx, marketplaceClient, teamID)
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
