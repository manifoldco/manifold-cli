package stack

import (
	"io/ioutil"
	"os"
	"fmt"

	"gopkg.in/yaml.v2"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/prompts"
)

func Init(cliCtx *cli.Context) error {
	yes := cliCtx.Bool("yes")

	existing, err := ioutil.ReadFile("stack.yaml");
	if err != nil && !os.IsNotExist(err) || err == nil && len(existing) != 0 {
		if !yes {
			_, err := prompts.Confirm("Overwrite stack.yaml?")
			if err != nil {
				return cli.NewExitError("Not overwriting stack.yaml", -1)
			}
		}
	}

	projectTitle, err := prompts.ProjectTitle("", false)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Unable to save project name: %s", err), -1)
	}

	stack := StackYaml{
		Project: projectTitle,
	}

	data, err := yaml.Marshal(stack)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to marshal YAML data: %s", err), -1)
	}

	err = ioutil.WriteFile("stack.yaml", data, 0644)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to save stack.yaml file"), -1)
	}

	return nil
}
