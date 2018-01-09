package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/events"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	eventsCmd := cli.Command{
		Name:     "events",
		Usage:    "Review the most recent activity of your account",
		Category: "ADMINISTRATIVE",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List all events",
				Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
					middleware.LoadTeamPrefs, eventsList),
				Flags: teamFlags,
			},
		},
	}

	cmds = append(cmds, eventsCmd)
}

func eventsList(cliCtx *cli.Context) error {
	var scopeID manifold.ID

	me := cliCtx.Bool("me")
	if me {
		userID, err := loadUserID(context.Background())
		if err != nil {
			return err
		}
		scopeID = *userID
	} else {
		teamID, err := validateTeamID(cliCtx)
		if err != nil {
			return err
		}
		scopeID = *teamID
	}

	client, err := api.New(api.Activity, api.Identity)
	if err != nil {
		return err
	}

	prompts.SpinStart("Fetching Activities")
	events, err := client.Events(scopeID)

	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve events: "+err.Error(), -1)
	}

	err = writeEventsList(events)
	if err != nil {
		return cli.NewExitError("Could not print activity eventsList: "+err.Error(), -1)
	}

	return nil
}

// writeEventsList prints the state of a event and returns an error if it occurs
func writeEventsList(evts []*events.Event) error {
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	for _, e := range evts {
		fmt.Fprintln(w)
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("ID"), e.ID))

		actor := e.Body.Actor()

		if actor == nil {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Actor"), e.Body.ActorID()))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Actor"), actor.Name))
		}

		scope := e.Body.Scope()

		if scope == nil {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Scope"), e.Body.ScopeID()))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Scope"), scope.Name))
		}

		source := e.Body.Source()
		if source != nil {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Source"), *source))
		}

		ip := e.Body.IPAddress()
		if ip != "" {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("IP"), ip))
		}

		cat := e.Body.CreatedAt()
		if cat != nil {
			t := time.Time(*cat)
			date := t.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Date"), date))
		}

		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Type"), color.Bold(e.Body.Type())))

		switch body := e.Body.(type) {
		case *events.OperationProvisioned:
			if body.Data.Resource != nil {
				resource := body.Data.Resource.Name
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Resource"), resource))
			}

			if body.Data.Source != "" {
				source := body.Data.Source
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Source"), source))
			}

			if body.Data.Project != nil {
				project := body.Data.Project.Name
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Project"), project))
			}

			if body.Data.User != nil {
				user := fmt.Sprintf("%s (%s)", body.Data.User.Name, body.Data.User.Email)
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("User"), user))
			}

			if body.Data.Team != nil {
				team := body.Data.Team.Name
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Team"), team))
			}

			if body.Data.Provider != nil {
				product := body.Data.Product.Name
				plan := body.Data.Plan.Name
				cost := body.Data.Plan.Cost
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s (%s $%d/month)",
					color.Faint("Summary"), product, plan, cost))
			}
		default:
			return fmt.Errorf("Unrecognized Event Type: %s", e.Body.Type())
		}

		w.Flush()
	}

	return nil
}
