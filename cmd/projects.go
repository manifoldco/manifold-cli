package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"time"

	"github.com/juju/ansiterm"

	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/color"
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
				Usage:     "Create a new project",
				Flags:     teamFlags,
				ArgsUsage: "[name]",
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, createProjectCmd),
			},
			{
				Name:  "list",
				Usage: "List projects",
				Flags: append(teamFlags, cli.BoolFlag{
					Name:  "all",
					Usage: "List all your projects and teams projects",
				}),
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, listProjectsCmd),
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
				Name:      "delete",
				Usage:     "Delete a project",
				Flags:     teamFlags,
				ArgsUsage: "[name]",
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, deleteProjectCmd),
			},
			{

				Name:      "add",
				Usage:     "Adds or moves a resource to a project",
				ArgsUsage: "[project-label] [resource-label]",
				Flags: append(teamFlags, []cli.Flag{
					skipFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, addProjectCmd),
			},
			{
				Name:  "remove",
				Usage: "Removes a resource from a project",
				Flags: append(teamFlags, []cli.Flag{
					skipFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, removeProjectCmd),
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

	spin.Stop()
	fmt.Printf("Your project '%s' has been created\n", projectName)
	return nil
}

func listProjectsCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	client, err := api.New(api.Marketplace)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	var projects []*mModels.Project

	prompts.SpinStart("Fetching Projects")
	if cliCtx.Bool("all") {
		projects, err = clients.FetchAllProjects(ctx, client.Marketplace)
	} else {
		projects, err = clients.FetchProjects(ctx, client.Marketplace, teamID)
	}
	prompts.SpinStop()

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "%s\n\n", color.Bold("Project"))

	for _, project := range projects {
		fmt.Fprintf(w, "%s\n", project.Body.Label)
	}
	return w.Flush()
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

	client, err := api.New(api.Marketplace)
	if err != nil {
		return err
	}

	p, err := selectProject(ctx, projectLabel, teamID, client.Marketplace)
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

	spin.Stop()
	fmt.Printf("\nYour project \"%s\" has been updated\n", newProjectName)
	return nil
}

func deleteProjectCmd(cliCtx *cli.Context) error {
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

	projectLabel, err := optionalArgName(cliCtx, 0, "project")
	if err != nil {
		return err
	}

	client, err := api.New(api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	p, err := selectProject(ctx, projectLabel, teamID, client.Marketplace)
	if err != nil {
		return err
	}

	spin := prompts.NewSpinner(fmt.Sprintf("Deleting %s", p.Body.Label))
	spin.Start()
	defer spin.Stop()

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return err
	}

	typeStr := "operation"
	version := int64(1)
	state := "delete"
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.ProjectDelete{
			ProjectID: p.ID,
			State:     &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	if teamID == nil {
		op.Body.SetUserID(userID)
	} else {
		op.Body.SetTeamID(teamID)
	}

	d := operation.NewPutOperationsIDParamsWithContext(ctx)
	d.SetBody(op)
	d.SetID(ID.String())

	res, err := client.Provisioning.Operation.PutOperationsID(d, nil)
	if err != nil {
		switch e := err.(type) {
		case *operation.PutOperationsIDBadRequest:
			return cli.NewExitError(e.Payload, -1)
		case *operation.PutOperationsIDUnauthorized:
			return cli.NewExitError(e.Payload, -1)
		case *operation.PutOperationsIDNotFound:
			return cli.NewExitError(e.Payload, -1)
		case *operation.PutOperationsIDConflict:
			return cli.NewExitError(e.Payload, -1)
		case *operation.PutOperationsIDInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	waitForOp(ctx, client.Provisioning, res.Payload)
	spin.Stop()
	fmt.Printf("Your project '%s' has been deleted\n", p.Body.Label)
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

	client, err := api.New(api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	ps, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch projects list: %s", err), -1)
	}

	p, err := selectProject(ctx, projectLabel, teamID, client.Marketplace)
	if err != nil {
		return err
	}

	res, err := clients.FetchResources(ctx, client.Marketplace, teamID, "")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of provisioned resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}
	resourceIdx, _, err := prompts.SelectResource(res, ps, resourceLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select Resource")
	}
	r := res[resourceIdx]

	if err := updateResourceProject(ctx, userID, teamID, r, p, client.Provisioning, dontWait); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not add resource to project: %s", err), -1)
	}

	if r.Body.ProjectID == nil {
		fmt.Printf("Adding %s to %s\n", r.Body.Label, p.Body.Label)
	} else {
		fmt.Printf("Moving %s to %s\n", r.Body.Label, p.Body.Label)
	}

	analytics, err := loadAnalytics()
	if err != nil {
		return err
	}

	params := map[string]string{
		"resource_id":    r.ID.String(),
		"resource_name":  string(r.Body.Name),
		"resource_label": string(r.Body.Label),
		"project_id":     p.ID.String(),
		"project_name":   string(p.Body.Name),
		"project_label":  string(p.Body.Label),
	}
	analytics.Track(ctx, "Added a Resource to a Project", &params)

	return nil
}

func removeProjectCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceLabel, err := optionalArgLabel(cliCtx, 0, "resource")
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

	client, err := api.New(api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	resources, err := clients.FetchResources(ctx, client.Marketplace, teamID, "")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of provisioned resource: %s", err), -1)
	}
	if len(resources) == 0 {
		return errs.ErrNoResources
	}

	var filtered []*mModels.Resource
	for _, r := range resources {
		if r.Body.ProjectID != nil {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		return cli.NewExitError("No resources with projects found", -1)
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}
	idx, _, err := prompts.SelectResource(filtered, projects, resourceLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select Resource")
	}
	r := filtered[idx]

	if err := updateResourceProject(ctx, userID, teamID, r, nil, client.Provisioning, dontWait); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not remove the project from the resource: %s", err), -1)
	}

	fmt.Printf("Removed %s from project\n", r.Body.Label)

	analytics, err := loadAnalytics()
	if err != nil {
		return err
	}

	params := map[string]string{
		"resource_id":    r.ID.String(),
		"resource_name":  string(r.Body.Name),
		"resource_label": string(r.Body.Label),
	}
	analytics.Track(ctx, "Removed a Resource from a Project", &params)

	return nil
}

func createProject(params *projectClient.PostProjectsParams) error {
	client, err := api.New(api.Marketplace)
	if err != nil {
		return err
	}

	_, err = client.Marketplace.Project.PostProjects(params, nil)
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
	client, err := api.New(api.Marketplace)
	if err != nil {
		return err
	}

	_, err = client.Marketplace.Project.PatchProjectsID(params, nil)
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

// updateResourceProject updates a resource to add or remove an existing project
func updateResourceProject(ctx context.Context, uid, tid *manifold.ID, r *mModels.Resource,
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
	}

	if p != nil {
		opBody.Body = &pModels.Move{
			ResourceID: r.ID,
			ProjectID:  &p.ID,
			State:      &state,
		}
	} else {
		opBody.Body = &pModels.Move{
			ResourceID: r.ID,
			State:      &state,
		}
	}

	opBody.Body.SetCreatedAt(&curTime)
	opBody.Body.SetUpdatedAt(&curTime)

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

// selectProject prompts a user to select a project (if selects the one provided automatically)
func selectProject(ctx context.Context, projectLabel string, teamID *manifold.ID, marketplaceClient *mClient.Marketplace) (*mModels.Project, error) {
	projects, err := clients.FetchProjects(ctx, marketplaceClient, teamID)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	if len(projects) == 0 {
		return nil, errs.ErrNoProjects
	}

	idx, _, err := prompts.SelectProject(projects, projectLabel, false)
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
