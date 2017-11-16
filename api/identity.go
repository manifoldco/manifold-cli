package api

import (
	"github.com/manifoldco/manifold-cli/generated/identity/client/invite"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
)

// AcceptInvite accepts an invitation to join a team. A nil error means the
// invitation was accepted correctly.
func (api *API) AcceptInvite(token string) error {
	params := invite.NewPostInvitesAcceptParamsWithContext(api.ctx)
	t := models.LimitedLifeTokenBase32(token)
	params.SetBody(&models.AcceptInvite{Token: t})

	_, err := api.Identity.Invite.PostInvitesAccept(params, nil)
	return err
}
