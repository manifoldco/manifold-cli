package search

import (
	"strings"

	"github.com/manifoldco/manifold-cli/prompts/templates"
)

var productTags map[string][]string

func init() {
	// FIXME bootstrap product tags while isn't supported by the backend.
	productTags = map[string][]string{
		"jawsdb-mysql":     []string{"database"},
		"jawsdb-maria":     []string{"database"},
		"jawsdb-postgres":  []string{"database"},
		"mailgun":          []string{"email"},
		"cloudamqp":        []string{"rabbitmq"},
		"memcachier-cache": []string{"memcache"},
		"scoutapp":         []string{"memory leak", "monitoring", "ruby", "elixir"},
	}
}

// ProductSearch returns a search function using a hardcoded list of tags for
// production products.
func ProductSearch(products []templates.Product) func(string, int) bool {
	return func(input string, idx int) bool {
		product := products[idx]
		name := string(product.Name)
		tags := productTags[name]
		tags = append(tags, name)

		for _, tag := range tags {
			if strings.Contains(tag, input) {
				return true
			}
		}
		return false
	}
}

// PlanSearch returns a search function to filter by plan name.
func PlanSearch(plans []templates.Plan) func(string, int) bool {
	return func(input string, idx int) bool {
		plan := plans[idx]
		name := string(plan.Name)
		return strings.Contains(name, input)
	}
}

// RegionSearch returns a search function to filter by region name.
func RegionSearch(regions []templates.Region) func(string, int) bool {
	return func(input string, idx int) bool {
		region := regions[idx]
		name := strings.ToLower(string(region.Name))
		return strings.Contains(name, input)
	}
}
