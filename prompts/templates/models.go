package templates

import (
	"github.com/manifoldco/manifold-cli/config"

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

type StackResource struct {
	Name    string
	Title   string
	Product string
	Plan    string
	Region  string
}

func (sr StackResource) String() string {
	return sr.Name
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

func StackResources(stackResources map[string]config.StackResource) []StackResource {
	resources := make([]StackResource, len(stackResources))

	var i int = 0
	for k, v := range stackResources {
		resources[i] = StackResource{
			Name:    k,
			Title:   v.Title,
			Product: v.Product,
			Plan:    v.Plan,
			Region:  v.Region,
		}
		i++
	}

	return resources
}
