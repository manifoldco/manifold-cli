package middleware

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/urfave/cli"
	"gopkg.in/oleiade/reflections.v1"

	"github.com/manifoldco/go-manifold"
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

// LoadTeamPrefs tries to load team from config or flag. If none is present,
// sets --me to true
func LoadTeamPrefs(ctx *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	teamName := ctx.String("team")
	teamID := cfg.Team

	if teamName == "" && teamID == "" {
		ctx.Set("me", "true")
		return nil
	}

	return EnsureTeamPrefs(ctx)
}

// EnsureTeamPrefs ensures a team is set from the configuration or --team
// arguments. --me flag can be used to disable team verification.
func EnsureTeamPrefs(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var teamID string
	teamLabel := cliCtx.String("team")
	me := cliCtx.Bool("me")

	if teamLabel != "" && me {
		return cli.NewExitError("Cannot use --me with --team", -1)
	}

	if teamLabel == "" {
		teamID = cfg.Team
		// try to decode an ID, otherwise assume a label
		if _, err := manifold.DecodeIDFromString(teamID); err != nil {
			teamLabel = cfg.Team
			teamID = ""
		}
	}

	identityClient, err := clients.NewIdentity(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load identity client: %s", err), -1)
	}

	teams, err := clients.FetchTeams(ctx, identityClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load teams: %s", err), -1)
	}

	if !me && teamLabel == "" && teamID == "" {
		if len(teams) == 0 {
			return cli.NewExitError(errs.ErrNoTeams, -1)
		}

		s, err := session.Retrieve(ctx, cfg)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not retrieve session: %s", err), -1)
		}

		teamIdx, _, err := prompts.SelectTeam(teams, "", s.LabelInfo())
		if err != nil {
			prompts.HandleSelectError(err, "Could not select team")
		}

		if teamIdx == -1 {
			cliCtx.Set("me", "true")
		} else {
			teamLabel = string(teams[teamIdx].Body.Label)
			cliCtx.Set("team-id", teams[teamIdx].ID.String())
		}

	} else if teamLabel != "" {
		for _, t := range teams {
			if string(t.Body.Label) == teamLabel {
				cliCtx.Set("team-id", t.ID.String())
			}
		}

		if !isSet(cliCtx, "team-id") {
			return cli.NewExitError(fmt.Sprintf("Team \"%s\" not found", teamLabel), -1)
		}
	} else if teamID != "" {
		cliCtx.Set("team-id", teamID)
	}

	return cliCtx.Set("team", teamLabel)
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
