package clients

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/manifoldco/manifold-cli/config"

	bClient "github.com/manifoldco/manifold-cli/generated/billing/client"
	cClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	conClient "github.com/manifoldco/manifold-cli/generated/connector/client"
	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	aClient "github.com/manifoldco/manifold-cli/generated/activity/client"
)

const defaultUserAgent = "manifold-cli"

// EnvManifoldToken describes the environment variable name used to reference a
// Manifold api token
const EnvManifoldToken string = "MANIFOLD_API_TOKEN"

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

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
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

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return mClient.New(transport, strfmt.Default), nil
}

// NewBilling returns a swagger generated client for the Billing
// service
func NewBilling(cfg *config.Config) (*bClient.Billing, error) {
	u, err := deriveURL(cfg, "billing")
	if err != nil {
		return nil, err
	}

	c := bClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return bClient.New(transport, strfmt.Default), nil
}

// NewActivity returns a swagger generated client for Activity
func NewActivity(cfg *config.Config) (*aClient.Activity, error) {
	u, err := deriveURL(cfg, "activity")
	if err != nil {
		return nil, err
	}

	c := aClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return aClient.New(transport, strfmt.Default), nil
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

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return pClient.New(transport, strfmt.Default), nil
}

// NewConnector returns a new swagger generated client for the Connector service
func NewConnector(cfg *config.Config) (*conClient.Connector, error) {
	u, err := deriveURL(cfg, "connector")
	if err != nil {
		return nil, err
	}

	c := conClient.DefaultTransportConfig()
	c.WithHost(u.Host)
	c.WithBasePath(u.Path)
	c.WithSchemes([]string{u.Scheme})

	transport := httptransport.New(c.Host, c.BasePath, c.Schemes)
	transport.Transport = newRoundTripper(transport.Transport)

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return conClient.New(transport, strfmt.Default), nil
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

	authToken := retrieveToken(cfg)
	if authToken != "" {
		transport.DefaultAuthentication = NewBearerToken(authToken)
	}

	return cClient.New(transport, strfmt.Default), nil
}

func retrieveToken(cfg *config.Config) string {
	if cfg.AuthToken != "" {
		return cfg.AuthToken
	}
	apiToken := os.Getenv(EnvManifoldToken)
	if apiToken != "" {
		return apiToken
	}
	return ""
}

func deriveURL(cfg *config.Config, service string) (*url.URL, error) {
	u := fmt.Sprintf("%s://api.%s.%s/v1", cfg.TransportScheme, service, cfg.Hostname)
	return url.Parse(u)
}
