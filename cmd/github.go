package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/juju/ansiterm"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/prompts"
)

var (
	sourceGitHub      = "github"
	sourceGitHubToken = "github-token"
	stateChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
)

// githubAuthorizationRequest is the request sent to GitHub for Personal Authorization Tokens
type githubAuthorizationRequest struct {
	Scopes      []string `json:"scopes"`
	Notes       string   `json:"note"`
	Fingerprint string   `json:"fingerprint"`
}

// githubAuthorizationResponse is the response that is given for Personal Authorization Token
// requests
type githubAuthorizationResponse struct {
	ID    int
	Token string

	Errors []struct {
		Code string
	}
}

// githubErrors are the human readable error messages from GitHub during the Personal Authorization
// Token flow
var githubErrors = map[string]string{
	"already_exists": "A Personal Authorization Token already exists for Manifold",
}

// oauthStoreFunc is responsible for storing the OAuth authentication with Manifold, either by
// linking accounts or logging in
type oauthStoreFunc func(ctx context.Context, cfg *config.Config, a *analytics.Analytics, req models.OAuthAuthenticationRequest) error

// githubWithCallback identified with GitHub via. the normal browser-based OAuth flow
func githubWithCallback(ctx context.Context, cfg *config.Config, a *analytics.Analytics, store oauthStoreFunc) error {
	authConfig := &oauth2.Config{
		ClientID:    config.GitHubClientID,
		Scopes:      []string{"user"},
		Endpoint:    github.Endpoint,
		RedirectURL: "http://127.0.0.1:49152/github/callback",
	}

	// set up the oauth client
	state := genRandomString(12)
	url := authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// let the user auth in the browser
	authMsg := color.Color(ansiterm.Blue,
		"A browser window will open for authentication. Please close this browser when finished")
	fmt.Println(authMsg)

	time.Sleep(time.Second * 2) // give them time to read the message

	err := open.Start(url)
	if err != nil {
		return cli.NewExitError("Unable to open authorization in browser", -1)
	}

	// handle http callback
	done := make(chan bool)
	errCh := make(chan error)
	go func() {
		http.HandleFunc("/github/callback", githubCallback(ctx, cfg, state, store, a, done, errCh))

		errCh <- http.ListenAndServe(":49152", nil)
	}()

	select {
	case <-done:
		authedMsg := color.Color(ansiterm.Green, "You are now authenticated with GitHub, and can close your browser window")
		fmt.Printf("\n%s\n", authedMsg)
	case err := <-errCh:
		errMsg := color.Color(ansiterm.Red, fmt.Sprintf("Error with authentication: %s", err))
		return errors.New(errMsg)
	}

	return nil
}

// githubCallback is the OAuth callback function for the browser-based flow
func githubCallback(ctx context.Context, cfg *config.Config, state string, store oauthStoreFunc,
	a *analytics.Analytics, done chan bool, errCh chan error) func(w http.ResponseWriter, r *http.Request) {

	badRequest := func(w http.ResponseWriter, msg string, vals ...interface{}) {
		valMsg := fmt.Sprintf(msg, vals...)
		fmtMsg := fmt.Sprintf("<html><body><h2>Bad request</h2><p>%s</p></body>", valMsg)

		w.WriteHeader(500)
		w.Header().Set("Content-Length", strconv.Itoa(len(fmtMsg)))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		w.Write([]byte(fmtMsg))

		errCh <- errors.New(valMsg)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		givenState := req.URL.Query().Get("state")

		if state != givenState {
			badRequest(w, "Given state did not match provided, please try again")
			return
		}
		authReq := models.OAuthAuthenticationRequest{
			Code:   &code,
			Source: &sourceGitHub,
		}

		err := store(ctx, cfg, a, authReq)
		if err != nil {
			badRequest(w, "Unable to authenticate with GitHub: %s", err)
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, "🎉 You are logged in! You can now close this window")

		done <- true
	}
}

// githubWithToken is a method to log into GitHub via. a token to pass authentication to Manifold
func githubWithToken(ctx context.Context, cfg *config.Config, a *analytics.Analytics, token string, store oauthStoreFunc) error {
	authReq := models.OAuthAuthenticationRequest{
		Code:   &token,
		Source: &sourceGitHubToken,
	}

	err := store(ctx, cfg, a, authReq)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to save GitHub authentication: %s", err), -1)
	}

	err = cfg.Write()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to write config: %s", err), -1)
	}

	a.Track(ctx, "Logged In", nil)

	return nil
}

