package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/activity/client/event"
	aModels "github.com/manifoldco/manifold-cli/generated/activity/models"
)

/*on "manifold logs list" get this to select a team to give the logs for
func SelectTeam(teams []*iModels.Team, name string, userTuple *[]string) (int, string, error) {
	return selectTeam(teams, "Select Team", name, userTuple)
}*/

func init() {
	logsCmd := cli.Command{
		Name:     "logs",
		Usage:    "Logs gets the most recent information of your activity",
		Category: "RESOURCES",
		ArgsUsage: "[team-name]",
		Flags: []cli.Flag{
			meFlag(),
		},
		Action: middleware.Chain(middleware.EnsureSession, logsCmd),
	}

	cmds = append(cmds, logsCmd)
}

func logsCmd(cliCtx *cli.Context) error {
	ctx := context.Background()


	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamName, err := optionalArgName(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	me := cliCtx.Bool("me")

	var team *models.Team
	/*if !me {

	}*/

	if err := logs(ctx, team); err != nil {
		return cli.NewExitError(err, -1)
	}

	if me {
		fmt.Println("These are logs under your personal account.")
	} else {
		fmt.Printf("These are logs under the \"%s\" team.\n", team.Body.Name)
	}

	return nil
}

func logs(ctx context.Context, team *models.Team) error{
	
	client, err := api.New(api.Analytics, api.Marketplace)
	if err != nil {
		return err
	}

	//get the event logs for this team
	prompts.SpinStart("Fetching Resources")
	activities, err := clients.FetchActivities(ctx, client.Activity, team.ID)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve activity logs: "+err.Error(), -1)
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Body.Log < activities[j].Body.Log
	})

	params := map[string]string{}
	if team != nil {
		params["team-name"] = string(team.Body.Name)
	} else {
		params["team-name"] = ""
	}
	client.Analytics.Track(ctx, "Displayed logs", &params)

	return nil
}
