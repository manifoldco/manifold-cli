package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"

	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

// FetchOperations returns the resources for the authenticated user
func FetchOperations(ctx context.Context, c *pClient.Provisioning, teamID *manifold.ID) ([]*pModels.Operation, error) {
	res, err := c.Operation.GetOperations(
		operation.NewGetOperationsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*pModels.Operation
	for _, o := range res.Payload {
		if (teamID == nil && o.Body.TeamID() == nil) ||
			(teamID != nil && o.Body.TeamID() != nil && teamID.String() == o.Body.TeamID().String()) {
			results = append(results, o)
		}
	}
	return results, nil
}

// FetchResources returns the resources for the authenticated user
func FetchResources(ctx context.Context, c *mClient.Marketplace, teamID *manifold.ID, projectLabel string) ([]*mModels.Resource, error) {
	var project *mModels.Project

	if projectLabel != "" {
		var err error
		project, err = FetchProjectByLabel(ctx, c, teamID, projectLabel)
		if err != nil {
			return nil, err
		}
	}

	res, err := c.Resource.GetResources(
		resource.NewGetResourcesParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*mModels.Resource
	for _, r := range res.Payload {
		if (teamID == nil && r.Body.TeamID == nil) ||
			(teamID != nil && r.Body.TeamID != nil && teamID.String() == r.Body.TeamID.String()) {
			results = append(results, r)
		}
	}

	var matches []*mModels.Resource
	for _, r := range results {
		id := r.Body.ProjectID

		if project == nil || (id != nil && project != nil && *id == project.ID) {
			matches = append(matches, r)
		}
	}

	return matches, nil
}
