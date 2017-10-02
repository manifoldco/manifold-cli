package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/product"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/provider"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/region"
	"github.com/manifoldco/manifold-cli/generated/catalog/models"
)

// FetchProviders returns a list of all available providers.
func (api *API) FetchProviders() ([]*models.Provider, error) {
	params := provider.NewGetProvidersParamsWithContext(api.ctx)
	res, err := api.Catalog.Provider.GetProviders(params)
	if err != nil {
		return nil, err
	}

	providers := res.Payload

	sort.Slice(providers, func(i, j int) bool {
		a := string(providers[i].Body.Label)
		b := string(providers[j].Body.Label)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return providers, nil
}

// FetchProvider returns a provider based on a label.
func (api *API) FetchProvider(label string) (*models.Provider, error) {
	if label == "" {
		return nil, fmt.Errorf("Provider label is missing")
	}

	params := provider.NewGetProvidersParamsWithContext(api.ctx)
	params.SetLabel(&label)
	res, err := api.Catalog.Provider.GetProviders(params)
	if err != nil {
		return nil, err
	}

	if len(res.Payload) == 0 {
		return nil, fmt.Errorf("Provider with label %q not found", label)
	}

	return res.Payload[0], nil
}

// FetchProduct returns a product based on a label.
func (api *API) FetchProduct(label string, providerID string) (*models.Product, error) {
	if label == "" {
		return nil, fmt.Errorf("Product label is missing")
	}

	if providerID == "" {
		return nil, fmt.Errorf("Provider id is missing")
	}

	params := product.NewGetProductsParamsWithContext(api.ctx)
	params.SetLabel(&label)
	params.SetProviderID(&providerID)
	res, err := api.Catalog.Product.GetProducts(params, nil)
	if err != nil {
		return nil, err
	}

	if len(res.Payload) == 0 {
		return nil, fmt.Errorf("Product with label %q not found", label)
	}

	return res.Payload[0], nil
}

// FetchProducts returns a list of all products for a provider.
func (api *API) FetchProducts(providerID string) ([]*models.Product, error) {
	params := product.NewGetProductsParamsWithContext(api.ctx)
	params.SetProviderID(&providerID)

	res, err := api.Catalog.Product.GetProducts(params, nil)
	if err != nil {
		return nil, err
	}

	var products []*models.Product

	for _, p := range res.Payload {
		if *p.Body.State == "available" {
			products = append(products, p)
		}
	}

	sort.Slice(products, func(i, j int) bool {
		a := string(products[i].Body.Label)
		b := string(products[j].Body.Label)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return products, nil
}

// FetchPlans returns a list of all plans for a product.
func (api *API) FetchPlans(productID string) ([]*models.Plan, error) {
	params := plan.NewGetPlansParamsWithContext(api.ctx)
	params.SetProductID([]string{productID})

	res, err := api.Catalog.Plan.GetPlans(params, nil)
	if err != nil {
		return nil, err
	}

	plans := res.Payload

	sort.Slice(plans, func(i, j int) bool {
		a := plans[i]
		b := plans[j]

		if *a.Body.Cost == *b.Body.Cost {
			return strings.ToLower(string(a.Body.Label)) <
				strings.ToLower(string(b.Body.Label))
		}
		return *a.Body.Cost < *b.Body.Cost
	})

	return plans, nil
}

// FetchPlan returns a plan from a product.
func (api *API) FetchPlan(label string, productID string) (*models.Plan, error) {
	if label == "" {
		return nil, fmt.Errorf("Plan label is missing")
	}

	if productID == "" {
		return nil, fmt.Errorf("Product id is missing")
	}

	params := plan.NewGetPlansParamsWithContext(api.ctx)
	params.SetProductID([]string{productID})
	params.SetLabel(&label)

	res, err := api.Catalog.Plan.GetPlans(params, nil)
	if err != nil {
		return nil, err
	}

	if len(res.Payload) == 0 {
		return nil, fmt.Errorf("Plan with label %q not found", label)
	}

	return res.Payload[0], nil
}

// FetchRegions returns a list of all available regions.
func (api *API) FetchRegions() ([]*models.Region, error) {
	params := region.NewGetRegionsParamsWithContext(api.ctx)

	res, err := api.Catalog.Region.GetRegions(params)
	if err != nil {
		return nil, err
	}

	regions := res.Payload

	sort.Slice(regions, func(i, j int) bool {
		a := regions[i]
		b := regions[j]
		return strings.ToLower(string(a.Body.Name)) < strings.ToLower(string(b.Body.Name))
	})

	return regions, nil
}
