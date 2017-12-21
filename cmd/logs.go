package main

import (
	"context"
	"fmt"
	"os"
	"io"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/go-manifold/events"
	"github.com/manifoldco/manifold-cli/generated/identity/models"

	//"github.com/manifoldco/manifold-cli/generated/activity/client/events"
	//aModels "github.com/manifoldco/manifold-cli/generated/activity/models"
)

func init() {
	logsCmd := cli.Command{
		Name:     "logs",
		Usage:    "Logs gets the most recent information of your activity",
		Category: "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, logs),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
		}...),
	}

	cmds = append(cmds, logsCmd)
}

func logs(cliCtx *cli.Context) error {
	ctx := context.Background()

	projectName, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	var team *models.Team
	
//
	cfg, err := config.LoadIgnoreLegacy()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	activityClient, err := clients.NewActivity(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	client, err := api.New(api.Analytics, api.Marketplace)
	if err != nil {
		return err
	}

	prompts.SpinStart("Fetching Activities")
	activities, err := clients.FetchActivitiesWithProject(ctx, client.Marketplace, activityClient, teamID, projectName)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}

	if projectName == "" {
		activities = filterActivitiesWithoutProjects(activities)
	}

	w := os.Stdout

	err = writeLogs(w, activities, "Type: %+v, State: %+v, RefID: %+v, Created at: %+v\n")
	if err != nil {
		return cli.NewExitError("Could not print activity logs: "+err.Error(), -1)
	}
//

	fmt.Printf("These are logs under the \"%s\" team.\n", team.Body.Name)
	return nil
}

// writeLogs prints the state of a event and returns an error if it occurs
func writeLogs(w io.Writer, events []*events.Event, format string) error {
	for _, c := range events {
			
			name := c.Body.Type
			state := c.Body.State
			ref := c.Body.RefID
			created := c.Body.CreatedAt

			fmt.Fprintf(w, format, name, state, ref, created)

		fmt.Fprintf(w, "\n")
	}

	return nil
}

// filterResourcesWithoutProjects returns resources without a project id
func filterActivitiesWithoutProjects(activities []*events.Event) []*events.Event {
	var results []*events.Event
	for _, a := range activities {
		if a.Body.RefID == nil {
			results = append(results, a)
		}
	}
	return results
}
