package clients

import (
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/generated/identity/client"
)

// NewIdentity returns a swagger generated client for the Identity service
func NewIdentity(cfg *config.Config) (*client.Identity, error) {
	identityURL := cfg.TransportScheme + "://api.identity." + cfg.Hostname + "/v1"
	u, err := url.Parse(identityURL)
	if err != nil {
		return nil, err
	}

	c := client.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.DefaultAuthentication = httptransport.BearerToken(cfg.AuthToken)

	return client.New(transport, strfmt.Default), nil
}
