package clients

import (
	"context"

	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/identity/client/team"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/prompts"
)

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

	var results []*iModels.Team
	for _, t := range res.Payload {
		results = append(results, t)
	}

	return results, nil
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
