package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/session"
	"github.com/urfave/cli"
)

func loadAnalytics() (*analytics.Analytics, error) {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load authenticated session: %s", err), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load analytics agent: %s", err), -1)
	}

	return a, nil
}
