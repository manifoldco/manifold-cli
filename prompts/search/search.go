package search

import (
	"strings"

	"github.com/manifoldco/manifold-cli/generated/catalog/models"
)

var productTags map[string][]string

func init() {
	// FIXME bootstrap product tags while isn't supported by the backend.
	productTags = map[string][]string{
		"jawsmysql":    []string{"jawsdb", "mysql", "db", "database"},
		"jawsmaria":    []string{"jawsdb", "mariadb", "mysql", "db", "database"},
		"jawspostgres": []string{"jawsdb", "postgresql", "database"},
	}
}

// ProductSearch returns a search function using a hardcoded list of tags for
// production products.
func ProductSearch(products []*models.Product) func(string, int) bool {
	return func(input string, idx int) bool {
		product := products[idx]
		label := string(product.Body.Label)
		tags := productTags[label]

		for _, tag := range tags {
			if strings.Contains(tag, input) {
				return true
			}
		}
		return false
	}
}
