package clients

import (
	"net/url"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/manifoldco/manifold-cli/config"
	catalogClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	identityClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	marketplaceClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
)

// NewIdentity returns a swagger generated client for the Identity service
func NewIdentity(cfg *config.Config) (*identityClient.Identity, error) {
	identityURL := cfg.TransportScheme + "://api.identity." + cfg.Hostname + "/v1"
	u, err := url.Parse(identityURL)
	if err != nil {
		return nil, err
	}

	c := identityClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return identityClient.New(transport, strfmt.Default), nil
}

// NewBearerToken returns a bearer token authenticator for use with a
// go-swagger generated client.
func NewBearerToken(token string) runtime.ClientAuthInfoWriter {
	return httptransport.BearerToken(token)
}

// NewMarketplace returns a swagger generated client for the Marketplace service
func NewMarketplace(cfg *config.Config) (*marketplaceClient.Marketplace, error) {
	marketplaceURL := cfg.TransportScheme + "://api.marketplace." + cfg.Hostname + "/v1"
	u, err := url.Parse(marketplaceURL)
	if err != nil {
		return nil, err
	}

	c := marketplaceClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return marketplaceClient.New(transport, strfmt.Default), nil
}

// NewCatalog returns a swagger generated client for the Catalog service
func NewCatalog(cfg *config.Config) (*catalogClient.Catalog, error) {
	marketplaceURL := cfg.TransportScheme + "://api.catalog." + cfg.Hostname + "/v1"
	u, err := url.Parse(marketplaceURL)
	if err != nil {
		return nil, err
	}

	c := catalogClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return catalogClient.New(transport, strfmt.Default), nil
}
