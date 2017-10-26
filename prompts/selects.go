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
func SelectProduct(list []*cModels.Product, name string) (int, string, error) {
	products := templates.Products(list)
	tpls := templates.TplProduct

	var idx int
	if name != "" {
		found := false
		for i, p := range products {
			if p.Name == name {
				idx = i
				found = true
				break
			}
		}

		if !found {
			msg := templates.PromptFailure("Product", name)
			fmt.Println(msg)
			return 0, "", errs.ErrProductNotFound
		}

		msg := templates.SelectSuccess(tpls, products[idx])
		fmt.Println(msg)
		return idx, name, nil
	}

	prompt := promptui.Select{
		Label:     "Select Product",
		Items:     products,
		Templates: tpls,
	}

	return prompt.Run()
}

// SelectPlan prompts the user to select a plan from the given list.
func SelectPlan(list []*cModels.Plan, name string) (int, string, error) {
	plans := templates.Plans(list)
	tpls := templates.TplPlan

	var idx int
	if name != "" {
		found := false
		for i, p := range plans {
			if p.Name == name {
				idx = i
				found = true
				break
			}
		}

		if !found {
			msg := templates.PromptFailure("Plan", name)
			fmt.Println(msg)
			return 0, "", errs.ErrPlanNotFound
		}

		msg := templates.SelectSuccess(tpls, plans[idx])
		fmt.Println(msg)
		return idx, name, nil
	}

	prompt := promptui.Select{
		Label:     "Select Plan",
		Items:     plans,
		Templates: tpls,
	}

	return prompt.Run()
}

// SelectResource promps the user to select a provisioned resource from the given list
func SelectResource(list []*mModels.Resource, projects []*mModels.Project,
	name string) (int, string, error) {
	resources := templates.Resources(list, projects)
	tpls := templates.TplResource

	var idx int
	if name != "" {
		found := false
		for i, r := range resources {
			if r.Name == name {
				idx = i
				found = true
				break
			}
		}

		if !found {
			msg := templates.PromptFailure("Resource", name)
			fmt.Println(msg)
			return 0, "", errs.ErrResourceNotFound
		}

		msg := templates.SelectSuccess(tpls, resources[idx])
		fmt.Println(msg)

		return idx, name, nil
	}

	prompt := promptui.Select{
		Label:     "Select Resource",
		Items:     resources,
		Templates: tpls,
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
func SelectRegion(list []*cModels.Region) (int, string, error) {
	regions := templates.Regions(list)
	tpls := templates.TplRegion

	// TODO: Build "auto" resolve into promptui in case of only one item
	if len(regions) == 1 {
		msg := templates.SelectSuccess(tpls, regions[0])
		fmt.Println(msg)

		return 0, string(regions[0].Name), nil
	}

	prompt := promptui.Select{
		Label:     "Select Region",
		Items:     regions,
		Templates: tpls,
	}

	return prompt.Run()
}

// SelectProject prompts the user to select a project from the given list.
func SelectProject(list []*mModels.Project, name string, emptyOption, showResult bool) (int, string, error) {
	projects := templates.Projects(list)
	tpls := templates.TplProject

	var idx int
	if name != "" {
		found := false
		for i, p := range projects {
			if p.Name == name {
				idx = i
				found = true
				break
			}
		}

		if !found {
			msg := templates.PromptFailure("Project", name)
			fmt.Println(msg)
			return 0, "", errs.ErrProjectNotFound
		}

		if showResult {
			msg := templates.SelectSuccess(tpls, projects[idx])
			fmt.Println(msg)
		}

		return idx, name, nil
	}

	if emptyOption {
		projects = append([]templates.Project{{Name: "No Project"}}, projects...)
	}

	prompt := promptui.Select{
		Label:     "Select Project",
		Items:     projects,
		Templates: tpls,
	}

	projectIdx, pname, err := prompt.Run()

	if emptyOption {
		return projectIdx - 1, pname, err
	}

	return projectIdx, pname, err
}

// SelectContext runs a SelectTeam for context purposes
func SelectContext(teams []*iModels.Team, name string, userTuple *[]string) (int, string, error) {
	return selectTeam(teams, "Switch To", name, userTuple)
}

// SelectTeam prompts the user to select a team from the given list. -1 as the first return value
// indicates no team has been selected
func SelectTeam(teams []*iModels.Team, name string, userTuple *[]string) (int, string, error) {
	return selectTeam(teams, "Select Team", name, userTuple)
}

func selectTeam(list []*iModels.Team, label, name string, userTuple *[]string) (int, string, error) {
	if label == "" {
		label = "Select Team"
	}

	teams := templates.Teams(list)
	tpls := templates.TplTeam
	tpls.Selected = fmt.Sprintf(tpls.Selected, label)

	var idx int
	if name != "" {
		found := false
		for i, t := range teams {
			if t.Name == name {
				idx = i
				found = true
				break
			}
		}

		if !found {
			msg := templates.PromptFailure("Team", name)
			fmt.Println(msg)
			return 0, "", errs.ErrTeamNotFound
		}

		msg := templates.SelectSuccess(tpls, teams[idx])
		fmt.Println(msg)

		return idx, name, nil
	}

	if userTuple != nil {
		u := *userTuple
		user := templates.Team{
			Name:  u[0],
			Title: u[1],
		}

		// treat user as a team for display purposes
		teams = append([]templates.Team{user}, teams...)
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     teams,
		Templates: tpls,
	}

	teamIdx, tname, err := prompt.Run()

	if userTuple != nil {
		return teamIdx - 1, tname, err
	}

	return teamIdx, tname, err
}

// SelectProvider prompts the user to select a provider resource from the given
// list.
func SelectProvider(list []*cModels.Provider) (*cModels.Provider, error) {
	providers := templates.Providers(list)

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

	return list[idx-1], nil
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
