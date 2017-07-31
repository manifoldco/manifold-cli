package clients

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/manifoldco/manifold-cli/config"

	cClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
)

const defaultUserAgent = "manifold-cli"

// newRoundTripper applies a UserAgent header to the transport
func newRoundTripper(next http.RoundTripper) http.RoundTripper {
	version := config.Version
	if version != "" {
		version = "-" + version
	}
	return &roundTripper{
		next:      next,
		userAgent: defaultUserAgent + version,
	}
}

type roundTripper struct {
	next      http.RoundTripper
	userAgent string
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", rt.userAgent)
	return rt.next.RoundTrip(req)
}

// NewIdentity returns a swagger generated client for the Identity service
func NewIdentity(cfg *config.Config) (*iClient.Identity, error) {
	u, err := deriveURL(cfg, "identity")
	if err != nil {
		return nil, err
	}

	c := iClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.Transport = newRoundTripper(transport.Transport)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return iClient.New(transport, strfmt.Default), nil
}

// NewMarketplace returns a swagger generated client for the Marketplace service
func NewMarketplace(cfg *config.Config) (*mClient.Marketplace, error) {
	u, err := deriveURL(cfg, "marketplace")
	if err != nil {
		return nil, err
	}

	c := mClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.Transport = newRoundTripper(transport.Transport)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return mClient.New(transport, strfmt.Default), nil
}

// NewProvisioning returns a swagger generated client for the Provisioning
// service
func NewProvisioning(cfg *config.Config) (*pClient.Provisioning, error) {
	u, err := deriveURL(cfg, "provisioning")
	if err != nil {
		return nil, err
	}

	c := pClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.Transport = newRoundTripper(transport.Transport)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return pClient.New(transport, strfmt.Default), nil
}

// NewBearerToken returns a bearer token authenticator for use with a
// go-swagger generated client.
func NewBearerToken(token string) runtime.ClientAuthInfoWriter {
	return httptransport.BearerToken(token)
}

// NewCatalog returns a swagger generated client for the Catalog service
func NewCatalog(cfg *config.Config) (*cClient.Catalog, error) {
	u, err := deriveURL(cfg, "catalog")
	if err != nil {
		return nil, err
	}

	c := cClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.Transport = newRoundTripper(transport.Transport)

	if cfg.AuthToken != "" {
		transport.DefaultAuthentication = NewBearerToken(cfg.AuthToken)
	}

	return cClient.New(transport, strfmt.Default), nil
}

func deriveURL(cfg *config.Config, service string) (*url.URL, error) {
	u := fmt.Sprintf("%s://api.%s.%s/v1", cfg.TransportScheme, service, cfg.Hostname)
	return url.Parse(u)
}
