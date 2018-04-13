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

	projectName, err := validateName(cliCtx, "project")
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

	// we need to fetch all the credentials for the rMap for resource naming, etc
	prompts.SpinStart("Fetching Resources")
	resources, err := clients.FetchResources(ctx, client.Marketplace, teamID, projectName)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Body.Name < resources[j].Body.Name
	})

	cMap := make(map[manifold.ID][]*models.Credential)
	if projectName == "" {
		p, err := clients.FetchProjectByLabel(ctx, client.Marketplace, teamID, projectName)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not retrieve project: %s", err), -1)
		}

		cMap, err = fetchProjectCredentials(ctx, client.Marketplace, p, true)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not retrieve credentials: %s", err), -1)
		}
	} else {
		cMap, err = fetchResourceCredentials(ctx, client.Marketplace, resources, true)
		if err != nil {
			return cli.NewExitError("Could not retrieve credentials: "+err.Error(), -1)
		}
	}

	params := map[string]string{
		format: format,
	}
	if projectName != "" {
		params["project"] = projectName
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

func fetchResourceCredentials(ctx context.Context, m *mClient.Marketplace, resources []*models.Resource, customNames bool) (map[manifold.ID][]*models.Credential, error) {

	p := credential.NewGetCredentialsParamsWithContext(ctx)
	if customNames == false {
		noCustomNames := "false"
		p.SetCustomNames(&noCustomNames)
	}
	// Prep the credential map with the known resources, and build a list
	// of IDs to use in the resource_id query param
	var resourceIDs []string
	for _, r := range resources {
		resourceIDs = append(resourceIDs, r.ID.String())
	}
	p.SetResourceID(resourceIDs)

	return fetchCredentials(m, p)
}

func fetchProjectCredentials(ctx context.Context, m *mClient.Marketplace, project *models.Project, customNames bool) (map[manifold.ID][]*models.Credential, error) {
	_ = make(map[manifold.ID][]*models.Credential)
	p := credential.NewGetCredentialsParamsWithContext(ctx)

	if customNames == false {
		noCustomNames := "false"
		p.SetCustomNames(&noCustomNames)
	}

	pid := project.ID.String()
	p.SetProjectID(&pid)

	return fetchCredentials(m, p)
}

func fetchCredentials(m *mClient.Marketplace, params *credential.GetCredentialsParams) (map[manifold.ID][]*models.Credential, error) {
	cMap := make(map[manifold.ID][]*models.Credential)

	// Get credentials for all the defined resources withring(call
	c, err := m.Credential.GetCredentials(params, nil)
	if err != nil {
		return nil, err
	}
	// Append credential results to the map
	for _, credential := range c.Payload {
		rid := credential.Body.ResourceID
		if _, ok := cMap[rid]; !ok {
			cMap[rid] = []*models.Credential{}
		}
		cMap[credential.Body.ResourceID] = append(cMap[credential.Body.ResourceID], credential)
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
