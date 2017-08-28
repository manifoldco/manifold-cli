package main

import (
	"context"
	"errors"
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
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
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
			{
				Name:      "invite",
				ArgsUsage: "[email] [name]",
				Usage:     "Invite a new user to join your team",
				Flags: []cli.Flag{
					teamFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, inviteToTeamCmd),
			},
			{
				Name:      "leave",
				ArgsUsage: "[name]",
				Usage:     "Remove yourself from a team",
				Action:    middleware.Chain(middleware.EnsureSession, leaveTeamCmd),
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

	identityClient, err := loadIdentityClient()
	if err != nil {
		return err
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

	identityClient, err := loadIdentityClient()
	if err != nil {
		return err
	}

	team, err := selectTeam(ctx, teamName, identityClient)
	if err != nil {
		return err
	}

	autoSelect := newTeamName != ""
	newTeamName, err = prompts.TeamName(newTeamName, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not validate name")
	}

	if err := updateTeam(ctx, team, newTeamName, identityClient); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not update team: %s", err), -1)
	}

	fmt.Printf("Your team \"%s\" has been updated\n", newTeamName)
	return nil
}

func inviteToTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	email, err := optionalArgEmail(cliCtx, 0, "user")
	if err != nil {
		return err
	}

	args := cliCtx.Args().Tail()
	name := strings.Join(args, " ")

	teamName, err := validateName(cliCtx, "team")
	if err != nil {
		return err
	}

	identityClient, err := loadIdentityClient()
	if err != nil {
		return err
	}

	team, err := selectTeam(ctx, teamName, identityClient)
	if err != nil {
		return err
	}

	if email == "" {
		email, err = prompts.Email("")
		if err != nil {
			return err
		}
	}

	if name == "" {
		name, err = prompts.FullName("")
		if err != nil {
			return err
		}
	}

	if err := inviteToTeam(ctx, team, email, name, identityClient); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not invite to team: %s", err), -1)
	}

	fmt.Printf("An invite has been sent to %s <%s>\n", name, email)
	return nil
}

func leaveTeamCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamName, err := optionalArgLabel(cliCtx, 0, "team")
	if err != nil {
		return err
	}

	identityClient, err := loadIdentityClient()
	if err != nil {
		return err
	}

	team, err := selectTeam(ctx, teamName, identityClient)
	if err != nil {
		return err
	}

	memberships, err := clients.FetchMemberships(ctx, identityClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch user memberships: %s", err), -1)
	}

	var membershipID manifold.ID

	for _, m := range memberships {
		if m.Body.TeamID == team.ID {
			membershipID = m.ID
		}
	}

	if membershipID.IsEmpty() {
		return cli.NewExitError("No memberships found", -1)
	}

	if err := leaveTeam(ctx, membershipID, identityClient); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not leave team: %s", err), -1)
	}

	fmt.Printf("You have left the team \"%s\"\n", team.Body.Name)
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

func inviteToTeam(ctx context.Context, team *models.Team, email string, name string, identityClient *client.Identity) error {

	// TODO - luiz
	return errors.New("not implemented")
}

func leaveTeam(ctx context.Context, membershipID manifold.ID, identityClient *client.Identity) error {
	c := teamClient.NewDeleteMembershipsIDParamsWithContext(ctx)
	c.SetID(membershipID.String())

	_, err := identityClient.Team.DeleteMembershipsID(c, nil)

	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *teamClient.DeleteMembershipsIDUnauthorized:
		return e.Payload
	case *teamClient.DeleteMembershipsIDInternalServerError:
		return errs.ErrSomethingWentHorriblyWrong
	default:
		return err
	}
}

// loadIdentityClient returns an identify client based on the configuration file.
func loadIdentityClient() (*client.Identity, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	identityClient, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	return identityClient, nil
}

// fetchTeams retrieves all user's team and prompt to select which team the cmd
// will be applied to.
func selectTeam(ctx context.Context, teamName string, identityClient *client.Identity) (*iModels.Team, error) {
	teams, err := clients.FetchTeams(ctx, identityClient)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to fetch list of teams: %s", err), -1)
	}

	if len(teams) == 0 {
		return nil, errs.ErrNoTeams
	}

	idx, _, err := prompts.SelectTeam(teams, teamName, false)
	if err != nil {
		return nil, prompts.HandleSelectError(err, "Could not select team")
	}

	team := teams[idx]

	return team, nil
}
