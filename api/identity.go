package api

import (
	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/generated/identity/client/invite"
	"github.com/manifoldco/manifold-cli/generated/identity/client/role"
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

func (api *API) Roles(teamID *manifold.ID) ([]models.RoleLabel, error) {
	params := role.NewGetRolesParamsWithContext(api.ctx)

	if teamID != nil {
		tid := teamID.String()
		params.SetTeamID(&tid)
	}

	res, err := api.Identity.Role.GetRoles(params, nil)
	if err != nil {
		return nil, err
	}

	payload := res.Payload

	return payload, nil
}
