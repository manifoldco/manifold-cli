package main

import (
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/catalog/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	money "github.com/rhymond/go-money"

	"github.com/urfave/cli"
)

func init() {
	appCmd := cli.Command{
		Name:     "services",
		Usage:    "List services available on Manifold.co",
		Category: "RESOURCES",
		Subcommands: []cli.Command{
			{
				Name:   "providers",
				Usage:  "List all providers",
				Action: middleware.Chain(middleware.EnsureSession, listProvidersCmd),
			},
			{
				Name:      "products",
				Usage:     "List all products for a provider",
				ArgsUsage: "[provider]",
				Action:    middleware.Chain(middleware.EnsureSession, listProductsCmd),
			},
			{
				Name:      "plans",
				Usage:     "List all plans for a product",
				ArgsUsage: "[product]",
				Action:    middleware.Chain(middleware.EnsureSession, listPlansCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func listProvidersCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Catalog)
	if err != nil {
		return err
	}

	providers, err := client.FetchProviders()
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	if len(providers) == 0 {
		return errs.ErrNoProviders
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName\tDocumentation URL")
	w.Reset()

	for _, p := range providers {
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Body.Label, p.Body.Name, p.Body.DocumentationURL)
	}

	fmt.Fprintf(w, "\nSee provider products with `manifold services products [provider]`\n")

	return w.Flush()
}

func listProductsCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Catalog)
	if err != nil {
		return err
	}

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	label, err := optionalArgName(cliCtx, 0, "provider")
	if err != nil {
		return err
	}

	var provider *models.Provider

	if label != "" {
		provider, err = client.FetchProvider(label)
		if err != nil {
			return err
		}
	} else {
		providers, err := client.FetchProviders()
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		if len(providers) == 0 {
			return errs.ErrNoProviders
		}

		idx, _, err := prompts.SelectProvider(providers)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		provider = providers[idx]
	}

	id := provider.ID.String()
	products, err := client.FetchProducts(id)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	if len(products) == 0 {
		return errs.ErrNoProducts
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName\tState\tDesc")
	w.Reset()

	for _, p := range products {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Body.Label, p.Body.Name, *p.Body.State, p.Body.Tagline)
	}

	fmt.Fprintf(w, "\nSee product plans with `manifold services plans [product]`\n")

	return w.Flush()
}

func listPlansCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Catalog)
	if err != nil {
		return err
	}

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	label, err := optionalArgName(cliCtx, 0, "product")
	if err != nil {
		return err
	}

	var product *models.Product

	if label != "" {
		product, err = client.FetchProduct(label)
		if err != nil {
			return err
		}
	} else {
		providers, err := client.FetchProviders()
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		if len(providers) == 0 {
			return errs.ErrNoProviders
		}

		idx, _, err := prompts.SelectProvider(providers)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		provider := providers[idx]
		products, err := client.FetchProducts(provider.ID.String())
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		if len(products) == 0 {
			return errs.ErrNoProducts
		}

		idx, _, err = prompts.SelectProduct(products, "")
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		product = products[idx]
	}

	id := product.ID.String()
	plans, err := client.FetchPlans(id)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	if len(plans) == 0 {
		return errs.ErrNoPlans
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName\tState\tCost\tTrial Days")
	w.Reset()

	for _, p := range plans {
		cost := money.New(*p.Body.Cost, "USD").Display()
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", p.Body.Label, p.Body.Name,
			*p.Body.State, cost, *p.Body.TrialDays)
	}

	fmt.Fprintf(w, "\nSee plan details with `manifold services plan [plan]`\n")

	return w.Flush()
}
