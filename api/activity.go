package api

import (
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/events"
	"github.com/manifoldco/manifold-cli/generated/activity/client/event"
)

func (api *API) Events(scopeID manifold.ID) ([]*events.Event, error) {
	params := event.NewGetEventsParamsWithContext(api.ctx)
	id := scopeID.String()
	params.SetScopeID(&id)

	res, err := api.Activity.Event.GetEvents(params, nil)
	if err != nil {
		return nil, err
	}

	payload := res.Payload

	return payload, nil
}
