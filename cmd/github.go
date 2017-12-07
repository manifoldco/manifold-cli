package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/juju/ansiterm"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/session"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
)

var (
	stateChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
)

// stateLength is the length of the generate state for GitHub
const stateLength = 64
const pollingTimeout = time.Minute * 2
const pollingTick = time.Second * 5

func githubWithCallback(ctx context.Context, cfg *config.Config, a *analytics.Analytics, stateType string) error {
	// set up the oauth client
	state := genRandomString(stateLength)
	source := models.OAuthAuthenticationPollSourceGithub
	_, _, pub, err := session.NewKeyMaterial(state)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load keys from state: %s", err), -1)
	}

	identityClient, err := api.New(api.Identity)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to create identity client: %s", err), -1)
	}

	if stateType == models.OAuthAuthenticationPollTypeLink {
		// initialize the state authorization
		err := startOAuthlink(ctx, identityClient, state, source, *pub)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Unable to start OAuth link process: %s", err), -1)
		}
	}

	authConfig := &oauth2.Config{
		ClientID:    config.GitHubClientID,
		Scopes:      []string{"user"},
		Endpoint:    github.Endpoint,
		RedirectURL: fmt.Sprintf("%s?cli=true&public_key=%s&type=%s", cfg.GitHubCallback, *pub, stateType),
	}

	url := authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// let the user auth in the browser
	authMsg := color.Color(ansiterm.Cyan,
		"A browser window will open for authentication. Please close this browser when finished")
	fmt.Println(authMsg)

	time.Sleep(time.Second * 2) //
	// give them time to read the message

	err = open.Start(url)
	if err != nil {
		return cli.NewExitError("Unable to open authorization in browser", -1)
	}

	timeout := time.After(pollingTimeout)
	tick := time.Tick(pollingTick)

	op := authentication.NewPostTokensOauthPollParamsWithContext(ctx)
	op.SetBody(&models.OAuthAuthenticationPoll{
		PublicKey: pub,
		Source: &source,
		State: &state,
		Type: &stateType,
	})
	for {
		select {
		case <-timeout:
			return cli.NewExitError("Unable to fetch authentication", -1)
		case <-tick:
			loginResp, linkResp, err := identityClient.Identity.Authentication.PostTokensOauthPoll(op)
			if err != nil {
				switch err.(type) {
				case *authentication.PostTokensOauthPollNotFound:
					continue
				default:
					return err
				}
			}

			if loginResp != nil {
				cfg.AuthToken = *loginResp.Payload.Body.Token
				return cfg.Write()
			}

			if linkResp != nil {
				return nil
			}
		}
	}

	return nil
}

// startOAuthLink stores the state authorization to start the link request in the redirect
func startOAuthlink(ctx context.Context, identityClient *api.API, state, source, publicKey string) error {
	stateType := models.OAuthAuthenticationPollTypeStartLink
	op := user.NewPostUsersLinkOauthStartParamsWithContext(ctx)
	op.SetBody(&models.OAuthAuthenticationPoll{
		PublicKey: &publicKey,
		Source: &source,
		State: &state,
		Type: &stateType,
	})

	_, err := identityClient.Identity.User.PostUsersLinkOauthStart(op, nil)
	if err != nil {
		return err
	}

	return nil
}

// genRandomString generates a random string of length length, can be used for OAuth state
func genRandomString(length int) string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	size := int64(len(stateChars))

	b := make([]byte, length)
	for i := range b {
		b[i] = stateChars[seed.Int63n(size)]
	}

	return string(b)
}
