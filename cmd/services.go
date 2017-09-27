package main

import (
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/catalog/models"
	"github.com/manifoldco/manifold-cli/prompts"

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
				Action: listProvidersCmd,
			},
			{
				Name:      "products",
				Usage:     "List all products for a provider",
				ArgsUsage: "[provider]",
				Action:    listProductsCmd,
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
