package main

import (
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/catalog/models"
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
				Name:      "providers",
				Usage:     "List all providers",
				ArgsUsage: "[provider label]",
				Action:    listProvidersCmd,
			},
			{
				Name:      "products",
				Usage:     "List all products for a provider",
				ArgsUsage: "[product label]",
				Flags:     []cli.Flag{providerFlag()},
				Action:    listProductsCmd,
			},
			{
				Name:   "plans",
				Usage:  "List all plans for a product",
				Flags:  []cli.Flag{providerFlag(), productFlag()},
				Action: listPlansCmd,
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

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	label, err := optionalArgName(cliCtx, 0, "provider")
	if err != nil {
		return err
	}

	var providers []*models.Provider

	if label == "" {
		providers, err = client.FetchProviders()
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		provider, err := client.FetchProvider(label)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
		providers = append(providers, provider)
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

	fmt.Fprintf(w, "\nSee provider products with `manifold services products [--provider label]`\n")

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

	label, err := optionalArgName(cliCtx, 0, "product")
	if err != nil {
		return err
	}

	providerLabel, err := validateLabel(cliCtx, "provider")
	if err != nil {
		return err
	}

	var products []*models.Product

	if label == "" {
		var provider *models.Provider

		if providerLabel == "" {
			providers, err := client.FetchProviders()
			if err != nil {
				return cli.NewExitError(err, -1)
			}

			if len(providers) == 0 {
				return errs.ErrNoProviders
			}

			provider, err = prompts.SelectProvider(providers)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
		} else {
			provider, err = client.FetchProvider(providerLabel)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
		}

		id := ""
		if provider != nil {
			id = provider.ID.String()
		}
		products, err = client.FetchProducts(id)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		product, err := client.FetchProduct(label)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		products = append(products, product)
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

	fmt.Fprintf(w, "\nSee product plans with `manifold services plans [--product label]`\n")

	return w.Flush()
}

func listPlansCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Catalog)
	if err != nil {
		return err
	}

	if err := maxOptionalArgsLength(cliCtx, 0); err != nil {
		return err
	}

	providerLabel, err := validateLabel(cliCtx, "provider")
	if err != nil {
		return err
	}

	productLabel, err := validateLabel(cliCtx, "product")
	if err != nil {
		return err
	}

	var product *models.Product

	if productLabel != "" {
		product, err = client.FetchProduct(productLabel)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		var provider *models.Provider

		if providerLabel != "" {
			provider, err = client.FetchProvider(providerLabel)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
		} else {
			providers, err := client.FetchProviders()
			if err != nil {
				return cli.NewExitError(err, -1)
			}

			if len(providers) == 0 {
				return errs.ErrNoProviders
			}

			provider, err = prompts.SelectProvider(providers)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
		}

		id := ""
		if provider != nil {
			id = provider.ID.String()
		}

		products, err := client.FetchProducts(id)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
		if len(products) == 0 {
			return errs.ErrNoProducts
		}

		idx, _, err := prompts.SelectProduct(products, "")
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
		price := *p.Body.Cost
		cost := "Free"
		if price != 0 {
			cost = money.New(price, "USD").Display()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", p.Body.Label, p.Body.Name,
			*p.Body.State, cost, *p.Body.TrialDays)
	}

	return w.Flush()
}
