package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/prompts"
)

// FetchProjects returns all user or team projects
func FetchProjects(ctx context.Context, c *client.Marketplace, teamID *manifold.ID, shouldSpin bool) ([]*models.Project, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching Projects")
		spin.Start()
		defer spin.Stop()
	}
	params := project.NewGetProjectsParamsWithContext(ctx)

	if teamID != nil {
		id := teamID.String()
		params.SetTeamID(&id)
	}

	res, err := c.Project.GetProjects(params, nil)
	if err != nil {
		return nil, err
	}

	var results []*models.Project

	for _, p := range res.Payload {
		results = append(results, p)
	}

	return results, nil
}
