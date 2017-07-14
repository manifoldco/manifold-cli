package catalog

import (
	"context"
	"errors"

	hierr "github.com/reconquest/hierr-go"

	catalogClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	catalogClientPlan "github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	catalogModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

// Catalog represents a local in memory cache of catalog data
type Catalog struct {
	client   *catalogClient.Catalog
	products map[string]*catalogModels.Product
	plans    map[string]*catalogModels.Plan
	regions  map[string]*catalogModels.Region
}

// GetProduct returns the Product data model based on the provided string id
func (c *Catalog) GetProduct(id string) (*catalogModels.Product, error) {
	product, ok := c.products[id]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return product, nil
}

// GetPlan returns the Plan data model based on the provided string id
func (c *Catalog) GetPlan(id string) (*catalogModels.Plan, error) {
	plan, ok := c.plans[id]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return plan, nil
}

// GetRegion returns the Region data model based on the provided string id
func (c *Catalog) GetRegion(id string) (*catalogModels.Region, error) {
	region, ok := c.regions[id]
	if !ok {
		return nil, errors.New("Product not found")
	}
	return region, nil
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
	products, err := cache.client.Product.GetProducts(nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest products list")
	}

	// Map products catalog, so its searchable by id through hashmap
	cache.products = make(map[string]*catalogModels.Product)
	productIDs := make([]string, len(products.Payload))
	for i, product := range products.Payload {
		productID := product.ID.String()
		cache.products[productID] = product
		productIDs[i] = productID
	}

	// Get plans for known productIDs
	planParams := catalogClientPlan.NewGetPlansParamsWithContext(ctx)
	planParams.SetProductID(productIDs)
	plans, err := cache.client.Plan.GetPlans(planParams)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the latest product plan data")
	}

	// Map plan catalog, so its searchable by id through hashmap
	cache.plans = make(map[string]*catalogModels.Plan)
	for _, plan := range plans.Payload {
		cache.plans[plan.ID.String()] = plan
	}

	// Get regions
	regions, err := cache.client.Region.GetRegions(nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest region data")
	}

	// Map region catalog, so its searchable by id through hashmap
	cache.regions = make(map[string]*catalogModels.Region)
	for _, region := range regions.Payload {
		cache.regions[region.ID.String()] = region
	}

	return cache, nil
}
