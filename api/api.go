package api

import (
	"fmt"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	bClient "github.com/manifoldco/manifold-cli/generated/billing/client"
	cClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	conClient "github.com/manifoldco/manifold-cli/generated/connector/client"
	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/urfave/cli"
)

// API is a composition of all clients generated from the spec. Use the `New`
// function to load the necessary clients for the operation.
type API struct {
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
	api := &API{}

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
