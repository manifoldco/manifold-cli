package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"
	iClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

// FetchProjects fetches and returns the projects for an authenticated user
func FetchProjects(ctx context.Context, c *iClient.Marketplace, teamID *manifold.ID) ([]*mModels.Project, error) {
	res, err := c.Project.GetProjects(
		project.NewGetProjectsParamsWithContext(ctx), nil,
	)
	if err != nil {
		return nil, err
	}

	var results []*mModels.Project
	for _, p := range res.Payload {
		if teamID != nil && p.Body.TeamID != nil && teamID.String() == p.Body.TeamID.String() ||
			teamID == nil && p.Body.TeamID == nil {
			results = append(results, p)
		}
	}

	return results, err
}
