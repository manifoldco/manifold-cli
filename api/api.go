package api

import (
	"fmt"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	bClient "github.com/manifoldco/manifold-cli/generated/billing/client"
	cClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/urfave/cli"
)

type Api struct {
	*bClient.Billing
	*cClient.Catalog
	*iClient.Identity
	*mClient.Marketplace
	*pClient.Provisioning
}

type Client int

const (
	Billing Client = iota
	Catalog
	Identity
	Marketplace
	Provisioning
)

func New(list ...Client) (*Api, error) {
	api := &Api{}

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
		}

		if err != nil {
			msg := fmt.Errorf("Failed to create %s client: %s", e, err)
			return nil, cli.NewExitError(msg, -1)
		}
	}

	return api, nil
}

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
