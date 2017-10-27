package catalog

import (
	"context"
	"errors"

	"github.com/manifoldco/go-manifold"
	hierr "github.com/reconquest/hierr-go"

	catalogClientPlan "github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	catalogModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

func (c *Catalog) FetchPlanByLabel(ctx context.Context,
	productID manifold.ID, planLabel string) (*catalogModels.Plan, error) {
	// Get plans for known productIDs
	planParams := catalogClientPlan.NewGetPlansParamsWithContext(ctx)
	planParams.SetProductID([]string{productID.String()})
	planParams.SetLabel(&planLabel)
	plans, err := c.client.Plan.GetPlans(planParams, nil)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the latest product plan data")
	}
	if len(plans.Payload) < 1 {
		return nil, errors.New("Plan does not exist")
	}
	return plans.Payload[0], nil
}

func (c *Catalog) FetchPlanById(ctx context.Context, planID manifold.ID) (*catalogModels.Plan, error) {
	// Get plan by id
	planParams := catalogClientPlan.NewGetPlansIDParamsWithContext(ctx)
	planParams.SetID(planID.String())
	plan, err := c.client.Plan.GetPlansID(planParams, nil)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the latest product plan data")
	}
	return plan.Payload, nil
}
