package clients

import (
	"context"

	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/identity/client/team"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
)

// FetchTeams returns the teams for the authenticated user
func FetchTeams(ctx context.Context, c *iClient.Identity) ([]*iModels.Team, error) {
	res, err := c.Team.GetTeams(
		team.NewGetTeamsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*iModels.Team
	for _, t := range res.Payload {
		results = append(results, t)
	}

	return results, nil
}

// FetchMemberships returns all memberships for the authenticated user
func FetchMemberships(ctx context.Context, c *iClient.Identity) ([]iModels.TeamMembership, error) {
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
