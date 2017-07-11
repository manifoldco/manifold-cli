package session

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"

	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
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
	user *models.User
}

// Authenticated returns if the session is authenticated or not, in this case
//  true
func (a *Authorized) Authenticated() bool { return true }

// User returns the user object associated with this session, in this case nil
func (a *Authorized) User() *models.User { return a.user }

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
	r, err := c.User.GetSelf(p, nil)
	if err != nil {
		switch e := err.(type) {
		case *user.GetSelfUnauthorized:
			return &Unauthorized{}, nil
		case *user.GetSelfInternalServerError:
			return nil, e
		default:
			return nil, e
		}
	}

	return &Authorized{user: r.Payload}, nil
}

// Create a new session with the Manifold API based on the provided credentials
func Create(ctx context.Context, cfg *config.Config, email, password string) (Session, error) {
	// Erase the auth token, if it's set
	// XXX: Come back and attempt a logout
	if cfg.AuthToken != "" {
		cfg.AuthToken = ""
		err := cfg.Write()
		if err != nil {
			return nil, err
		}
	}

	c, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, err
	}

	// Get a login token, to sign, and to trade in for an auth token
	strfmtEmail := strfmt.Email(email)
	p := authentication.NewPostTokensLoginParamsWithContext(ctx).WithBody(
		&models.LoginTokenRequest{
			Email: &strfmtEmail,
		},
	)
	result, err := c.Authentication.PostTokensLogin(p)
	if err != nil {
		switch e := err.(type) {
		case *authentication.PostTokensLoginBadRequest:
			return nil, e.Payload
		case *authentication.PostTokensLoginInternalServerError:
			return nil, e.Payload
		default:
			return nil, err
		}
	}

	salt, err := base64.NewFromString(*result.Payload.Salt)
	if err != nil {
		return nil, err
	}

	_, privkey, err := deriveKeypair(password, salt)
	if err != nil {
		return nil, err
	}

	// Using the login token for auth and the signed token, retrieve an auth
	// token
	tokenType := "auth"
	signedToken := sign(privkey, *result.Payload.Token).String()
	authP := authentication.NewPostTokensAuthParamsWithContext(ctx).WithBody(
		&models.AuthTokenRequest{
			LoginTokenSig: &signedToken,
			Type:          &tokenType,
		},
	).WithAuthorization("Bearer " + *result.Payload.Token)

	authResult, err := c.Authentication.PostTokensAuth(authP, nil)
	if err != nil {
		switch e := err.(type) {
		case *authentication.PostTokensAuthBadRequest:
			return nil, e.Payload
		case *authentication.PostTokensAuthUnauthorized:
			return nil, e.Payload
		case *authentication.PostTokensAuthInternalServerError:
			return nil, e.Payload
		}
	}

	cfg.AuthToken = *authResult.Payload.Body.Token
	err = cfg.Write()
	if err != nil {
		return nil, err
	}

	return Retrieve(ctx, cfg)
}
