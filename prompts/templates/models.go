package templates

import (
	"sort"
	"strings"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

type Provider struct {
	Name  string
	Title string
}

type Product struct {
	Name     string
	Title    string
	Tagline  string
	Features []string
}

type Plan struct {
	Name     string
	Title    string
	Cost     int
	Features []Feature
}

type Feature struct {
	Name        string
	Description string
}

type Region struct {
	Name     string
	Platform string
	Location string
}

type Resource struct {
	Name    string
	Title   string
	Project string
}

type Project struct {
	Name  string
	Title string
}

type Team struct {
	Name  string
	Title string
}

func Resources(list []*mModels.Resource, projects []*mModels.Project) []Resource {
	resources := make([]Resource, len(list))

	for i, m := range list {
		r := Resource{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}

		if m.Body.ProjectID != nil {
			for _, p := range projects {
				if *m.Body.ProjectID == p.ID {
					r.Project = string(p.Body.Label)
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
		p := Provider{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}
		providers[i] = p
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

func Projects(list []*mModels.Project) []Project {
	projects := make([]Project, len(list))

	for i, m := range list {
		p := Project{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}
		projects[i] = p
	}

	sort.Slice(projects, func(i, j int) bool {
		a := string(projects[i].Name)
		b := string(projects[j].Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return projects
}

func Products(list []*cModels.Product) []Product {
	products := make([]Product, len(list))

	for i, m := range list {
		p := Product{
			Name:    string(m.Body.Label),
			Title:   string(m.Body.Name),
			Tagline: m.Body.Tagline,
		}
		for _, v := range m.Body.ValueProps {
			p.Features = append(p.Features, v.Header)
		}

		products[i] = p
	}

	sort.Slice(products, func(i, j int) bool {
		a := string(products[i].Name)
		b := string(products[j].Name)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return products
}

func Plans(list []*cModels.Plan) []Plan {
	plans := make([]Plan, len(list))

	for i, m := range list {
		p := Plan{
			Name:  string(m.Body.Label),
			Title: string(m.Body.Name),
		}

		if m.Body.Cost != nil {
			p.Cost = int(*m.Body.Cost)
		}

		features := make([]Feature, len(m.Body.Features))

		for j, v := range m.Body.Features {
			f := Feature{
				Name: string(v.Feature),
			}
			if v.Value != nil {
				f.Description = *v.Value
			}
			features[j] = f
		}

		p.Features = features
		plans[i] = p
	}

	return plans
}

func Regions(list []*cModels.Region) []Region {
	regions := make([]Region, len(list))

	for i, m := range list {
		r := Region{
			Name: string(m.Body.Name),
		}

		if m.Body.Location != nil {
			r.Location = *m.Body.Location
		}

		if m.Body.Platform != nil {
			r.Platform = *m.Body.Platform
		}

		regions[i] = r
	}

	return regions
}
