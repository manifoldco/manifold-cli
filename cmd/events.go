package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
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
				Flags: append(teamFlags, limitFlag(), offsetFlag(), verboseFlag()),
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

	err = writeEventsList(cliCtx, events)
	if err != nil {
		return cli.NewExitError("Could not print activity eventsList: "+err.Error(), -1)
	}

	return nil
}

// writeEventsList prints the state of a event and returns an error if it occurs
func writeEventsList(cliCtx *cli.Context, evts []*events.Event) error {
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	limit := cliCtx.Int("limit")
	offset := cliCtx.Int("offset")
	verbose := cliCtx.Bool("verbose")

	min, max := limitCollection(len(evts), limit, offset)

	for i := min; i < max; i++ {
		e := evts[i]

		fmt.Fprintln(w)
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("ID"), e.ID))

		actor := e.Body.Actor()
		if actor != nil {
			output := actor.Name
			if actor.Email != "" {
				output += fmt.Sprintf(" (%s)", actor.Email)
			}
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Actor"), output))
		}
		if verbose || actor == nil {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("ActorID"), e.Body.ActorID()))
		}

		scope := e.Body.Scope()
		if scope != nil {
			output := scope.Name
			if scope.Email != "" {
				output += fmt.Sprintf(" (%s)", scope.Email)
			}
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Scope"), output))
		}
		if verbose || scope == nil {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("ScopeID"), e.Body.ScopeID()))
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
			printResource(w, body.Data.Resource, verbose)
			printSource(w, body.Data.Source)
			printProject(w, body.Data.Project, verbose)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.Plan, "Summary", verbose)
		case *events.OperationDeprovisioned:
			printResource(w, body.Data.Resource, verbose)
			printSource(w, body.Data.Source)
			printProject(w, body.Data.Project, verbose)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.Plan, "Summary", verbose)
		case *events.OperationResized:
			printResource(w, body.Data.Resource, verbose)
			printSource(w, body.Data.Source)
			printProject(w, body.Data.Project, verbose)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.NewPlan, "New Plan", verbose)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.OldPlan, "Old Plan", verbose)
		}

		w.Flush()
	}

	return nil
}

func printResource(w io.Writer, resource *events.Resource, verbose bool) {
	if resource != nil {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Resource"), resource.Name))
		if verbose {
			fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("ResourceID"), resource.ID))
		}
	}
}

func printSource(w io.Writer, source string) {
	if source != "" {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Source"), source))
	}
}

func printProject(w io.Writer, project *events.Project, verbose bool) {
	if project != nil {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Project"), project.Name))
		if verbose {
			fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("ProjectID"), project.ID))
		}
	}
}

func printPlan(w io.Writer, provider *events.Provider, product *events.Product,
	plan *events.Plan, label string, verbose bool) {

	if provider != nil {
		product := product.Name
		name := plan.Name
		val := plan.Cost

		var cost string
		if val == 0 {
			cost = "free"
		} else {
			cost = fmt.Sprintf("$%d/month", val)
		}

		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s (%s %s)", color.Faint(label),
			product, name, cost))
	}

	if verbose {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("ProviderID"), provider.ID))
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("ProductID"), product.ID))
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("PlanID"), plan.ID))
	}
}

func limitCollection(length, limit, offset int) (int, int) {
	min := offset
	if min < 0 {
		min = 0
	}

	max := limit + offset
	if max < 0 {
		max = 0
	}

	if min >= length {
		return 0, 0
	}

	if max > length {
		max = length
	}

	return min, max
}
