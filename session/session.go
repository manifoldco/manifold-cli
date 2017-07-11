package session

import (
	"context"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"

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
func (*Unauthorized) Authenticated() bool { return false }

// User returns the user object associated with this session, in this case nil
func (*Unauthorized) User() *models.User { return nil }

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
 * Public
 */

// Retrieve a session struct from the Manifold API based on the auth token in
//  the config
func Retrieve(ctx context.Context, cfg *config.Config) (Session, error) {
	if cfg.AuthToken == "" {
		return &Unauthorized{}, nil
	}

	c, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, err
	}

	p := user.NewGetSelfParamsWithContext(ctx)
	userResult, err := c.User.GetSelf(p, nil)
	if err != nil {
		switch e := err.(type) {
		case *user.GetSelfUnauthorized:
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
