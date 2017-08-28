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
// TODO: teamID should be *manifold.ID not ...*manifold.ID but this would break existing API while we work  on adding teams support throughout
func FetchOperations(ctx context.Context, c *pClient.Provisioning, teamID ...*manifold.ID) ([]*pModels.Operation, error) {
	if len(teamID) > 1 {
		panic("Only one team can be provided!")
	}

	res, err := c.Operation.GetOperations(
		operation.NewGetOperationsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*pModels.Operation
	for _, o := range res.Payload {
		if teamID != nil && o.Body.TeamID() != teamID[0] {
			continue
		}
		results = append(results, o)
	}
	return results, nil
}

// FetchResources returns the resources for the authenticated user
// TODO teamID should be *manifold.ID not ...*manifold.ID but this would break existing API while we work on adding teams support throughout
func FetchResources(ctx context.Context, c *mClient.Marketplace, teamID ...*manifold.ID) ([]*mModels.Resource, error) {
	if len(teamID) > 1 {
		panic("Only one team can be provided!")
	}
	res, err := c.Resource.GetResources(
		resource.NewGetResourcesParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*mModels.Resource
	for _, r := range res.Payload {
		// TODO: remove this once CLI has first-class Teams support
		if teamID != nil && r.Body.TeamID != teamID[0] {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}
