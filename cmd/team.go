package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"strings"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/identity/client"
	teamClient "github.com/manifoldco/manifold-cli/generated/identity/client/team"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:  "teams",
		Usage: "Manage Teams in Manifold",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				ArgsUsage: "[name]",
				Action:    middleware.Chain(middleware.EnsureSession, createTeamCmd),
			},
			{
				Name:      "update",
				ArgsUsage: "[label]",
				Flags: []cli.Flag{
					nameFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, updateTeamCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func createTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamName, err := optionalArgLabel(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	identityClient, err := clients.NewIdentity(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	autoSelect := teamName != ""
	teamName, err = prompts.TeamName(teamName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to name team")
	}

	if err := createTeam(ctx, teamName, identityClient); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not create team: %s", err), -1)
	}

	fmt.Printf("Your team '%s' has been created\n", teamName)
	return nil
}

func updateTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamName, err := optionalArgLabel(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	newTeamName, err := validateName(cliCtx, "name", "team")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	identityClient, err := clients.NewIdentity(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	teams, err := clients.FetchTeams(ctx, identityClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch list of teams: %s", err), -1)
	}

	if len(teams) == 0 {
		return errs.ErrNoTeams
	}

	teamIdx, _, err := prompts.SelectTeam(teams, teamName, false)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select team")
	}

	autoSelect := newTeamName != ""
	newTeamName, err = prompts.TeamName(newTeamName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not validate name")
	}

	if err := updateTeam(ctx, teams[teamIdx], newTeamName, identityClient); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not update team: %s", err), -1)
	}

	fmt.Printf("Your team \"%s\" has been updated\n", newTeamName)
	return nil
}

func createTeam(ctx context.Context, teamName string, identityClient *client.Identity) error {

	createTeam := &models.CreateTeam{
		Body: &models.CreateTeamBody{
			Name:  manifold.Name(teamName),
			Label: manifold.Label(strings.Replace(strings.ToLower(teamName), " ", "-", -1)),
		},
	}

	c := teamClient.NewPostTeamsParamsWithContext(ctx)
	c.SetBody(createTeam)

	_, err := identityClient.Team.PostTeams(c, nil)
	if err != nil {

		switch e := err.(type) {
		case *teamClient.PostTeamsBadRequest:
			return e.Payload
		case *teamClient.PostTeamsUnauthorized:
			return e.Payload
		case *teamClient.PostTeamsConflict:
			return e.Payload
		case *teamClient.PostTeamsInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	return nil
}

func updateTeam(ctx context.Context, team *models.Team, teamName string, identityClient *client.Identity) error {
	updateTeam := &models.UpdateTeam{
		Body: &models.UpdateTeamBody{
			Name:  manifold.Name(teamName),
			Label: manifold.Label(strings.Replace(strings.ToLower(teamName), " ", "-", -1)),
		},
	}

	c := teamClient.NewPatchTeamsIDParamsWithContext(ctx)
	c.SetBody(updateTeam)
	c.SetID(team.ID.String())

	_, err := identityClient.Team.PatchTeamsID(c, nil)
	if err != nil {
		switch e := err.(type) {
		case *teamClient.PatchTeamsIDBadRequest:
			return e.Payload
		case *teamClient.PatchTeamsIDInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	return nil
}
