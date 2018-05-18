package main

import (
	"fmt"
	"os"
	"strings"

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
				Name:   "categories",
				Usage:  "List all products by category. Specify a category to see its details",
				Action: listCategoriesCmd,
			},
			{
				Name:      "providers",
				Usage:     "List all providers",
				ArgsUsage: "[provider-name]",
				Action:    listProvidersCmd,
			},
			{
				Name:      "products",
				Usage:     "List all products. Specify a product to see its details",
				ArgsUsage: "[product-name]",
				Flags:     []cli.Flag{providerFlag()},
				Action:    listProductsCmd,
			},
			{
				Name:      "plans",
				Usage:     "List all plans for a product. Specify a plan to see its details",
				ArgsUsage: "[plan-name]",
				Flags:     []cli.Flag{providerFlag(), productFlag()},
				Action:    listPlansCmd,
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func listCategoriesCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	if err := maxOptionalArgsLength(cliCtx, 0); err != nil {
		return err
	}

	var products []*models.Product
	var providers []*models.Provider
	var categories map[string][]*models.Product
	var providerNames map[string]*models.Provider

	categories = make(map[string][]*models.Product)
	providerNames = make(map[string]*models.Provider)

	products, err = client.FetchProducts("")
	if err != nil {
		return cli.NewExitError(err, -1)
	}
	providers, err = client.FetchProviders()
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	for _, p := range products {
		for _, category := range p.Body.Tags {
			var categoryTitled = strings.Title(category)
			categories[categoryTitled] = append(categories[categoryTitled], p)
		}
	}

	for _, p := range providers {
		providerNames[p.ID.String()] = p
	}

	categoryName, err := prompts.SelectCategory(categories)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "%d products in %s\n", len(categories[categoryName]), categoryName)
	fmt.Fprintf(w, "Use `manifold services categories` to view products with a specified category\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Name\tTitle\tProvider\tTagline")
	w.Reset()

	for _, p := range categories[categoryName] {
		providerID := p.Body.ProviderID.String()
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Body.Label, p.Body.Name, providerNames[providerID].Body.Name, p.Body.Tagline)
	}

	return w.Flush()
}

func listProvidersCmd(cliCtx *cli.Context) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	providerName, err := optionalArgName(cliCtx, 0, "provider")
	if err != nil {
		return err
	}

	var providers []*models.Provider

	if providerName == "" {
		providers, err = client.FetchProviders()
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		provider, err := client.FetchProvider(providerName)
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
	if providerName != "" {
		params["provider_label"] = providerName
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "Use `manifold services products --provider [provider-name]` to view list of products\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Name\tTitle")
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

	productName, err := optionalArgName(cliCtx, 0, "product")
	if err != nil {
		return err
	}

	if productName != "" {
		return viewProduct(cliCtx, productName)
	}

	var products []*models.Product
	var providers []*models.Provider
	var provider *models.Provider

	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerName, err := validateName(cliCtx, "provider")
	if err != nil {
		return err
	}

	if providerName == "" {
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
		provider, err = client.FetchProvider(providerName)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
		providers = append(providers, provider)
	}

	providerID := ""

	if provider != nil {
		providerID = provider.ID.String()
	}

	if productName == "" {
		products, err = client.FetchProducts(providerID)
		if err != nil {
			return cli.NewExitError(err, -1)
		}
	} else {
		product, err := client.FetchProduct(productName, providerID)
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

	if providerName != "" {
		params["provider_label"] = providerName
	}

	if productName != "" {
		params["product_label"] = productName
	}

	client.Analytics.Track(client.Context(), "Viewed Services", &params)

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)

	fmt.Fprintf(w, "%d products from %d providers\n", len(products), len(providers))
	fmt.Fprintf(w, "Use `manifold services products [product-name] --provider [provider-name]` to view product details\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Name\tTitle\tProvider\tTagline")
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

func viewProduct(cliCtx *cli.Context, productName string) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerName, err := requiredName(cliCtx, "provider")
	if err != nil {
		return err
	}

	provider, err := client.FetchProvider(providerName)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	providerID := provider.ID.String()
	product, err := client.FetchProduct(productName, providerID)
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

	fmt.Fprintf(w, "Use `manifold services plans --product [product-name] --provider [provider-name]` to view product plans\n\n")
	fmt.Fprintln(w, "Product\tProvider")
	fmt.Fprintf(w, "%s (%s)\t%s (%s)\n\n", product.Body.Label, color.Faint(product.Body.Name),
		provider.Body.Label, color.Faint(provider.Body.Name))
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

	planName, err := optionalArgName(cliCtx, 0, "plan")
	if err != nil {
		return err
	}

	if planName != "" {
		return viewPlan(cliCtx, planName)
	}

	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerName, err := validateName(cliCtx, "provider")
	if err != nil {
		return err
	}

	productName, err := validateName(cliCtx, "product")
	if err != nil {
		return err
	}

	var product *models.Product
	var provider *models.Provider

	if providerName != "" {
		provider, err = client.FetchProvider(providerName)
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

	if productName != "" {
		product, err = client.FetchProduct(productName, providerID)
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
	fmt.Fprintf(w, "Use `manifold services plans [plan-name] --product [product-name] --provider [provider-name]` to view plan details\n\n")

	w.SetForeground(ansiterm.Gray)
	fmt.Fprintln(w, "Name\tTitle\tCost")
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

func viewPlan(cliCtx *cli.Context, planName string) error {
	client, err := api.New(api.Analytics, api.Catalog)
	if err != nil {
		return err
	}

	providerName, err := validateName(cliCtx, "provider")
	if err != nil {
		return err
	}

	productName, err := validateName(cliCtx, "product")
	if err != nil {
		return err
	}

	provider, err := client.FetchProvider(providerName)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	product, err := client.FetchProduct(productName, provider.ID.String())
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	plan, err := client.FetchPlan(planName, product.ID.String())
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
	fmt.Fprintf(w, "%s (%s)\t%s (%s)\t%s (%s)\n\n", plan.Body.Label,
		color.Faint(plan.Body.Name), product.Body.Label, color.Faint(product.Body.Name),
		provider.Body.Label, color.Faint(provider.Body.Name))

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
