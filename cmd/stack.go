package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/cmd/stack"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	projectClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/project"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
	"github.com/manifoldco/manifold-cli/session"
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
			{
				Name:  "apply",
				Usage: "Apply a stack.yaml to create services",
				Flags: append(teamFlags, []cli.Flag{
					projectFlag(), yesFlag(), skipFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadTeamPrefs, apply),
			},
			// {
			// 	Name:      "plan",
			// 	Usage:     "Performs a dry run of the stack.yml, showing you the changes that will be made from an apply",
			// 	Action: stack.PlanCMD,
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
	product := products[productIdx]
	productName = string(product.Body.Label)

	if planName != "" {
		plan, err = catalog.FetchPlanByLabel(ctx, product.ID, planName)
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
	planName = string(plan.Body.Label)

	regions := filterRegionsForPlan(catalog.Regions(), plan.Body.Regions)
	regionIdx, _, err := prompts.SelectRegion(regions)
	if err != nil {
		return prompts.HandleSelectError(err, "Region does not exist.")
	}
	regionName = *regions[regionIdx].Body.Location

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

	stackFile := &config.StackYaml{}
	err = yaml.Unmarshal(data, stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to unmarshal YAML data: %s", err), -1)
	}

	// Modify
	stackFile.Resources[resourceName] = config.StackResource{
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

	// Read
	data, err := ioutil.ReadFile("stack.yml")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to read stack.yml file"), -1)
	}

	stackFile := &config.StackYaml{}
	err = yaml.Unmarshal(data, stackFile)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to unmarshal YAML data: %s", err), -1)
	}

	if len(stackFile.Resources) == 0 {
		return cli.NewExitError(fmt.Sprintf("No resources to remove, stack.yml is empty"), -1)
	}

	if resourceName == "" {
		_, resourceName, err = prompts.SelectStackResource(stackFile.Resources, resourceName)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to select resource from stack.yml file"), -1)
		}
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

	fmt.Printf("\n%s has been removed from your stack.yml!\n", resourceName)

	return nil
}
func apply(cliCtx *cli.Context) error {
	ctx := context.Background()

	stackYaml, err := ioutil.ReadFile("stack.yml")
	if err != nil {
		return cli.NewExitError("Cannot read stack.yml", -1)
	}

	var stack stack.StackYaml
	err = yaml.Unmarshal(stackYaml, &stack)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Cannot decode stack.yaml: %s", err), -1)
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	dontWait := cliCtx.Bool("skip")

	client, err := api.New(api.Marketplace, api.Provisioning, api.Catalog)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	projectName := cliCtx.String("project")
	if projectName != "" {
		stack.Project = projectName
	}

	project, err := clients.FetchProjectByLabel(ctx, client.Marketplace, teamID, stack.Project)
	if err != nil {
		params := projectClient.NewPostProjectsParamsWithContext(ctx)
		params.SetBody(&mModels.CreateProject{
			&mModels.CreateProjectBody{
				Label: manifold.Label(stack.Project),
				Name:  manifold.Name(generateTitle(stack.Project)),
			},
		})

		if teamID != nil {
			params.Body.Body.TeamID = teamID
		} else {
			params.Body.Body.UserID = userID
		}

		err := createProject(params)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Cannot create project: %s", err), -1)
		}

		fmt.Printf("Project %s created\n", stack.Project)
	} else {
		fmt.Printf("Project %s exists, not re-creating\n", stack.Project)
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Unable to load config", -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Unable to load session", -1)
	}

	cat, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return err
	}

	resources, err := clients.FetchResources(ctx, client.Marketplace, teamID, stack.Project)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to fetch resources %s:", err), -1)
	}

	resourceMap := make(map[manifold.Label]*mModels.Resource)
	for _, resource := range resources {
		resourceMap[resource.Body.Label] = resource
	}

	for resourceLabel, details := range stack.Resources {
		fmt.Printf("Updating resource %s\n", resourceLabel)

		if _, ok := resourceMap[manifold.Label(resourceLabel)]; !ok {
			fmt.Printf("Resource %s not found, creating\n", resourceLabel)

			resourceID, err := manifold.NewID(idtype.Resource)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Failed to generate resource ID: %s", err), -1)
			}

			var product *cModels.Product
			products := cat.Products()
			for _, p := range products {
				if p.Body.Label == manifold.Label(details.Product) {
					product = p
					break
				}
			}

			plan, err := cat.FetchPlanByLabel(ctx, product.ID, details.Plan)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Unable to find plan by label %s: %s", details.Plan, err), -1)
			}

			var region *cModels.Region
			regions := cat.Regions()
			for _, r := range regions {
				if *r.Body.Location == details.Region {
					region = r
					break
				}
			}

			spin := prompts.NewSpinner(fmt.Sprintf("Creating %s", resourceLabel))
			if !dontWait {
				spin.Start()
				defer spin.Stop()
			}

			op, err := createResource(ctx, cfg, &resourceID, teamID, s, client.Provisioning,
				false, product, plan, region, project, resourceLabel, details.Title, dontWait)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Unable to create resource: %s", err), -1)
			}

			provision := op.Body.(*pModels.Provision)
			if !dontWait {
				spin.Stop()
			}

			fmt.Printf("\nAn instance of %s named \"%s\" has been created\n",
				product.Body.Name, *provision.Name)
		}

		// TODO: resource exist, apply updates
	}

	return nil
}
