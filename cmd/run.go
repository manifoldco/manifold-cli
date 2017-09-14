package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	runCmd := cli.Command{
		Name:     "run",
		Usage:    "Run a process and inject secrets into its environment",
		Category: "CONFIGURATION",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs,
			middleware.LoadTeamPrefs, run),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
		}...),
	}

	cmds = append(cmds, runCmd)
}

func run(cliCtx *cli.Context) error {
	ctx := context.Background()
	args := cliCtx.Args()

	if len(args) == 0 {
		return errs.NewUsageExitError(cliCtx, fmt.Errorf("A command is required"))
	} else if len(args) == 1 { //only one arg, maybe it was quoted
		args = strings.Split(args[0], " ")
	}

	projectLabel, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	marketplace, err := loadMarketplaceClient()
	if err != nil {
		return cli.NewExitError("Could not create marketplace client: "+err.Error(), -1)
	}

	rs, err := clients.FetchResources(ctx, marketplace, teamID, projectLabel)
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}

	cMap, err := fetchCredentials(ctx, marketplace, rs)
	if err != nil {
		return cli.NewExitError("Could not retrieve credentials: "+err.Error(), -1)
	}

	credentials, err := flattenCMap(cMap)
	if err != nil {
		return cli.NewExitError("Could not flatten credential map: "+err.Error(), -1)
	}

	a, err := loadAnalytics()
	if err != nil {
		return err
	}

	params := map[string]string{}
	if projectLabel != "" {
		params["project"] = projectLabel
	}

	a.Track(ctx, "Project Run", &params)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = filterEnv()

	for name, value := range credentials {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	err = cmd.Start()
	if err != nil {
		return cli.NewExitError("Could not execute command: "+err.Error(), -1)
	}

	done := make(chan bool)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c)

		select {
		case s := <-c:
			cmd.Process.Signal(s)
		case <-done:
			signal.Stop(c)
			return
		}
	}()

	err = cmd.Wait()
	close(done)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
				return nil
			}

			return err
		}
	}

	return nil
}

func filterEnv() []string {
	env := []string{}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, session.EnvManifoldEmail+"=") || strings.HasPrefix(e, session.EnvManifoldPass+"=") {
			continue
		}

		env = append(env, e)
	}

	return env
}
