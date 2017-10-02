package api

import (
	"fmt"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/session"
	"github.com/urfave/cli"
)

func (api *API) loadAnalytics(cfg *config.Config) (*analytics.Analytics, error) {
	s, err := session.Retrieve(api.ctx, cfg)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load authenticated session: %s", err), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load analytics agent: %s", err), -1)
	}

	return a, nil
}
