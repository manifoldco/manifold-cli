package main

import (
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/color"
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
				Name:   "providers",
				Usage:  "List all providers",
				Flags:  []cli.Flag{providerFlag()},
				Action: listProvidersCmd,
			},
			{
				Name:      "products",
				Usage:     "List all products. Specify a product to see its details",
				ArgsUsage: "[label]",
				Flags:     []cli.Flag{providerFlag()},
				Action:    listProductsCmd,
			},
			{
				Name:      "plans",
				Usage:     "List all plans for a product. Specify a plan to see its details",
				ArgsUsage: "[label]",
				Flags:     []cli.Flag{providerFlag(), productFlag()},
				Action:    listPlansCmd,
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func listProvidersCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Analytics, api.Catalog)
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

	var providers []*models.Provider

	if providerLabel == "" {
		providers, err = client.FetchProviders()
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		provider, err := client.FetchProvider(providerLabel)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
		providers = append(providers, provider)
	}

	if len(providers) == 0 {
		return errs.ErrNoProviders
	}

	params := map[string]string{
		"subcommand": "providers",
	}
	if providerLabel != "" {
		params["provider_label"] = providerLabel
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "Use `manifold services products --provider [label]` to view list of products\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName")
	w.Reset()

	for _, p := range providers {
		fmt.Fprintf(w, "%s\t%s\n", p.Body.Label, p.Body.Name)
	}

	return w.Flush()
}

func listProductsCmd(cliCtx *cli.Context) error {
	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	productLabel, err := optionalArgName(cliCtx, 0, "product")
	if err != nil {
		return err
	}

	if productLabel != "" {
		return viewProduct(cliCtx, productLabel)
	}

	var products []*models.Product
	var providers []*models.Provider
	var provider *models.Provider

	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerLabel, err := validateLabel(cliCtx, "provider")
	if err != nil {
		return err
	}

	if providerLabel == "" {
		providers, err = client.FetchProviders()
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
		providers = append(providers, provider)
	}

	providerID := ""

	if provider != nil {
		providerID = provider.ID.String()
	}

	if productLabel == "" {
		products, err = client.FetchProducts(providerID)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		product, err := client.FetchProduct(productLabel, providerID)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		products = append(products, product)
	}

	if len(products) == 0 {
		return errs.ErrNoProducts
	}

	params := map[string]string{
		"subcommand": "products",
	}

	if providerLabel != "" {
		params["provider_label"] = providerLabel
	}

	if productLabel != "" {
		params["product_label"] = productLabel
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "%d products from %d providers\n", len(products), len(providers))
	fmt.Fprintf(w, "Use `manifold services products [label] --provider [label]` to view product details\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName\tProvider\tTagline")
	w.Reset()

	for _, p := range products {
		provider := ""

		for _, pp := range providers {
			if pp.ID == p.Body.ProviderID {
				provider = string(pp.Body.Label)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Body.Label, p.Body.Name, provider, p.Body.Tagline)
	}

	return w.Flush()
}

func viewProduct(cliCtx *cli.Context, productLabel string) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerLabel, err := requiredLabel(cliCtx, "provider")
	if err != nil {
		return err
	}

	provider, err := client.FetchProvider(providerLabel)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	providerID := provider.ID.String()
	product, err := client.FetchProduct(productLabel, providerID)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	params := map[string]string{
		"subcommand":     "products",
		"provider_label": string(provider.Body.Label),
		"product_label":  string(product.Body.Label),
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "Use `manifold services plans --product [label] --provider [label]` to view product plans\n\n")
	fmt.Fprintln(w, "Product\tProvider")
	fmt.Fprintf(w, "%s (%s)\t%s (%s)\n\n", product.Body.Name, product.Body.Label,
		provider.Body.Name, provider.Body.Label)
	fmt.Fprintf(w, "%s\n\n", product.Body.Tagline)
	fmt.Fprintf(w, "%s\t%s\n", color.Faint("Support"), product.Body.SupportEmail)
	if product.Body.DocumentationURL != nil {
		fmt.Fprintf(w, "%s\t%s\n", color.Faint("Documentation"), *product.Body.DocumentationURL)
	}

	if product.Body.Terms.Provided {
		fmt.Fprintf(w, "%s\t%s\n", color.Faint("Terms"), *product.Body.Terms.URL)
	}

	if product.Body.Integration.Features.Sso {
		fmt.Fprintf(w, "%s\t%s\n\n", color.Faint("SSO"), "Available")
	} else {
		fmt.Fprintf(w, "%s\t%s\n\n", color.Faint("SSO"), "Unavailable")
	}

	for _, prop := range product.Body.ValueProps {
		fmt.Fprintf(w, "%s\n", color.Bold(prop.Header))
		fmt.Fprintf(w, "%s\n\n", prop.Body)
	}

	return w.Flush()
}

