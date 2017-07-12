package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	manifold "github.com/manifoldco/go-manifold"
	hierr "github.com/reconquest/hierr-go"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"

	catalogClient "github.com/manifoldco/manifold-cli/generated/catalog/client"
	catalogClientPlan "github.com/manifoldco/manifold-cli/generated/catalog/client/plan"
	catalogModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	marketplaceClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	marketplaceModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

// CompleteResource is a structure that encapsulates all the core data about a
// resource for the purpose of the CLI
type CompleteResource struct {
	ID       manifold.ID
	Resource *marketplaceModels.ResourceBody
	Product  *CompleteProduct
	Plan     *catalogModels.PlanBody
	Region   *catalogModels.RegionBody
}

// CompleteProduct is a structure that encapsulates all the core data about a
// product from the catalog for the purpose of the CLI
type CompleteProduct struct {
	ID      manifold.ID
	Product *catalogModels.ProductBody
	Plans   map[string]*catalogModels.PlanBody
}

// CatalogCache represents a local in memory cache of catalog data
type CatalogCache struct {
	catalogClient *catalogClient.Catalog

	products map[string]*CompleteProduct
	regions  map[string]*catalogModels.RegionBody
}

// ResourceCache represents a local in memory cache of a users resource data
type ResourceCache struct {
	marketplaceClient *marketplaceClient.Marketplace

	resources []*CompleteResource
}

type resourcesSortByName []*CompleteResource

func (r resourcesSortByName) Len() int {
	return len(r)
}
func (r resourcesSortByName) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r resourcesSortByName) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(fmt.Sprintf("%s", r[i].Resource.Name)),
		fmt.Sprintf("%s", r[j].Resource.Name)) < 0
}

// GenerateCatalogCache builds a local in memory cache of catalog data for
// searching
func GenerateCatalogCache(ctx context.Context,
	cfg *config.Config) (*CatalogCache, error) {

	cache := &CatalogCache{}

	// Initialize clients if needed
	var err error
	if cache.catalogClient == nil {
		cache.catalogClient, err = clients.NewCatalog(cfg)
		if err != nil {
			return nil, hierr.Errorf(err, "Failed to create the Catalog API client")
		}
	}

	// Get products
	products, err := cache.catalogClient.Product.GetProducts(nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest products list")
	}

	// Map products catalog, so its searchable by id through hashmap
	cache.products = make(map[string]*CompleteProduct)
	productIDs := make([]string, len(products.Payload))
	for i, product := range products.Payload {
		productID := product.ID.String()
		cache.products[productID] = &CompleteProduct{
			ID:      product.ID,
			Product: product.Body,
			Plans:   make(map[string]*catalogModels.PlanBody),
		}
		productIDs[i] = productID
	}

	// Get plans for known productIDs
	planParams := catalogClientPlan.NewGetPlansParamsWithContext(ctx)
	planParams.SetProductID(productIDs)
	plans, err := cache.catalogClient.Plan.GetPlans(planParams)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the latest product plan data")
	}

	// Map plan catalog, so its searchable by id through hashmap on each product
	for _, plan := range plans.Payload {
		cache.products[plan.Body.ProductID.String()].Plans[plan.ID.String()] =
			plan.Body
	}

	// Get regions
	regions, err := cache.catalogClient.Region.GetRegions(nil)
	if err != nil {
		return nil, hierr.Errorf(err, "Failed to fetch the latest region data")
	}

	// Map region catalog, so its searchable by id through hashmap
	cache.regions = make(map[string]*catalogModels.RegionBody)
	for _, region := range regions.Payload {
		cache.regions[region.ID.String()] = region.Body
	}

	return cache, nil
}

// GenerateResourceCache builds a local in memory representation of a users
// Manifold resources
func GenerateResourceCache(ctx context.Context,
	cfg *config.Config, catalog *CatalogCache) (*ResourceCache, error) {

	cache := &ResourceCache{}

	// Initialize clients if needed
	var err error
	if cache.marketplaceClient == nil {
		cache.marketplaceClient, err = clients.NewMarketplace(cfg)
		if err != nil {
			return nil, hierr.Errorf(err,
				"Failed to create the Marketplace API client")
		}
	}

	// Get resources
	res, err := cache.marketplaceClient.Resource.GetResources(nil, nil)
	if err != nil {
		return nil, hierr.Errorf(err,
			"Failed to fetch the list of provisioned resources")
	}

	// Map products and operations to resources to create complete resource
	// entities
	cache.resources = make([]*CompleteResource, len(res.Payload))
	for i, res := range res.Payload {
		resource := CompleteResource{}
		resource.ID = res.ID
		resource.Resource = res.Body
		resource.Product = catalog.products[res.Body.ProductID.String()]
		resource.Plan = resource.Product.Plans[res.Body.PlanID.String()]
		resource.Region = catalog.regions[res.Body.RegionID.String()]
		cache.resources[i] = &resource
	}
	// Sort resources by name
	sort.Sort(resourcesSortByName(cache.resources))

	return cache, nil
}
