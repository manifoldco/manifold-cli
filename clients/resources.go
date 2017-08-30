package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"

	"github.com/manifoldco/manifold-cli/prompts"

	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

// FetchOperations returns the resources for the authenticated user
func FetchOperations(ctx context.Context, c *pClient.Provisioning, teamID *manifold.ID, shouldSpin bool) ([]*pModels.Operation, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching operations")
		spin.Start()
		defer spin.Stop()
	}
	res, err := c.Operation.GetOperations(
		operation.NewGetOperationsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*pModels.Operation
	for _, o := range res.Payload {
		if teamID != nil && o.Body.TeamID() != nil && teamID.String() == o.Body.TeamID().String() ||
			teamID == nil && o.Body.TeamID() == nil {
			results = append(results, o)
		}
	}
	return results, nil
}

// FetchResources returns the resources for the authenticated user
func FetchResources(ctx context.Context, c *mClient.Marketplace, teamID *manifold.ID, shouldSpin bool) ([]*mModels.Resource, error) {
	if shouldSpin {
		spin := prompts.NewSpinner("Fetching resources")
		spin.Start()
		defer spin.Stop()
	}
	res, err := c.Resource.GetResources(
		resource.NewGetResourcesParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*mModels.Resource
	for _, r := range res.Payload {
		if teamID != nil && r.Body.TeamID != nil && teamID.String() == r.Body.TeamID.String() ||
			teamID == nil && r.Body.TeamID == nil {
			results = append(results, r)
		}
	}
	return results, nil
}