// githubUser is a method for authenticating with GitHub for Manifold using a user's username,
// password, and OTP for 2FA
func githubWithUser(ctx context.Context, cfg *config.Config, a *analytics.Analytics, store oauthStoreFunc) error {
	username, err := prompts.Username()
	if err != nil {
		return err
	}

	password, err := prompts.Hidden("Password")
	if err != nil {
		return err
	}

	resp, err := githubAuthRequest(ctx, username, password, "", "")
	if resp == nil && err != nil {
		return err
	}

	otpReq := resp.Header.Get("X-GitHub-OTP")
	if strings.Contains(otpReq, "required;") {
		otp, err := prompts.OTP()
		if err != nil {
			return err
		}

		resp.Body.Close()
		resp, err = githubAuthRequest(ctx, username, password, otp, "")
		if err != nil {
			return err
		}
	}

	ghResp, err := decodeGitHubResponse(ctx, resp)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to decode Personal Access Token: %s", err), -1)
	}

	authReq := models.OAuthAuthenticationRequest{
		Code:   &ghResp.Token,
		Source: &sourceGitHubToken,
	}

	err = store(ctx, cfg, a, authReq)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to fetch and store GitHub authentication: %s", err), -1)
	}

	return nil
}

// githubAuthRequest authenticates with GitHub by creating a new Personal Access Token
func githubAuthRequest(ctx context.Context, username, password, otp, token string) (*http.Response, error) {

	id := genRandomString(5)
	client := &http.Client{}
	authReq := githubAuthorizationRequest{
		Scopes:      []string{"user"},
		Notes:       fmt.Sprintf("Manifold-%s", id),
		Fingerprint: fmt.Sprintf("manifold-cli-%s", id),
	}

	auth := &bytes.Buffer{}
	err := json.NewEncoder(auth).Encode(authReq)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Unable to encode authentication request: %s", err), -1)
	}

	authUrl := fmt.Sprintf("%s/authorizations", config.GitHubHost)
	req, err := http.NewRequest(http.MethodPost, authUrl, auth)
	if err != nil {
		return nil, cli.NewExitError("Unable to create auth request for GitHub", -1)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if username != "" && password != "" {
		// auth for the API
		req.SetBasicAuth(username, password)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	// set the otp if one is provided
	if otp != "" {
		req.Header.Set("X-GitHub-OTP", otp)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Unable to send authentication request: %s", err), -1)
	}

	if resp.StatusCode != http.StatusCreated {
		_, err := decodeGitHubResponse(ctx, resp)
		if err != nil {
			return nil, cli.NewExitError(fmt.Sprintf("Bad login details: %s", err), -1)
		}
		return resp, cli.NewExitError("Authentication code not created, bad login", -1)
	}

	return resp, nil
}

// decodeGitHubResponse decodes the response from GitHub for an authentication request
func decodeGitHubResponse(ctx context.Context, resp *http.Response) (*githubAuthorizationResponse, error) {
	var authResp *githubAuthorizationResponse

	b := []byte{}
	resp.Body.Read(b)

	err := json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, err
	}

	if len(authResp.Errors) > 0 {
		// fail with the first error
		return nil, errors.New(githubErrors[authResp.Errors[0].Code])
	}

	return authResp, nil
}

// createOAuthAuth registers a new user (who is not part of Manifold), or logs a user in (with a
// linked GitHub identity)
func createOAuthAuth(ctx context.Context, cfg *config.Config, a *analytics.Analytics, authReq models.OAuthAuthenticationRequest) error {
	apiClient, err := api.New(api.Identity)
	if err != nil {
		return err
	}

	op := authentication.NewPostTokensOauthParamsWithContext(ctx)
	op.SetBody(&authReq)

	resp, err := apiClient.Identity.Authentication.PostTokensOauth(op)
	if err != nil {
		return err
	}

	cfg.AuthToken = *resp.Payload.Body.Token
	a.Track(ctx, "Logged In", nil)
	return cfg.Write()
}

// linkOAuthAuth links an existing user account (defined by the session) to a GitHub identity
func linkOAuthAuth(ctx context.Context, cfg *config.Config, a *analytics.Analytics, linkReq models.OAuthAuthenticationRequest) error {
	apiClient, err := api.New(api.Identity)
	if err != nil {
		return err
	}

	op := user.NewPostUsersLinkOauthParamsWithContext(ctx)
	op.SetBody(&linkReq)

	_, err = apiClient.Identity.User.PostUsersLinkOauth(op, nil)
	if err != nil {
		return err
	}

	a.Track(ctx, "Link User", nil)
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
