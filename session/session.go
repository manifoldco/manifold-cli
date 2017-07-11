package session

import (
	"context"
	"fmt"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/manifoldco/manifold-cli/config"

	"github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
)

// Session interface to describe user session and authentication with Manifold
//  API
type Session interface {
	Authenticated() bool
	User() *models.User
}

/**
 * TYPES
 */

// Unauthorized struct to represent an unauthorized user session
type Unauthorized struct{}

// Authenticated returns if the session is authenticated or not, in this case
//  false
func (u *Unauthorized) Authenticated() bool { return false }

// User returns the user object associated with this session, in this case nil
func (u *Unauthorized) User() *models.User { return nil }

// Authorized struct to represent an authorized user session
type Authorized struct {
	userObj *models.User
}

// Authenticated returns if the session is authenticated or not, in this case
//  true
func (a *Authorized) Authenticated() bool { return true }

// User returns the user object associated with this session, in this case nil
func (a *Authorized) User() *models.User { return a.userObj }

/**
 * PRIVATE
 */

func newIdentityClient(cfg *config.Config) (*client.Identity, error) {
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
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Bearer "+cfg.AuthToken)

	return client.New(transport, strfmt.Default), nil
}

/**
 * Public
 */

// Retrieve a session struct from the Manifold API based on the auth token in
//  the config
func Retrieve(ctx context.Context, cfg *config.Config) (Session, error) {
	fmt.Printf("hello%s?\n", cfg.AuthToken)
	if cfg.AuthToken == "" {
		return &Unauthorized{}, nil
	}

	c, err := newIdentityClient(cfg)
	if err != nil {
		return nil, err
	}

	p := user.NewGetSelfParamsWithContext(ctx)
	userResult, err := c.User.GetSelf(p, nil)
	if err != nil {
		switch e := err.(type) {
		case *user.GetSelfUnauthorized:
			fmt.Printf("Unauthorized yo: " + err.Error() + "\n")
			return &Unauthorized{}, nil
		default:
			return nil, e
		}
	}

	return &Authorized{userObj: userResult.Payload}, nil
}

// Create a new session with the Manifold API based on the provided credentials
func Create(ctx context.Context, cfg *config.Config, username,
	password string) (Session, error) {
	return nil, nil
}
