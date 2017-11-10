package search

import (
	"testing"

	"github.com/manifoldco/manifold-cli/prompts/templates"
)

func TestProductSearch(t *testing.T) {
	products := []templates.Product{
		{Name: "jawsdb-mysql"},
	}

	tcs := []struct {
		scenario string
		input    string
		idx      int
		match    bool
	}{
		{
			scenario: "when input is a partial match",
			input:    "sql",
			idx:      0,
			match:    true,
		},
		{
			scenario: "when input is a complete match",
			input:    "database",
			idx:      0,
			match:    true,
		},
		{
			scenario: "when input has additional characters",
			input:    "databases",
			idx:      0,
			match:    false,
		},
	}

	search := ProductSearch(products)

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			match := search(tc.input, tc.idx)
			product := products[tc.idx]
			name := product.Name

			if match != tc.match {
				t.Errorf("Expected %q to match %v", tc.input, productTags[name])
			}
		})
	}

}
