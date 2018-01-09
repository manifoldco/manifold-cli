package api

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	activity "github.com/manifoldco/manifold-cli/generated/activity/client"
	billing "github.com/manifoldco/manifold-cli/generated/billing/client"
	catalog "github.com/manifoldco/manifold-cli/generated/catalog/client"
	connector "github.com/manifoldco/manifold-cli/generated/connector/client"
	identity "github.com/manifoldco/manifold-cli/generated/identity/client"
	marketplace "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	provisioning "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/urfave/cli"
)

// API is a composition of all clients generated from the spec. Use the `New`
// function to load the necessary clients for the operation.
type API struct {
	ctx          context.Context
	Activity     *activity.Activity
	Analytics    *analytics.Analytics
	Billing      *billing.Billing
	Catalog      *catalog.Catalog
	Identity     *identity.Identity
	Marketplace  *marketplace.Marketplace
	Provisioning *provisioning.Provisioning
	Connector    *connector.Connector
}

// Client represents one of the clients generated from the spec.
type Client int

const (
	// Activity represents the analytics client
	Activity Client = iota

	// Analytics represents the analytics client
	Analytics

	// Billing represents the billing client
	Billing

	// Catalog represents the catalog client
	Catalog

	// Connector represents the connector client
	Connector

	// Identity represents the identity client
	Identity

	// Marketplace represents the marketplace client
	Marketplace

	// Provisioning represents the provisioning client
	Provisioning
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
		case Activity:
			api.Activity, err = clients.NewActivity(cfg)
		case Analytics:
			api.Analytics, err = api.loadAnalytics(cfg)
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
	case Activity:
		return "Activity"
	case Analytics:
		return "Analytics"
	case Billing:
		return "Billing"
	case Catalog:
		return "Catalog"
	case Connector:
		return "Connector"
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

// Context returns API underlining context.
func (api *API) Context() context.Context {
	return api.ctx
}
