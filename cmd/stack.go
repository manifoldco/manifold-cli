package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/cmd/stack"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

func init() {
	stackCmd := cli.Command{
		Name:     "stack",
		Usage:    "Organize project resources into stack.yml",
		Category: "RESOURCES",
		Subcommands: []cli.Command{
			{
				Name:  "init",
				Usage: "Initialize a new stack.yml",
				Flags: append(teamFlags, []cli.Flag{
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Set a project name in your stack",
					},
					cli.BoolFlag{
						Name:  "generate, g",
						Usage: "Auto-generate a new stack.yml based on existing resource",
					},
					yesFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadTeamPrefs, stack.Init),
			},
			{
				Name:      "add",
				ArgsUsage: "[resource-name]",
				Usage:     "Add a resource definition to the stack.yml",
				Flags: []cli.Flag{
					productFlag(),
					planFlag(),
					regionFlag(),
					titleFlag(),
				},
				Action: stackAddCmd,
			},
			{
				Name:      "remove",
				ArgsUsage: "[resource-name]",
				Usage:     "Remove a resource definition from the stack.yml",
				Action:    stackRemoveCmd,
			},
			// {
			// 	Name:      "plan",
			// 	Usage:     "Performs a dry run of the stack.yml, showing you the changes that will be made from an apply",
			// 	Action: stack.PlanCMD,
			// },
			// {
			// 	Name:  "apply",
			// 	Usage: "List members of a team",
			// 	Flags: append(teamFlags, []cli.Flag{
			// 		projectFlag(),
			// 		cli.BoolFlag{
			// 			Name:   "upgrade, u",
			// 			Usage:  "Upgrade any already provisioned products to plans set in stack.yml if upgradable",
			// 		},
			// 		cli.BoolFlag{
			// 			Name:   "interactive, i",
			// 			Usage:  "Prompt to confirm the project name and prompt to select a plan at every new resource creation, " +
			// 				"does a dry run and prompts for confirmation before actually running",
			// 			EnvVar: "MANIFOLD_INTERACTIVE",
			// 		},
			// 	}),
			// 	Action: stack.ApplyCMD,
			// },
		},
	}

	cmds = append(cmds, stackCmd)
}

func stackAddCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	planName, err := validateName(cliCtx, "plan")
	if err != nil {
		return err
	}

	productName, err := validateName(cliCtx, "product")
	if err != nil {
		return err
	}

	regionName, err := validateName(cliCtx, "region")
	if err != nil {
		return err
	}

	client, err := api.New(api.Catalog)
	if err != nil {
		return err
	}

	catalog, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	var plan *cModels.Plan

	products := catalog.Products()
	productIdx, _, err := prompts.SelectProduct(products, productName)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select product.")
	}

	if planName != "" {
		plan, err = catalog.FetchPlanByLabel(ctx, products[productIdx].ID, planName)
		if err != nil {
			return prompts.HandleSelectError(err, "Plan does not exist.")
		}
		_, _, err = prompts.SelectPlan([]*cModels.Plan{plan}, planName)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select unlisted plan.")
		}
	} else {
		plans := filterPlansByProductID(catalog.Plans(), products[productIdx].ID)
		planIdx, _, err := prompts.SelectPlan(plans, planName)
		if err != nil {
			return prompts.HandleSelectError(err, "Plan does not exist.")
		}
		plan = plans[planIdx]
	}

	regions := filterRegionsForPlan(catalog.Regions(), plan.Body.Regions)
	_, _, err = prompts.SelectRegion(regions)
	if err != nil {
		return prompts.HandleSelectError(err, "Region does not exist.")
	}

	resourceName, resourceTitle, err := promptNameAndTitle(cliCtx, nil, "resource", true, false)
	if err != nil {
		return err
	}

	// TODO: Init if does not exist
	// Read
	data, err := ioutil.ReadFile("stack.yml")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to read stack.yml file"), -1)
	}

	stackFile := &stack.StackYaml{}
	err = yaml.Unmarshal(data, stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to unmarshal YAML data: %s", err), -1)
	}

	// Modify
	stackFile.Resources[resourceName] = stack.StackResource{
		Title:   resourceTitle,
		Product: productName,
		Plan:    planName,
		Region:  regionName,
	}

	// Write
	data, err = yaml.Marshal(stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to marshal YAML data: %s", err), -1)
	}

	err = ioutil.WriteFile("stack.yml", data, 0644)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to save stack.yml file"), -1)
	}

	return nil
}

func stackRemoveCmd(cliCtx *cli.Context) error {
	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceName, err := optionalArgName(cliCtx, 0, "resource name")
	if err != nil {
		return err
	}
	if resourceName == "" {
		return cli.NewExitError(fmt.Sprintf("Please specify a resource name to remove"), -1)
	}

	// Read
	data, err := ioutil.ReadFile("stack.yml")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to read stack.yml file"), -1)
	}

	stackFile := &stack.StackYaml{}
	err = yaml.Unmarshal(data, stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to unmarshal YAML data: %s", err), -1)
	}

	// Check
	if _, ok := stackFile.Resources[resourceName]; !ok {
		return cli.NewExitError(fmt.Sprintf("Resource definition does not exist: %s", resourceName), -1)
	}

	// Modify
	delete(stackFile.Resources, resourceName)

	// Write
	data, err = yaml.Marshal(stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to marshal YAML data: %s", err), -1)
	}

	err = ioutil.WriteFile("stack.yml", data, 0644)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to save stack.yml file"), -1)
	}

	return nil
}
