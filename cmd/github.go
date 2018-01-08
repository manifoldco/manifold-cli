package main

import (
	"context"
	"fmt"
	"crypto/rand"
	"time"
	"math/big"
	url2 "net/url"

	"github.com/juju/ansiterm"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
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
	state, err := genRandomString(stateLength)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Unable to generate state: %s", err), -1)
	}
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load keys from state: %s", err), -1)
	}

	identityClient, err := api.New(api.Identity)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to create identity client: %s", err), -1)
	}

	query := url2.Values{
		"cli": []string{"true"},
		"type": []string{stateType},
		"auth_token": []string{cfg.AuthToken},
	}

	uri := fmt.Sprintf("%s?%s", cfg.GitHubCallback, query.Encode())
	authConfig := &oauth2.Config{
		ClientID:    config.GitHubClientID,
		Scopes:      []string{"user:email"},
		Endpoint:    github.Endpoint,
		RedirectURL: uri,
	}

	url := authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// let the user auth in the browser
	authMsg := color.Color(ansiterm.Cyan,
		"A browser window will open for authentication. Please close this browser when finished")
	fmt.Println(authMsg)

	// give them time to read the message
	time.Sleep(time.Second * 2)

	err = open.Start(url)
	if err != nil {
		return cli.NewExitError("Unable to open authorization in browser", -1)
	}

	timeout := time.After(pollingTimeout)
	tick := time.Tick(pollingTick)

	op := authentication.NewGetTokensOauthPollParamsWithContext(ctx)
	op.Secret = state
	for {
		select {
		case <-timeout:
			return cli.NewExitError("Unable to fetch authentication", -1)
		case <-tick:
			loginResp, linkResp, err := identityClient.Identity.Authentication.GetTokensOauthPoll(op)
			if err != nil {
				switch err.(type) {
				case *authentication.GetTokensOauthPollNotFound:
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

// genRandomString generates a random string of length length, can be used for OAuth state
func genRandomString(length int) (string, error) {
	size := big.NewInt(int64(len(stateChars)))

	b := make([]byte, length)
	for i := range b {
		l, err := rand.Int(rand.Reader, size)
		if err != nil {
			return "", err
		}
		b[i] = stateChars[l.Int64()]
	}

	return string(b), nil
}
