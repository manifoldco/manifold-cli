package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	hierr "github.com/reconquest/hierr-go"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"

	marketplaceClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	marketplaceModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

// ResourceCache represents a local in memory cache of a users resource data
type ResourceCache struct {
	marketplaceClient *marketplaceClient.Marketplace

	resources []*marketplaceModels.Resource
}

type resourcesSortByName []*marketplaceModels.Resource

func (r resourcesSortByName) Len() int {
	return len(r)
}
func (r resourcesSortByName) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r resourcesSortByName) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(fmt.Sprintf("%s", r[i].Body.Name)),
		fmt.Sprintf("%s", r[j].Body.Name)) < 0
}

// GenerateResourceCache builds a local in memory representation of a users
// Manifold resources
func GenerateResourceCache(ctx context.Context, cfg *config.Config,
	cache *ResourceCache) (*ResourceCache, error) {

	if cache == nil {
		cache = &ResourceCache{}
	}

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
	cache.resources = make([]*marketplaceModels.Resource, len(res.Payload))
	for i, res := range res.Payload {
		cache.resources[i] = res
	}
	// Sort resources by name
	sort.Sort(resourcesSortByName(cache.resources))

	return cache, nil
}
