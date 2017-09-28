package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/product"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/provider"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

// FetchProviders returns a list of all available providers.
func (api *API) FetchProviders() ([]*cModels.Provider, error) {
	params := provider.NewGetProvidersParamsWithContext(api.ctx)
	res, err := api.Catalog.Provider.GetProviders(params)
	if err != nil {
		return nil, err
	}

	providers := res.Payload

	sort.Slice(providers, func(i, j int) bool {
		a := string(providers[i].Body.Name)
		b := string(providers[j].Body.Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return providers, nil
}

// FetchProvider returns a provider based on a label.
func (api *API) FetchProvider(label string) (*cModels.Provider, error) {
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
func (api *API) FetchProduct(label string, providerID string) (*cModels.Product, error) {
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
func (api *API) FetchProducts(providerID string) ([]*cModels.Product, error) {
	params := product.NewGetProductsParamsWithContext(api.ctx)
	params.SetProviderID(&providerID)

	res, err := api.Catalog.Product.GetProducts(params, nil)
	if err != nil {
		return nil, err
	}

	products := res.Payload

	sort.Slice(products, func(i, j int) bool {
		a := string(products[i].Body.Label)
		b := string(products[j].Body.Label)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return products, nil
}

// FetchPlans returns a list of all plans for a product.
func (api *API) FetchPlans(productID string) ([]*cModels.Plan, error) {
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
			return strings.ToLower(string(a.Body.Name)) <
				strings.ToLower(string(b.Body.Name))
		}
		return *a.Body.Cost < *b.Body.Cost
	})

	return plans, nil
}
