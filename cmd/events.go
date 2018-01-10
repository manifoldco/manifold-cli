package main

import (
	"context"
	"fmt"
	"io"
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
				Flags: append(teamFlags, limitFlag(), offsetFlag()),
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

	err = writeEventsList(events, cliCtx.Int("limit"), cliCtx.Int("offset"))
	if err != nil {
		return cli.NewExitError("Could not print activity eventsList: "+err.Error(), -1)
	}

	return nil
}

// writeEventsList prints the state of a event and returns an error if it occurs
func writeEventsList(evts []*events.Event, limit, offset int) error {
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	min, max := limitCollection(len(evts), limit, offset)

	for i := min; i < max; i++ {
		e := evts[i]

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
			printResource(w, body.Data.Resource)
			printSource(w, body.Data.Source)
			printProject(w, body.Data.Project)
			printUser(w, body.Data.User)
			printTeam(w, body.Data.Team)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.Plan, "Summary")
		case *events.OperationDeprovisioned:
			printUser(w, body.Data.User)
			printTeam(w, body.Data.Team)
		case *events.OperationResized:
			printResource(w, body.Data.Resource)
			printProject(w, body.Data.Project)
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.NewPlan, "New Plan")
			printPlan(w, body.Data.Provider, body.Data.Product, body.Data.OldPlan, "Old Plan")
		}

		w.Flush()
	}

	return nil
}

func printResource(w io.Writer, resource *events.Resource) {
	if resource != nil {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Resource"), resource.Name))
	}
}

func printSource(w io.Writer, source string) {
	if source != "" {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Source"), source))
	}
}

func printProject(w io.Writer, project *events.Project) {
	if project != nil {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Project"), project.Name))
	}
}

func printUser(w io.Writer, user *events.User) {
	if user != nil {
		output := fmt.Sprintf("%s (%s)", user.Name, user.Email)
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("User"), output))
	}
}

func printTeam(w io.Writer, team *events.Team) {
	if team != nil {
		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s", color.Faint("Team"), team.Name))
	}
}

func printPlan(w io.Writer, provider *events.Provider, product *events.Product,
	plan *events.Plan, label string) {

	if provider != nil {
		product := product.Name
		name := plan.Name
		val := plan.Cost

		var cost string
		if val == 0 {
			cost = "free"
		} else {
			cost = fmt.Sprint("$%d/month", val)
		}

		fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s (%s %s)", color.Faint(label),
			product, name, cost))
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
