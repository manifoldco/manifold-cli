package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

type resourceList struct {
	totalResources int
	totalProjects  int
	groups         []resourceGroup
}

type resourceGroup struct {
	owner     string
	project   string
	resources []*models.Resource
}

func init() {
	listCmd := cli.Command{
		Name:     "list",
		Usage:    "List the status of your provisioned resources",
		Category: "RESOURCES",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, list),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
		}...),
	}

	cmds = append(cmds, listCmd)
}

func list(cliCtx *cli.Context) error {
	ctx := context.Background()

	projectName, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Catalog, api.Identity, api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	// Get catalog
	catalog, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	res, err := clients.FetchResources(ctx, client.Marketplace, teamID, projectName)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of provisioned "+
			"resources: "+err.Error(), -1)
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, client.Provisioning, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
	}

	resources, statuses := buildResourceList(res, oRes)

	list, err := groupResources(ctx, client, resources, teamID)
	if err != nil {
		return err
	}

	fmt.Printf("%d resources in %d projects\n", list.totalResources, list.totalProjects)
	fmt.Println("Use `manifold view [resource-name]` to display resource details")

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	for _, group := range list.groups {
		w.SetStyle(ansiterm.Bold)
		fmt.Fprintf(w, "\n%s", group.owner)
		w.ClearStyle(ansiterm.Bold)

		if group.project != "" {
			fmt.Fprint(w, "/")
			w.SetStyle(ansiterm.Bold)
			fmt.Fprint(w, group.project)
			w.ClearStyle(ansiterm.Bold)
		}

		fmt.Fprintf(w, "\n")

		w.SetForeground(ansiterm.Gray)
		fmt.Fprintln(w, "Name\tTitle\tType\tStatus")
		w.Reset()

		for _, resource := range group.resources {
			rType := "Custom"

			if *resource.Body.Source != "custom" {
				// Get catalog data
				product, err := catalog.GetProduct(*resource.Body.ProductID)
				if err != nil {
					return cli.NewExitError("Product referenced by resource does not exist: "+
						err.Error(), -1)
				}
				if product == nil {
					return cli.NewExitError("Product not found", -1)
				}
				plan, err := catalog.GetPlan(*resource.Body.PlanID)
				if err != nil {
					// Try and get unlisted plan not in local cache
					plan, err = catalog.FetchPlanById(ctx, *resource.Body.PlanID)
					if err != nil {
						return cli.NewExitError("Plan referenced by resource does not exist: "+
							err.Error(), -1)
					}
				}
				if plan == nil {
					return cli.NewExitError("Product not found", -1)
				}

				rType = fmt.Sprintf("%s %s", product.Body.Name, plan.Body.Name)
			}

			status, ok := statuses[resource.ID]
			if !ok {
				status = "Ready"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", resource.Body.Label, resource.Body.Name, rType, status)
		}
	}

	w.Flush()
	return nil
}

func buildResourceList(resources []*models.Resource, operations []*pModels.Operation) (
	[]*models.Resource, map[manifold.ID]string) {
	out := []*models.Resource{}
	statuses := make(map[manifold.ID]string)

	for _, op := range operations {
		switch op.Body.Type() {
		case "provision":
			body := op.Body.(*pModels.Provision)
			if body.State == nil {
				panic("State value was nil")
			}

			// if its a terminal state, then we can just ignore the op
			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[body.ResourceID] = "Creating"
			out = append(out, &models.Resource{
				ID: body.ResourceID,
				Body: &models.ResourceBody{
					CreatedAt: op.Body.CreatedAt(),
					UpdatedAt: op.Body.UpdatedAt(),
					Label:     manifold.Label(*body.Label),
					Name:      manifold.Name(*body.Name),
					Source:    body.Source,
					PlanID:    body.PlanID,
					ProductID: body.ProductID,
					RegionID:  body.RegionID,
					UserID:    op.Body.UserID(),
					TeamID:    op.Body.TeamID(),
					ProjectID: body.ProjectID,
				},
			})
		case "resize":
			body := op.Body.(*pModels.Resize)
			if body.State == nil {
				panic("State value was nil")
			}

			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[body.ResourceID] = "Resizing"
		case "deprovision":
			body := op.Body.(*pModels.Deprovision)
			if body.State == nil {
				panic("State value was nil")
			}

			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[body.ResourceID] = "Deleting"
		}
	}

	for _, r := range resources {
		out = append(out, r)
	}

	return out, statuses
}

func groupResources(ctx context.Context, client *api.API, resources []*models.Resource, teamID *manifold.ID) (resourceList, error) {
	list := resourceList{
		totalResources: len(resources),
	}

	email, err := userEmail(ctx)
	if err != nil {
		return list, err
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return list, cli.NewExitError("Failed to fetch the list of projects: "+err.Error(), -1)
	}

	teams, err := clients.FetchTeams(ctx, client.Identity)
	if err != nil {
		return list, err
	}

	type group struct {
		user    manifold.ID
		team    manifold.ID
		project manifold.ID
	}

	m := make(map[group][]*models.Resource)

	// Group resources by team/me + projects
	for _, r := range resources {
		key := group{}

		if r.Body.UserID != nil {
			key.user = *r.Body.UserID
		} else {
			key.team = *r.Body.TeamID
		}

		if r.Body.ProjectID != nil {
			key.project = *r.Body.ProjectID
		}

		list := m[key]
		list = append(list, r)
		m[key] = list
	}

	// Assemble groups into a single list, sorting resources by name
	var groups []resourceGroup
	for k, v := range m {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Body.Label < v[j].Body.Label
		})

		var owner string
		var project string

		// Find the correct owner, either a team name or the user email
		if !k.user.IsEmpty() {
			owner = email
		} else {
			for _, t := range teams {
				if t.ID == k.team {
					owner = string(t.Body.Label)
				}
			}
		}

		// Find the project name if any
		if !k.project.IsEmpty() {
			list.totalProjects++
			for _, p := range projects {
				if p.ID == k.project {
					project = string(p.Body.Label)
				}
			}
		}

		groups = append(groups, resourceGroup{
			owner:     owner,
			project:   project,
			resources: v,
		})
	}

	// Sort groups by owner, project and name
	sort.Slice(groups, func(i, j int) bool {
		a := groups[i]
		b := groups[j]

		if a.owner != b.owner {
			return a.owner < b.owner
		}

		return a.project < b.project
	})

	list.groups = groups

	return list, nil
}

// userEmail returns the user email based on the authenticated session
func userEmail(ctx context.Context) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return "", cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	// If the session is not a user, we cannot return their identity
	if !s.IsUser() {
		return "", nil
	}

	userDetail := s.LabelInfo()

	if userDetail == nil && len(*userDetail) < 2 {
		return "", cli.NewExitError("Could not retrieve user email", -1)
	}

	return (*userDetail)[1], nil
}