func listPlansCmd(cliCtx *cli.Context) error {
	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	planLabel, err := optionalArgName(cliCtx, 0, "plan")
	if err != nil {
		return err
	}

	if planLabel != "" {
		return viewPlan(cliCtx, planLabel)
	}

	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
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

	providerID := ""

	if provider != nil {
		providerID = provider.ID.String()
	}

	if productLabel != "" {
		product, err = client.FetchProduct(productLabel, providerID)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		products, err := client.FetchProducts(providerID)
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

	productID := product.ID.String()
	plans, err := client.FetchPlans(productID)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	if len(plans) == 0 {
		return errs.ErrNoPlans
	}

	params := map[string]string{
		"subcommand":     "plans",
		"provider_label": string(provider.Body.Label),
		"product_label":  string(product.Body.Label),
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintf(w, "Use `manifold services plans label --product [label] --provider [label]` to view plan details\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Label\tName\tCost")
	w.Reset()

	for _, p := range plans {
		price := *p.Body.Cost
		cost := "Free"
		if price != 0 {
			cost = money.New(price, "USD").Display()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Body.Label, p.Body.Name, cost)
	}

	return w.Flush()
}

func viewPlan(cliCtx *cli.Context, planLabel string) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
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

	provider, err := client.FetchProvider(providerLabel)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	product, err := client.FetchProduct(productLabel, provider.ID.String())
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	plan, err := client.FetchPlan(planLabel, product.ID.String())
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	regions, err := client.FetchRegions()
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	params := map[string]string{
		"subcommand":     "plans",
		"provider_label": string(provider.Body.Label),
		"product_label":  string(product.Body.Label),
		"plan_label":     string(plan.Body.Label),
	}
	client.Analytics.Track(client.Context(), "Viewed Services", &params)

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintln(w, "Plan\tProduct\tProvider")
	fmt.Fprintf(w, "%s (%s)\t%s (%s)\t%s (%s)\n\n", plan.Body.Name,
		color.Faint(plan.Body.Label), product.Body.Name, product.Body.Label,
		provider.Body.Name, provider.Body.Label)

	price := *plan.Body.Cost
	cost := "Free"
	if price != 0 {
		cost = money.New(price, "USD").Display()
	}

	fmt.Fprintf(w, "%s\t%s\n", color.Faint("Cost"), cost)
	fmt.Fprintf(w, "\n%s\n", color.Bold("Features"))
	for _, f := range plan.Body.Features {
		fmt.Fprintf(w, "%s\t%s\n", color.Faint(f.Feature), *f.Value)
	}

	fmt.Fprintf(w, "\n%s\n", color.Bold("Regions"))
	for _, r := range regions {
		for _, pr := range plan.Body.Regions {
			if r.ID == pr {
				fmt.Fprintf(w, "%s\t%s\n", color.Faint(r.Body.Name), *r.Body.Location)
			}
		}
	}

	return w.Flush()
}
