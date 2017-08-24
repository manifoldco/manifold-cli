package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	switchCmd := cli.Command{
		Name:      "switch",
		Usage:     "Switch to a team context",
		ArgsUsage: "[label]",
		Flags: []cli.Flag{
			meFlag(),
		},
		Action: middleware.Chain(middleware.EnsureSession, switchTeamCmd),
	}

	cmds = append(cmds, switchCmd)
}

func switchTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamLabel, err := optionalArgLabel(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	me := cliCtx.Bool("me")

	identityClient, err := clients.NewIdentity(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	teams, err := clients.FetchTeams(ctx, identityClient)
	if err != nil {
		return err
	}

	if len(teams) == 0 {
		return errs.ErrNoTeams
	}

	var team *models.Team
	if !me {
		teamIdx, _, err := prompts.SelectTeam(teams, teamLabel, true)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select team")
		}

		if teamIdx == 0 {
			me = true
		} else {
			teamIdx = teamIdx - 1
			team = teams[teamIdx]
			teamLabel = string(team.Body.Label)
		}
	} else {
		teamLabel = ""
	}

	if err := switchTeam(cfg, teamLabel); err != nil {
		return cli.NewExitError(err, -1)
	}

	if me {
		fmt.Println("You're now operating under your account, not a team.")
	} else {
		fmt.Printf("You're now operating under the \"%s\" team.\n", team.Body.Name)
	}

	return nil
}

func switchTeam(cfg *config.Config, label string) error {
	cfg.Team = label
	if err := cfg.Write(); err != nil {
		return fmt.Errorf("Could not switch teams context: %s", err)
	}

	return nil
}
