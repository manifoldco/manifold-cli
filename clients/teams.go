package clients

import (
	"context"

	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/identity/client/team"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/prompts"
)

// TeamMembersCount groups a team name with the amount of members the team has
type TeamMembersCount struct {
	Name    string
	Members int
}

// FetchTeams returns the teams for the authenticated user
func FetchTeams(ctx context.Context, c *iClient.Identity, shouldSpin bool) ([]*iModels.Team, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching teams")
		spin.Start()
		defer spin.Stop()
	}
	res, err := c.Team.GetTeams(
		team.NewGetTeamsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// FetchTeamMembers returns a list of members profile from a team.
func FetchTeamMembers(ctx context.Context, id string, c *iClient.Identity, shouldSpin bool) ([]*iModels.MemberProfile, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching memberships")
		spin.Start()
		defer spin.Stop()
	}
	params := team.NewGetTeamsIDMembersParamsWithContext(ctx)
	params.SetID(id)
	res, err := c.Team.GetTeamsIDMembers(params, nil)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// FetchTeamsMembersCount returns a list of all user teams with their names and
// number of members.
func FetchTeamsMembersCount(ctx context.Context, c *iClient.Identity, shouldSpin bool) ([]TeamMembersCount, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching memberships")
		spin.Start()
		defer spin.Stop()
	}
	teams, err := FetchTeams(ctx, c, false)
	if err != nil {
		return nil, err
	}

	// team payload doesn't contain a list of members. In order to find the
	// number for each team we fetch them in parallel.
	res := make(chan TeamMembersCount, len(teams))
	fail := make(chan error)

	for _, t := range teams {
		id := t.ID.String()
		name := string(t.Body.Name)

		go func() {
			members, err := FetchTeamMembers(ctx, id, c, false)

			if err != nil {
				fail <- err
				return
			}

			res <- TeamMembersCount{
				Name:    name,
				Members: len(members),
			}
		}()
	}

	var counts []TeamMembersCount
	for range teams {
		select {
		case err := <-fail:
			return nil, err
		case count := <-res:
			counts = append(counts, count)
		}
	}

	return counts, nil
}

// FetchMemberships returns all memberships for the authenticated user
func FetchMemberships(ctx context.Context, c *iClient.Identity, shouldSpin bool) ([]iModels.TeamMembership, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching memberships")
		spin.Start()
		defer spin.Stop()
	}
	params := team.NewGetMembershipsParamsWithContext(ctx)
	res, err := c.Team.GetMemberships(params, nil)

	if err != nil {
		return nil, err
	}

	var results []iModels.TeamMembership

	for _, m := range res.Payload {
		results = append(results, *m)
	}

	return results, nil
}
