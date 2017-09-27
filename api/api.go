package api

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	bClient "github.com/manifoldco/manifold-cli/generated/billing/client"
	cClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/product"
	"github.com/manifoldco/manifold-cli/generated/catalog/client/provider"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	conClient "github.com/manifoldco/manifold-cli/generated/connector/client"
	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/urfave/cli"
)

// API is a composition of all clients generated from the spec. Use the `New`
// function to load the necessary clients for the operation.
type API struct {
	ctx          context.Context
	Billing      *bClient.Billing
	Catalog      *cClient.Catalog
	Identity     *iClient.Identity
	Marketplace  *mClient.Marketplace
	Provisioning *pClient.Provisioning
	Connector    *conClient.Connector
}

// Client represents one of the clients generated from the spec.
type Client int

const (
	// Billing represents the billing client
	Billing Client = iota

	// Catalog represents the catalog client
	Catalog

	// Identity represents the identity client
	Identity

	// Marketplace represents the marketplace client
	Marketplace

	// Provisioning represents the provisioning client
	Provisioning

	// Connector represents the connector client
	Connector
)

// New loads all clients passed on the list. If any of the clients fails to load
// an error is returned.
func New(list ...Client) (*API, error) {
	api := &API{
		ctx: context.Background(),
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("Could not load configuration: %s", err)
	}

	for _, e := range list {
		var err error
		switch e {
		case Billing:
			api.Billing, err = clients.NewBilling(cfg)
		case Catalog:
			api.Catalog, err = clients.NewCatalog(cfg)
		case Identity:
			api.Identity, err = clients.NewIdentity(cfg)
		case Marketplace:
			api.Marketplace, err = clients.NewMarketplace(cfg)
		case Provisioning:
			api.Provisioning, err = clients.NewProvisioning(cfg)
		case Connector:
			api.Connector, err = clients.NewConnector(cfg)
		}

		if err != nil {
			msg := fmt.Errorf("Failed to create %s client: %s", e, err)
			return nil, cli.NewExitError(msg, -1)
		}
	}

	return api, nil
}

// String returns a string representation of the client type.
func (c Client) String() string {
	switch c {
	case Billing:
		return "Billing"
	case Catalog:
		return "Catalog"
	case Identity:
		return "Identity"
	case Marketplace:
		return "Marketplace"
	case Provisioning:
		return "Provisioning"
	default:
		return "Unknown"
	}
}

// FetchProviders returns a list of all available providers
func (api *API) FetchProviders() ([]*cModels.Provider, error) {
	params := provider.NewGetProvidersParamsWithContext(api.ctx)
	res, err := api.Catalog.Provider.GetProviders(params)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

// FetchProvider returns a provider based on a label
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

// FetchProducts returns a list of all products for a provider
func (api *API) FetchProducts(providerID string) ([]*cModels.Product, error) {
	params := product.NewGetProductsParamsWithContext(api.ctx)
	params.SetProviderID(&providerID)

	res, err := api.Catalog.Product.GetProducts(params, nil)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}
