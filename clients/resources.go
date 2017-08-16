package clients

import (
	"context"

	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

// FetchOperations returns the resources for the authenticated user
func FetchOperations(ctx context.Context, c *pClient.Provisioning) ([]*pModels.Operation, error) {
	res, err := c.Operation.GetOperations(
		operation.NewGetOperationsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*pModels.Operation
	for _, o := range res.Payload {
		// TODO: remove this once CLI has first-class Teams support
		if o.Body.TeamID() != nil {
			continue
		}
		results = append(results, o)
	}
	return results, nil
}

// FetchResources returns the resources for the authenticated user
func FetchResources(ctx context.Context, c *mClient.Marketplace) ([]*mModels.Resource, error) {
	res, err := c.Resource.GetResources(
		resource.NewGetResourcesParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*mModels.Resource
	for _, r := range res.Payload {
		// TODO: remove this once CLI has first-class Teams support
		if r.Body.TeamID != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}
