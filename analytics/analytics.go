// Package analytics is a tiny package that makes it easy to send an analytic
// events to the server for tracking purposes.
package analytics

import (
	"context"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
)

// Analytics represents a struct for submitting analytic tracking events to the
// server.
type Analytics struct {
	session session.Session
	cfg     *config.Config
	client  *client.Identity
}

// Track submits an event to the identity service
func (a *Analytics) Track(ctx context.Context, name string, params *map[string]string) error {
	if !a.session.Authenticated() {
		return errs.ErrMustLogin
	}

	// If analytics isn't enabled, don't tell anyone!
	if !a.cfg.Analytics {
		return nil
	}

	n := name
	additionalProps := make(map[string]interface{})
	additionalProps["version"] = config.Version

	if params != nil {
		for k, v := range *params {
			additionalProps[k] = v
		}
	}

	p := user.NewPostAnalyticsParams()
	p.SetContext(ctx)
	p.SetBody(&models.AnalyticsEvent{
		EventName: &n,
		UserID:    a.session.User().ID,
		Properties: &models.AnalyticsEventProperties{
			Platform: "cli",
			AnalyticsEventPropertiesAdditionalProperties: additionalProps,
		},
	})

	_, err := a.client.User.PostAnalytics(p)
	if err != nil {
		switch e := err.(type) {
		case *user.PostAnalyticsBadRequest:
			return e.Payload
		case *user.PostAnalyticsUnauthorized:
			return e.Payload
		case *user.PostAnalyticsInternalServerError:
			return e.Payload
		default:
			return e
		}
	}

	return nil
}

// New returns a new instance of Analytics
func New(cfg *config.Config, s session.Session) (*Analytics, error) {
	client, err := clients.NewIdentity(cfg)
	if err != nil {
		return nil, err
	}

	return &Analytics{
		session: s,
		cfg:     cfg,
		client:  client,
	}, nil
}
