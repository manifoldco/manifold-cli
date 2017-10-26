package catalog

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/manifoldco/go-manifold"
	hierr "github.com/reconquest/hierr-go"

	catalogClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	catalogClientPlan "github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	catalogModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

// Catalog represents a local in memory cache of catalog data
type Catalog struct {
	client   *catalogClient.Catalog
	products map[manifold.ID]*catalogModels.Product
	plans    map[manifold.ID]*catalogModels.Plan
	regions  map[manifold.ID]*catalogModels.Region
}

// GetProduct returns the Product data model based on the provided id
func (c *Catalog) GetProduct(ID manifold.ID) (*catalogModels.Product, error) {
	product, ok := c.products[ID]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return product, nil
}

// GetPlan returns the Plan data model based on the provided id
func (c *Catalog) GetPlan(ID manifold.ID) (*catalogModels.Plan, error) {
	plan, ok := c.plans[ID]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return plan, nil
}

// GetRegion returns the Region data model based on the provided string id
func (c *Catalog) GetRegion(ID manifold.ID) (*catalogModels.Region, error) {
	region, ok := c.regions[ID]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return region, nil
}

// Returns a list of Plans from the Catalog
func (c *Catalog) Plans() []*catalogModels.Plan {
	plans := make([]*catalogModels.Plan, 0, len(c.plans))
	for _, p := range c.plans {
		plans = append(plans, p)
	}

	sort.Slice(plans, func(i, j int) bool {
		a := plans[i]
		b := plans[j]

		if *a.Body.Cost == *b.Body.Cost {
			return strings.ToLower(string(a.Body.Label)) <
				strings.ToLower(string(b.Body.Label))
		}
		return *a.Body.Cost < *b.Body.Cost
	})

	return plans
}

// Returns a list of Products from the Catalog
func (c *Catalog) Products() []*catalogModels.Product {
	products := make([]*catalogModels.Product, 0, len(c.products))
	for _, p := range c.products {
		products = append(products, p)
	}

	sort.Slice(products, func(i, j int) bool {
		a := string(products[i].Body.Label)
		b := string(products[j].Body.Label)
		return strings.ToLower(a) < strings.ToLower(b)
	})

	return products
}

// Returns a list of Regions from the Catalog
func (c *Catalog) Regions() []*catalogModels.Region {
	regions := make([]*catalogModels.Region, 0, len(c.regions))
	for _, r := range c.regions {
		regions = append(regions, r)
	}

	sort.Slice(regions, func(i, j int) bool {
		a := regions[i]
		b := regions[j]
		return strings.ToLower(string(a.Body.Name)) < strings.ToLower(string(b.Body.Name))
	})

	return regions
}

// Sync attempts to update the catalog and returns an error if anything went
// wrong
func (c *Catalog) Sync(ctx context.Context) error {
	_, err := updateCatalog(ctx, c)
	return err
}

// New creates a new instance of a Catalog struct and populates it with data
// from the API using the provided Catalog API client and context
func New(ctx context.Context, client *catalogClient.Catalog) (*Catalog, error) {
	return updateCatalog(ctx, &Catalog{client: client})
}

func updateCatalog(ctx context.Context, cache *Catalog) (*Catalog, error) {
	// Get products
	products, err := cache.client.Product.GetProducts(nil, nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest products list")
	}

	// Map products catalog, so its searchable by id through hashmap
	cache.products = make(map[manifold.ID]*catalogModels.Product)
	productIDs := make([]string, len(products.Payload))
	for i, product := range products.Payload {
		productID := product.ID
		cache.products[productID] = product
		productIDs[i] = productID.String()
	}

	// Get plans for known productIDs
	planParams := catalogClientPlan.NewGetPlansParamsWithContext(ctx)
	planParams.SetProductID(productIDs)
	plans, err := cache.client.Plan.GetPlans(planParams, nil)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the latest product plan data")
	}

	// Map plan catalog, so its searchable by id through hashmap
	cache.plans = make(map[manifold.ID]*catalogModels.Plan)
	for _, plan := range plans.Payload {
		cache.plans[plan.ID] = plan
	}

	// Get regions
	regions, err := cache.client.Region.GetRegions(nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest region data")
	}

	// Map region catalog, so its searchable by id through hashmap
	cache.regions = make(map[manifold.ID]*catalogModels.Region)
	for _, region := range regions.Payload {
		cache.regions[region.ID] = region
	}

	return cache, nil
}
