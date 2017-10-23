package main

import (
	"context"
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	versionCmd := cli.Command{
		Name:     "context",
		Usage:    "Display current context of this tool",
		Category: "UTILITY",
		Flags: append(teamFlags, []cli.Flag{
			cli.BoolFlag{
				Name:  "short, s",
				Usage: "Display only the label of the current context",
			},
		}...),
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, contextLookup),
	}

	cmds = append(cmds, versionCmd)
}

func contextLookup(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	var me bool
	var teamName string
	var s session.Session

	if teamID != nil {
		teamName = cfg.TeamName
		if teamName == "" {
			teamName = "unknown"
		}
	} else {
		me = true
	}
	if !cliCtx.Bool("short") || me {
		s, err = session.Retrieve(ctx, cfg)
		if err != nil {
			return err
		}
	}

	switch cliCtx.Bool("short") {
	case true:
		if me {
			fmt.Println(s.User().Body.Email)
		} else {
			fmt.Println(teamName)
		}
	default:
		usr := s.User()

		fmt.Println("Use `manifold switch` to change contexts")
		fmt.Println("")
		var ctxValue, ctxType string
		if me {
			ctxType = "user"
			ctxValue = fmt.Sprintf("%s (%s)", usr.Body.Name, usr.Body.Email)
		} else {
			ctxType = "team"
			ctxValue = fmt.Sprintf("%s (%s)", cfg.TeamName, cfg.TeamTitle)
		}

		faint := func(i interface{}) string {
			return color.Color(ansiterm.Gray, i)
		}

		fmt.Println(color.Bold("Account"))
		w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Name"), usr.Body.Name))
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Email"), usr.Body.Email))
		w.Flush()
		fmt.Println("")
		fmt.Println(color.Bold("Context"))
		w = ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Type"), ctxType))
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Value"), ctxValue))
		w.Flush()
	}
	return nil
}
