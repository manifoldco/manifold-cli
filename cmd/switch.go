package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	switchCmd := cli.Command{
		Name:      "switch",
		Usage:     "Switch to a team context",
		Category:  "ADMINISTRATIVE",
		ArgsUsage: "[team-name]",
		Flags: []cli.Flag{
			meFlag(),
		},
		Action: middleware.Chain(middleware.EnsureSession, switchTeamCmd),
	}

	cmds = append(cmds, switchCmd)
}

func switchTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.LoadIgnoreLegacy()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamName, err := optionalArgName(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	me := cliCtx.Bool("me")

	var team *models.Team
	if !me {
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

		s, err := session.Retrieve(ctx, cfg)
		if err != nil {
			return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
		}

		teamIdx, _, err := prompts.SelectContext(teams, teamName, s.LabelInfo())
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select context")
		}

		if teamIdx == -1 {
			me = true
		} else {
			team = teams[teamIdx]
		}
	}

	if err := switchTeam(ctx, cfg, team); err != nil {
		return cli.NewExitError(err, -1)
	}

	if me {
		fmt.Println("You're now operating under your personal account.")
	} else {
		fmt.Printf("You're now operating under the \"%s\" team.\n", team.Body.Name)
	}

	return nil
}

func switchTeam(ctx context.Context, cfg *config.Config, team *models.Team) error {
	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return err
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return err
	}

	cfg.TeamID = ""
	cfg.TeamTitle = ""
	cfg.TeamName = ""
	if team != nil {
		cfg.TeamID = team.ID.String()
		cfg.TeamTitle = string(team.Body.Name)
		cfg.TeamName = string(team.Body.Label)
	}

	if err := cfg.Write(); err != nil {
		return fmt.Errorf("Could not switch context: %s", err)
	}

	params := map[string]string{}
	if team != nil {
		params["team-name"] = cfg.TeamName
		params["team-title"] = cfg.TeamTitle
	} else {
		params["team-name"] = ""
		params["team-title"] = ""
	}
	a.Track(ctx, "Switched Context", &params)

	return nil
}
