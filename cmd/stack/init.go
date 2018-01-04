package stack

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/prompts"
)

func Init(cliCtx *cli.Context) error {
	yes := cliCtx.Bool("yes")

	existing, err := ioutil.ReadFile("stack.yml")
	if err != nil && !os.IsNotExist(err) || err == nil && len(existing) != 0 {
		if !yes {
			_, err := prompts.Confirm("Overwrite stack.yml?")
			if err != nil {
				return cli.NewExitError("Not overwriting stack.yml", -1)
			}
		}
	}

	projectTitle := ""
	if cliCtx.IsSet("project") {
		projectTitle = cliCtx.String("project")
	}

	projectTitle, err = prompts.ProjectTitle(projectTitle, projectTitle != "")
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Unable to save project name: %s", err), -1)
	}

	stack := config.StackYaml{
		Project: projectTitle,
	}

	data, err := yaml.Marshal(stack)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to marshal YAML data: %s", err), -1)
	}

	err = ioutil.WriteFile("stack.yml", data, 0644)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to save stack.yml file"), -1)
	}

	return nil
}
