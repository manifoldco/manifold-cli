package clients

import (
	"context"

	"github.com/manifoldco/go-manifold"

	aClient "github.com/manifoldco/manifold-cli/generated/activity/client"
	"github.com/manifoldco/manifold-cli/generated/activity/client/event"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/go-manifold/events"
)

// FetchActivities returns the resources for the authenticated user
func FetchActivities(ctx context.Context, c *aClient.Activity, teamID *manifold.ID) ([]*events.Event, error) {

	res, err := c.Event.GetEvents(
		event.NewGetEventsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func FetchActivitiesWithProject(ctx context.Context, m *mClient.Marketplace, c *aClient.Activity, teamID *manifold.ID, projectLabel string) ([]*events.Event, error) {
	var project *mModels.Project

	if projectLabel != "" {
		var err error
		project, err = FetchProjectByLabel(ctx, m, teamID, projectLabel)
		if err != nil {
			return nil, err
		}
	}

	res, err := c.Event.GetEvents(
		event.NewGetEventsParamsWithContext(ctx), nil)
	if err != nil {
		return nil, err
	}

	var results []*events.Event
	for _, r := range res.Payload {
		if (teamID == nil && r.Body.RefID == nil) ||
			(teamID != nil && r.Body.RefID != nil && teamID.String() == r.Body.RefID().String()) {
			results = append(results, r)
		}
	}

	var matches []*events.Event
	for _, r := range results {
		id := r.Body.RefID()

		if project == nil || (&id != nil && project != nil && id == project.ID) {
			matches = append(matches, r)
		}
	}

	return matches, nil
}