package prompts

import (
	"fmt"

	"github.com/manifoldco/manifold-cli/errs"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/prompts/templates"
	"github.com/manifoldco/promptui"
)

// SelectProduct prompts the user to select a product from the given list.
func SelectProduct(products []*cModels.Product, label string) (int, string, error) {
	var idx int
	if label != "" {
		found := false
		for i, p := range products {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Product", label))
			return 0, "", errs.ErrProductNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Product", label)) // FIXME
		return idx, label, nil
	}

	prompt := promptui.Select{
		Label:     "Select Product",
		Items:     templates.Products(products),
		Templates: templates.TplProduct,
	}

	return prompt.Run()
}

// SelectPlan prompts the user to select a plan from the given list.
func SelectPlan(plans []*cModels.Plan, label string) (int, string, error) {
	var idx int
	if label != "" {
		found := false
		for i, p := range plans {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Plan", label))
			return 0, "", errs.ErrPlanNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Plan", label)) //FIXME
		return idx, label, nil
	}

	prompt := promptui.Select{
		Label:     "Select Plan",
		Items:     templates.Plans(plans),
		Templates: templates.TplPlan,
	}

	return prompt.Run()
}

// SelectResource promps the user to select a provisioned resource from the given list
func SelectResource(resources []*mModels.Resource, projects []*mModels.Project,
	label string) (int, string, error) {

	var idx int
	if label != "" {
		found := false
		for i, p := range resources {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Resource", label))
			return 0, "", errs.ErrResourceNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Resource", label)) //FIXME
		return idx, label, nil
	}

	prompt := promptui.Select{
		Label:     "Select Resource",
		Items:     templates.Resources(resources, projects),
		Templates: templates.TplResource,
	}

	return prompt.Run()
}

// SelectRole prompts the user to select a role from the given list.
func SelectRole() (string, error) {
	prompt := promptui.Select{
		Label: "Select Role",
		Items: []string{"read", "read-credentials", "write", "admin"},
	}
	_, role, err := prompt.Run()
	return role, err
}

// SelectRegion prompts the user to select a region from the given list.
func SelectRegion(regions []*cModels.Region) (int, string, error) {
	line := func(r *cModels.Region) string {
		return fmt.Sprintf("%s (%s::%s)", r.Body.Name, *r.Body.Platform, *r.Body.Location)
	}

	labels := make([]string, len(regions))
	for i, r := range regions {
		labels[i] = line(r)
	}

	// TODO: Build "auto" resolve into promptui in case of only one item
	if len(regions) == 1 {
		fmt.Println(promptui.SuccessfulValue("Region", line(regions[0])))
		return 0, string(regions[0].Body.Name), nil
	}

	prompt := promptui.Select{
		Label: "Select Region",
		Items: labels,
	}

	return prompt.Run()
}

// SelectProject prompts the user to select a project from the given list.
func SelectProject(mProjects []*mModels.Project, label string, emptyOption bool) (int, string, error) {
	projects := templates.Projects(mProjects)

	var idx int
	if label != "" {
		found := false
		for i, p := range mProjects {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Select Project", label))
			return 0, "", errs.ErrProjectNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Select Project", label)) //FIXME
		return idx, label, nil
	}

	if emptyOption {
		projects = append([]templates.Project{{Name: "No Project"}}, projects...)
	}

	prompt := promptui.Select{
		Label:     "Select Project",
		Items:     projects,
		Templates: templates.TplProject,
	}

	projectIdx, name, err := prompt.Run()

	if emptyOption {
		return projectIdx - 1, name, err
	}

	return projectIdx, name, err
}

// SelectContext runs a SelectTeam for context purposes
func SelectContext(teams []*iModels.Team, label string, userTuple *[]string) (int, string, error) {
	return selectTeam(teams, "Switch To", label, userTuple)
}

// SelectTeam prompts the user to select a team from the given list. -1 as the first return value
// indicates no team has been selected
func SelectTeam(teams []*iModels.Team, label string, userTuple *[]string) (int, string, error) {
	return selectTeam(teams, "Select Team", label, userTuple)
}

func selectTeam(mTeams []*iModels.Team, prefix, label string, userTuple *[]string) (int, string, error) {
	if prefix == "" {
		prefix = "Select Team"
	}

	var idx int
	if label != "" {
		found := false
		for i, t := range mTeams {
			if string(t.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Team", label))
			return 0, "", errs.ErrTeamNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Team", label)) // FIXME
		return idx, label, nil
	}

	teams := templates.Teams(mTeams)

	if userTuple != nil {
		u := *userTuple
		user := templates.Team{
			Name:  u[0],
			Title: u[1],
		}

		// treat user as a team for display purposes
		teams = append([]templates.Team{user}, teams...)
	}

	tpl := templates.TplTeam
	tpl.Selected = fmt.Sprintf(tpl.Selected, prefix)

	prompt := promptui.Select{
		Label:     prefix,
		Items:     teams,
		Templates: tpl,
	}

	teamIdx, name, err := prompt.Run()

	if userTuple != nil {
		return teamIdx - 1, name, err
	}

	return teamIdx, name, err
}

// SelectProvider prompts the user to select a provider resource from the given
// list.
func SelectProvider(mProviders []*cModels.Provider) (*cModels.Provider, error) {
	providers := templates.Providers(mProviders)

	label := templates.Provider{Name: "All Providers"}
	providers = append([]templates.Provider{label}, providers...)

	prompt := promptui.Select{
		Label:     "Select Provider",
		Items:     providers,
		Templates: templates.TplProvider,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	if idx == 0 {
		return nil, nil
	}

	return mProviders[idx-1], nil
}

// SelectAPIToken prompts the user to choose from a list of tokens
func SelectAPIToken(tokens []*iModels.APIToken) (*iModels.APIToken, error) {
	var labels []string
	for _, t := range tokens {
		val := fmt.Sprintf("%s****%s", *t.Body.FirstFour, *t.Body.LastFour)
		labels = append(labels, fmt.Sprintf("%s - %s", val, *t.Body.Description))
	}

	prompt := promptui.Select{
		Label: "Select API token",
		Items: labels,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return tokens[idx], nil
}
