package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	billing "github.com/manifoldco/manifold-cli/generated/billing/client"
	catalog "github.com/manifoldco/manifold-cli/generated/catalog/client"
	identity "github.com/manifoldco/manifold-cli/generated/identity/client"
	marketplace "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	provisioning "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/session"
	"github.com/urfave/cli"
)

// loadMarketplaceClient returns marketplace client based on the configuration file.
func loadMarketplaceClient() (*marketplace.Marketplace, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	c, err := clients.NewMarketplace(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	return c, nil
}

// loadProvisioningClient returns a provisioning client based on the configuration file.
func loadProvisioningClient() (*provisioning.Provisioning, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	c, err := clients.NewProvisioning(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Provisioning client: %s", err), -1)
	}

	return c, nil
}

// loadBillingClient returns billing client based on the configuration file.
func loadBillingClient() (*billing.Billing, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	c, err := clients.NewBilling(cfg)
	if err != nil {
		return nil, cli.NewExitError("Failed to create a Billing API client: "+
			err.Error(), -1)
	}

	return c, nil
}

// loadIdentityClient returns an identify client based on the configuration file.
func loadIdentityClient() (*identity.Identity, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	c, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to create Identity client: %s", err), -1)
	}

	return c, nil
}

func loadCatalogClient() (*catalog.Catalog, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	c, err := clients.NewCatalog(cfg)
	if err != nil {
		return nil, cli.NewExitError("Failed to create a Catalog API client: "+
			err.Error(), -1)
	}

	return c, nil
}

func loadAnalytics() (*analytics.Analytics, error) {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load authenticated session: %s", err), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load analytics agent: %s", err), -1)
	}

	return a, nil
}
