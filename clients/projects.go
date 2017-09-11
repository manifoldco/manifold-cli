package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

// FetchAllProjects returns all user and team projects
func FetchAllProjects(ctx context.Context, c *client.Marketplace) ([]*models.Project, error) {
	params := project.NewGetProjectsParamsWithContext(ctx)
	res, err := c.Project.GetProjects(params, nil)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

// FetchProjects returns all user or team projects
func FetchProjects(ctx context.Context, c *client.Marketplace, teamID *manifold.ID) ([]*models.Project, error) {
	params := project.NewGetProjectsParamsWithContext(ctx)

	if teamID == nil {
		me := true
		params.SetMe(&me)
	} else {
		id := teamID.String()
		params.SetTeamID(&id)
	}

	res, err := c.Project.GetProjects(params, nil)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

// FetchProject returns a project
func FetchProject(ctx context.Context, c *client.Marketplace, id string) (*models.Project, error) {
	params := project.NewGetProjectsIDParamsWithContext(ctx)
	params.SetID(id)

	res, err := c.Project.GetProjectsID(params, nil)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}
