package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

// FetchProjects returns all user or team projects
func FetchProjects(ctx context.Context, teamID *manifold.ID, c *client.Marketplace) ([]models.Project, error) {
	params := project.NewGetProjectsParamsWithContext(ctx)

	if teamID != nil {
		id := teamID.String()
		params.SetTeamID(&id)
	}

	res, err := c.Project.GetProjects(params, nil)
	if err != nil {
		return nil, err
	}

	var results []models.Project

	for _, p := range res.Payload {
		results = append(results, *p)
	}

	return results, nil
}
