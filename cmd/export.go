package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/credential"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

var formats = []string{"env", "bash", "powershell", "fish", "cmd", "json"}

func init() {

	formatFlagStr := fmt.Sprintf("Export format of the secrets (%s)", strings.Join(formats, ", "))
	exportCmd := cli.Command{
		Name:     "export",
		Usage:    "Export all environment variables from all resources",
		Category: "CONFIGURATION",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, export),
		Flags: append(teamFlags, []cli.Flag{
			formatFlag(formats[0], formatFlagStr),
			projectFlag(),
		}...),
	}

	cmds = append(cmds, exportCmd)
}

func export(cliCtx *cli.Context) error {
	ctx := context.Background()

	projectLabel, err := validateLabel(cliCtx, "project")
	if err != nil {
		return err
	}

	format := cliCtx.String("format")
	if !validFormat(format) {
		return cli.NewExitError("You provided an invalid format!", -1)
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Analytics, api.Marketplace)
	if err != nil {
		return err
	}

	prompts.SpinStart("Fetching Resources")
	resources, err := clients.FetchResources(ctx, client.Marketplace, teamID, projectLabel)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}

	if projectLabel == "" {
		resources = filterResourcesWithoutProjects(resources)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Body.Name < resources[j].Body.Name
	})

	cMap, err := fetchCredentials(ctx, client.Marketplace, resources)
	if err != nil {
		return cli.NewExitError("Could not retrieve credentials: "+err.Error(), -1)
	}

	params := map[string]string{
		format: format,
	}
	if projectLabel != "" {
		params["project"] = projectLabel
	}

	client.Analytics.Track(ctx, "Exported Credentials", &params)

	rMap := indexResources(resources)
	w := os.Stdout
	switch format {
	case "env":
		err = writeFormat(w, rMap, cMap, "%s=%s\n")
	case "bash":
		err = writeFormat(w, rMap, cMap, "export %s=%s\n")
	case "powershell":
		err = writeFormat(w, rMap, cMap, "$Env:%s = \"%s\"\n")
	case "cmd":
		err = writeFormat(w, rMap, cMap, "set %s=%s\n")
	case "fish":
		err = writeFormat(w, rMap, cMap, "set -x %s %s;\n")
	case "json":
		err = writeJSON(w, cMap)
	default:
		return cli.NewExitError("Unrecognized format value: "+format, -1)
	}

	if err != nil {
		cli.NewExitError("Could not output to format: "+err.Error(), -1)
	}

	return nil
}

func validFormat(format string) bool {
	for _, f := range formats {
		if f == format {
			return true
		}
	}

	return false
}

func writeFormat(w io.Writer, rMap map[manifold.ID]*models.Resource,
	cMap map[manifold.ID][]*models.Credential, format string) error {
	for rID, credentials := range cMap {
		resource := rMap[rID]
		fmt.Fprintf(w, "# %s\n", resource.Body.Name)
		for _, c := range credentials {
			for name, value := range c.Body.Values {
				fmt.Fprintf(w, format, name, value)
			}
		}

		fmt.Fprintf(w, "\n")
	}

	return nil
}

func writeJSON(w io.Writer, cMap map[manifold.ID][]*models.Credential) error {
	credentials, err := flattenCMap(cMap)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(credentials, "", "    ")
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%s\n", b)
	return nil
}

func flattenCMap(cMap map[manifold.ID][]*models.Credential) (map[string]string, error) {
	out := make(map[string]string)

	for _, credentials := range cMap {
		for _, c := range credentials {
			for name, value := range c.Body.Values {
				out[name] = value
			}
		}
	}

	return out, nil
}

func indexResources(resources []*models.Resource) map[manifold.ID]*models.Resource {
	index := make(map[manifold.ID]*models.Resource)
	for _, resource := range resources {
		index[resource.ID] = resource
	}

	return index
}

func fetchCredentials(ctx context.Context, m *mClient.Marketplace, resources []*models.Resource) (map[manifold.ID][]*models.Credential, error) {
	// XXX: Reduce this into a single HTTP Call
	//
	// Issue: https://www.github.com/manifoldco/engineering#2536
	cMap := make(map[manifold.ID][]*models.Credential)
	for _, r := range resources {
		p := credential.NewGetCredentialsParamsWithContext(ctx).WithResourceID([]string{r.ID.String()})
		c, err := m.Credential.GetCredentials(p, nil)
		if err != nil {
			return nil, err
		}

		if _, ok := cMap[r.ID]; !ok {
			cMap[r.ID] = []*models.Credential{}
		}

		for _, credential := range c.Payload {
			cMap[r.ID] = append(cMap[r.ID], credential)
		}
	}

	return cMap, nil
}

// filterResourcesWithoutProjects returns resources without a project id
func filterResourcesWithoutProjects(resources []*models.Resource) []*models.Resource {
	var results []*models.Resource
	for _, r := range resources {
		if r.Body.ProjectID == nil {
			results = append(results, r)
		}
	}
	return results
}
