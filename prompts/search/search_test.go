package search

import (
	"testing"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/generated/catalog/models"
)

func TestProductSearch(t *testing.T) {
	products := []*models.Product{
		{Body: &models.ProductBody{Label: manifold.Label("jawsmysql")}},
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
			label := string(product.Body.Label)

			if match != tc.match {
				t.Errorf("Expected %q to match %v", tc.input, productTags[label])
			}
		})
	}

}
