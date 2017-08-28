package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/billing/client/profile"
	bModels "github.com/manifoldco/manifold-cli/generated/billing/models"
)

func init() {
	billingCmd := cli.Command{
		Name:  "billing",
		Usage: "Manage your billing information",
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "Add a credit card",
				Action: addBillingProfile,
			},
			{
				Name:   "update",
				Usage:  "Change the credit card on file",
				Action: updateBillingProfile,
			},
		},
	}

	cmds = append(cmds, billingCmd)
}

func addBillingProfile(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	userID, token, err := inputAndTokenize(ctx, cfg)
	if err != nil {
		return err
	}

	bClient, err := clients.NewBilling(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Billing API client: "+
			err.Error(), -1)
	}

	p := profile.NewPostProfilesParamsWithContext(ctx)
	p.SetBody(&bModels.ProfileCreateRequest{
		Token:  token,
		UserID: userID,
	})

	_, err = bClient.Profile.PostProfiles(p, nil)
	if err != nil {
		return cli.NewExitError("Failed to add billing profile: "+err.Error(), -1)
	}

	fmt.Println("Your billing info has been saved.")
	return nil
}

func updateBillingProfile(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	userID, token, err := inputAndTokenize(ctx, cfg)
	if err != nil {
		return err
	}

	bClient, err := clients.NewBilling(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Billing API client: "+
			err.Error(), -1)
	}

	p := profile.NewPatchProfilesIDParamsWithContext(ctx)
	p.SetID(userID.String())
	p.SetBody(&bModels.ProfileUpdateRequest{
		Token: token,
	})

	_, err = bClient.Profile.PatchProfilesID(p, nil)
	if err != nil {
		return cli.NewExitError("Failed to update billing profile: "+err.Error(), -1)
	}

	fmt.Println("Your billing info has been saved.")
	return nil
}

func inputAndTokenize(ctx context.Context, cfg *config.Config) (*manifold.ID, *string, error) {
	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return nil, nil, cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	if !s.Authenticated() {
		return nil, nil, errs.ErrMustLogin
	}

	token, err := prompts.CreditCard()
	if err != nil {
		return nil, nil, cli.NewExitError("Failed to tokenize credit card: "+err.Error(), -1)
	}

	return &s.User().ID, &token.ID, nil
}
