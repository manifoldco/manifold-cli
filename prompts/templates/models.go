package templates

import (
	"sort"
	"strings"

	manifold "github.com/manifoldco/go-manifold"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"

	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

type Provider struct {
	Name  string
	Title string
}

type Resource struct {
	Name    manifold.Label
	Title   manifold.Name
	Project manifold.Label
}

type Project struct {
	Name  manifold.Label
	Title manifold.Name
}

type Team struct {
	Name  string
	Title string
}

func Resources(list []*mModels.Resource, projects []*mModels.Project) []Resource {
	resources := make([]Resource, len(list))

	for i, m := range list {
		r := Resource{
			Name:  m.Body.Label,
			Title: m.Body.Name,
		}

		if m.Body.ProjectID != nil {
			for _, p := range projects {
				if *m.Body.ProjectID == p.ID {
					r.Project = p.Body.Label
				}
			}
		}

		resources[i] = r
	}

	sort.Slice(resources, func(i, j int) bool {
		a := string(resources[i].Name)
		b := string(resources[j].Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return resources
}

func Providers(list []*cModels.Provider) []Provider {
	providers := make([]Provider, len(list))

	for i, m := range list {
		t := Provider{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}
		providers[i] = t
	}

	sort.Slice(providers, func(i, j int) bool {
		a := string(providers[i].Name)
		b := string(providers[j].Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return providers
}

func Teams(list []*iModels.Team) []Team {
	teams := make([]Team, len(list))

	for i, m := range list {
		t := Team{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}
		teams[i] = t
	}

	sort.Slice(teams, func(i, j int) bool {
		a := string(teams[i].Name)
		b := string(teams[j].Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return teams
}
