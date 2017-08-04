package session

import (
	"context"
	"os"

	"github.com/go-openapi/strfmt"
	"github.com/manifoldco/go-base64"
	"github.com/manifoldco/go-manifold"
	"github.com/reconquest/hierr-go"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"

	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
)

// EnvManifoldEmail describes the environment variable name used to reference a
// Manifold login email
const EnvManifoldEmail string = "MANIFOLD_EMAIL"

// EnvManifoldPass describes the environment variable name used to reference a
// Manifold login password
const EnvManifoldPass string = "MANIFOLD_PASS"

// Session interface to describe user session and authentication with Manifold
//  API
type Session interface {
	Authenticated() bool
	FromEnvVars() bool
	User() *models.User
}

// Unauthorized struct to represent an unauthorized user session
type Unauthorized struct{}

// Authenticated returns if the session is authenticated or not, in this case
// false
func (*Unauthorized) Authenticated() bool { return false }

// FromEnvVars returns if the session was recently authenticated from
// environment login variablesor not, in this case false
func (*Unauthorized) FromEnvVars() bool { return false }

// User returns the user object associated with this session, in this case nil
func (*Unauthorized) User() *models.User { return nil }

// Authorized struct to represent an authorized user session
type Authorized struct {
	user        *models.User
	fromEnvVars bool
}

// Authenticated returns if the session is authenticated or not, in this case
//  true
func (a *Authorized) Authenticated() bool { return true }

// User returns the user object associated with this session, in this case nil
func (a *Authorized) User() *models.User { return a.user }

// FromEnvVars returns if the session was recently authenticated from
// environment login variablesor not
func (a *Authorized) FromEnvVars() bool { return a.fromEnvVars }

/**
 * Public
 */

// Retrieve a session struct from the Manifold API based on the auth token in
//  the config
func Retrieve(ctx context.Context, cfg *config.Config) (Session, error) {
	sess, err := retrieveSession(ctx, cfg, false)
	if err != nil {
		return nil, err
	} else if !sess.Authenticated() {
		// Attempt to login via environment variables if present
		return loginFromEnv(ctx, cfg)
	}
	return sess, nil
}

func retrieveSession(ctx context.Context, cfg *config.Config,
	fromEnvVars bool) (Session, error) {
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
		switch err.(type) {
		case *user.GetSelfUnauthorized:
			// Stored token is not valid

			// Clear stored token
			cfg.AuthToken = ""
			err = cfg.Write()
			if err != nil {
				return nil, hierr.Errorf(err,
					"Failed to update config after clearing expired auth token.")
			}

			return &Unauthorized{}, nil
		default:
			return nil, err
		}
	}

	return &Authorized{user: r.Payload, fromEnvVars: fromEnvVars}, nil
}

func loginFromEnv(ctx context.Context, cfg *config.Config) (Session, error) {
	envUser := os.Getenv(EnvManifoldEmail)
	envPass := os.Getenv(EnvManifoldPass)

	if envUser == "" || envPass == "" {
		return &Unauthorized{}, nil
	}

	sess, err := createSession(ctx, cfg, envUser, envPass, true)
	if err != nil {
		return nil, hierr.Errorf(err, "Attempted to login with the %s and %s "+
			"environment variables and failed", EnvManifoldEmail, EnvManifoldPass)
	}
	return sess, nil
}

// Create a new session with the Manifold API based on the provided credentials
func Create(ctx context.Context, cfg *config.Config, email,
	password string) (Session, error) {
	return createSession(ctx, cfg, email, password, false)
}

func createSession(ctx context.Context, cfg *config.Config, email,
	password string, fromEnvVars bool) (Session, error) {
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
			return nil, errs.ErrSomethingWentHorriblyWrong
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
			return nil, errs.ErrSomethingWentHorriblyWrong
		}
	}

	cfg.AuthToken = *authResult.Payload.Body.Token
	err = cfg.Write()
	if err != nil {
		return nil, err
	}

	return retrieveSession(ctx, cfg, fromEnvVars)
}

// Destroy the session by invalidating the token through the Manifold API and
// clearing the local auth token cache
func Destroy(ctx context.Context, cfg *config.Config) error {
	c, err := clients.NewIdentity(cfg)
	if err != nil {
		return err
	}

	p := authentication.NewDeleteTokensTokenParamsWithContext(ctx)
	p.SetToken(cfg.AuthToken)
	_, err = c.Authentication.DeleteTokensToken(p, nil)
	if err != nil {
		switch err.(type) {
		case *authentication.DeleteTokensTokenUnauthorized:
			// Handle gracefully session already expired... this should have been
			// caught on retrieve though
		default:
			return err
		}
	}

	// Dispose of local auth token cache
	cfg.AuthToken = ""
	err = cfg.Write()
	if err != nil {
		return hierr.Errorf(err,
			"Failed to update config to clear auth token from logout.")
	}

	return nil
}

// Signup makes a request to create a new account
func Signup(ctx context.Context, cfg *config.Config, name, email, password string) (*models.User, error) {
	c, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, err
	}

	alg, salt, pubkey, err := newKeyMaterial(password)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to derive publickey")
	}

	p := user.NewPostUsersParamsWithContext(ctx)
	p.Body = &models.CreateUser{
		Body: &models.CreateUserBody{
			Email: manifold.Email(email),
			Name:  models.UserDisplayName(name),
			PublicKey: &models.LoginPublicKey{
				Alg:   alg,
				Salt:  salt,
				Value: pubkey,
			},
		},
	}
	resp, err := c.User.PostUsers(p)

	if err != nil {
		switch e := err.(type) {
		case *user.PostUsersBadRequest:
			return nil, e.Payload
		default:
			return nil, errs.ErrSomethingWentHorriblyWrong
		}
	}

	return resp.Payload, nil
}
