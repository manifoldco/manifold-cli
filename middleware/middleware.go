package middleware

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/urfave/cli"
	"gopkg.in/oleiade/reflections.v1"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

// Chain allows easy sequential calling of BeforeFuncs and AfterFuncs.
// chain will exit on the first error seen.
func Chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {

		for _, f := range funcs {
			err := f(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// LoadDirPrefs loads argument values from the .torus.json file
func LoadDirPrefs(ctx *cli.Context) error {
	d, err := config.LoadYaml(true)
	if err != nil {
		return err
	}

	return reflectArgs(ctx, d, "flag")
}

// LoadTeamPrefs loads the team from the configuration, and then --team arguments
func LoadTeamPrefs(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	teamName := cliCtx.String("team")
	me := cliCtx.Bool("me")

	if teamName != "" && me {
		return cli.NewExitError("Cannot use --me with --team", -1)
	}

	if teamName == "" {
		teamName = cfg.Team
	}

	if !me && teamName == "" {
		identityClient, err := clients.NewIdentity(cfg)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not load identity client: %s", err), -1)
		}

		teams, err := clients.FetchTeams(ctx, identityClient)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not load teams: %s", err), -1)
		}
		if len(teams) == 0 {
			return cli.NewExitError(errs.ErrNoTeams, -1)
		}

		teamIdx, _, err := prompts.SelectTeam(teams, "", true)
		if err != nil {
			prompts.HandleSelectError(err, "Could not select team")
		}

		if teamIdx == -1 {
			cliCtx.Set("me", "true")
		} else {
			teamName = string(teams[teamIdx].Body.Label)
		}
	}

	return cliCtx.Set("team", teamName)
}

// EnsureSession checks that the user has an active session
func EnsureSession(_ *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not retrieve session: %s", err), -1)
	}

	if !s.Authenticated() {
		return errs.ErrNotLoggedIn
	}

	return nil
}

func reflectArgs(ctx *cli.Context, i interface{}, tagName string) error {
	// tagged field names match the argument names
	tags, err := reflections.Tags(i, tagName)
	if err != nil {
		return err
	}

	flags := make(map[string]bool)
	for _, flagName := range ctx.FlagNames() {
		// This value is already set via arguments or env vars. skip it.
		if isSet(ctx, flagName) {
			continue
		}

		flags[flagName] = true
	}

	for fieldName, tag := range tags {
		name := strings.SplitN(tag, ",", 2)[0] // remove omitempty if its there
		if _, ok := flags[name]; ok {
			field, err := reflections.GetField(i, fieldName)
			if err != nil {
				return err
			}

			if f, ok := field.(string); ok && f != "" {
				ctx.Set(name, field.(string))
			}
		}
	}

	return nil
}

func isSet(ctx *cli.Context, name string) bool {
	value := ctx.Generic(name)
	if value != nil {
		v := reflect.Indirect(reflect.ValueOf(value))
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			return v.Len() != 0
		}

		return true
	}

	return false
}
