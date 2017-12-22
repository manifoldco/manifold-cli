package main

import (
	"context"
	"fmt"
	"os"
	"io"

	"github.com/urfave/cli"
	"github.com/juju/ansiterm"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/go-manifold/events"
	//"github.com/manifoldco/manifold-cli/generated/identity/models"

	//"github.com/manifoldco/manifold-cli/generated/activity/client/events"
	//aModels "github.com/manifoldco/manifold-cli/generated/activity/models"
)

func init() {
	eventsCmd := cli.Command{
		Name:     "events",
		Usage:    "Events List gets the most recent information of your activity",
		Category: "ADMINISTRATIVE",
		Subcommands: []cli.Command{
			{
				Name:      "list",
				Usage:     "List all events",
				Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
					middleware.LoadTeamPrefs, eventsList),
				Flags: append(teamFlags, []cli.Flag{
				projectFlag(),
				}...),
			},
		},
	}

	cmds = append(cmds, eventsCmd)
}

func eventsList(cliCtx *cli.Context) error {
	ctx := context.Background()

	projectName, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

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
	
	w := os.Stdout
	
	err = writeEventsList(w, activities)
	if err != nil {
		return cli.NewExitError("Could not print activity eventsList: "+err.Error(), -1)
	}
	
	fmt.Printf("These are EventsList under the \"%s\" team and \"%s\" project.\n", teamID, projectName)
	return nil
}

// writeEventsList prints the state of a event and returns an error if it occurs
func writeEventsList(w io.Writer, allEvents []*events.Event) error {

	f := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	for _, c := range allEvents {

			name 		:= c.Body.Type()
			//ref 		:= c.Body.RefID()
			scope		:= c.Body.ScopeID()
			created 	:= c.Body.CreatedAt()
			id 			:= c.ID
			Actor 		:= ""//c.GetActor()
			resource 	:= ""
			user 		:= ""
			product 	:= ""
			project 	:= ""
			
			f.SetStyle(ansiterm.Bold)
			fmt.Fprintf(f, "ID:")
			f.ClearStyle(ansiterm.Bold)
			f.Reset()
			fmt.Fprintf(f, "\t%s\n", id)
			
			f.SetStyle(ansiterm.Bold)
			fmt.Fprintf(f, "Actor:")
			f.ClearStyle(ansiterm.Bold)
			f.Reset()
			fmt.Fprintf(f, "\t%s\n", Actor)
			
			f.SetStyle(ansiterm.Bold)
			fmt.Fprintf(f, "Scope:")
			f.ClearStyle(ansiterm.Bold)
			f.Reset()
			fmt.Fprintf(f, "\t%s\n", scope)
			
			f.SetStyle(ansiterm.Bold)
			fmt.Fprintf(f, "Date:")
			f.ClearStyle(ansiterm.Bold)
			f.Reset()
			fmt.Fprintf(f, "\t%s\n", created)

			f.SetStyle(ansiterm.Bold)
			fmt.Fprintf(f, "Type:")
			f.ClearStyle(ansiterm.Bold)
			f.Reset()
			fmt.Fprintf(f, "\t%s\n", name)
			
			switch Op := c.Body.(type) {
				case *events.OperationProvisioned:
					if Op.Data.Resource != nil {
						resource 	= Op.Data.Resource.Name
						f.SetStyle(ansiterm.Bold)
						fmt.Fprintf(f, "\tResource:")
						f.ClearStyle(ansiterm.Bold)
						f.Reset()
						fmt.Fprintf(f, "\t%s\n", resource)
					}

					if Op.Data.User != nil {
						user = Op.Data.User.Name 	
							f.SetStyle(ansiterm.Bold)
							fmt.Fprintf(f, "\tUser:")
							f.ClearStyle(ansiterm.Bold)
							f.Reset()
							fmt.Fprintf(f, "\t%s\n", user)
					}

					if Op.Data.Project != nil {
						project = Op.Data.Project.Name
							f.SetStyle(ansiterm.Bold)
							fmt.Fprintf(f, "\tProject:")
							f.ClearStyle(ansiterm.Bold)
							f.Reset()
							fmt.Fprintf(f, "\t%s\n", project)
					}

					if Op.Data.Product != nil {
						product = Op.Data.Product.Name
							f.SetStyle(ansiterm.Bold)
							fmt.Fprintf(f, "\tSummary:")
							f.ClearStyle(ansiterm.Bold)
							f.Reset()
							fmt.Fprintf(f, "\t%s\n", product)
							
					}
				default:
					return fmt.Errorf("Unrecognized Operation Type: %s", c.Body.Type())
				}

		fmt.Fprintf(f, "\n")
		f.Flush()
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
