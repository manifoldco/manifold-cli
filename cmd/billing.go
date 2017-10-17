package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/billing/client/discount"
	"github.com/manifoldco/manifold-cli/generated/billing/client/profile"
	bModels "github.com/manifoldco/manifold-cli/generated/billing/models"
)

func init() {
	billingCmd := cli.Command{
		Name:     "billing",
		Usage:    "Manage your billing information",
		Category: "ADMINISTRATIVE",
		Subcommands: []cli.Command{
			{
				Name:  "add",
				Usage: "Add a credit card",
				Flags: teamFlags,
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, addBillingProfileCmd),
			},
			{
				Name:  "update",
				Usage: "Change the credit card on file",
				Flags: teamFlags,
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, updateBillingProfileCmd),
			},
			{
				Name:      "redeem",
				Usage:     "Redeem a coupon code",
				Flags:     teamFlags,
				ArgsUsage: "[code]",
				Action: middleware.Chain(middleware.EnsureSession,
					middleware.LoadTeamPrefs, redeemCouponCmd),
			},
		},
	}

	cmds = append(cmds, billingCmd)
}

func addBillingProfileCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	token, err := creditCardInput(ctx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Billing)
	if err != nil {
		return err
	}

	params := profile.NewPostProfilesParamsWithContext(ctx)

	if teamID == nil {
		params.SetBody(&bModels.ProfileCreateRequest{
			Token:  &token,
			UserID: userID,
		})
	} else {
		params.SetBody(&bModels.ProfileCreateRequest{
			Token:  &token,
			TeamID: teamID,
		})
	}

	spin := prompts.NewSpinner("Creating billing profile")
	spin.Start()
	defer spin.Stop()
	_, err = client.Billing.Profile.PostProfiles(params, nil)
	if err != nil {
		return cli.NewExitError("Failed to add billing profile: "+err.Error(), -1)
	}

	fmt.Println("Your billing info has been saved.")
	return nil
}

func updateBillingProfileCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	token, err := creditCardInput(ctx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Billing)
	if err != nil {
		return err
	}

	params := profile.NewPatchProfilesIDParamsWithContext(ctx)

	if teamID == nil {
		params.SetID(userID.String())
	} else {
		params.SetID(teamID.String())
	}

	params.SetBody(&bModels.ProfileUpdateRequest{
		Token: &token,
	})

	spin := prompts.NewSpinner("Updating billing profile")
	spin.Start()
	defer spin.Stop()
	_, err = client.Billing.Profile.PatchProfilesID(params, nil)
	if err != nil {
		return cli.NewExitError("Failed to update billing profile: "+err.Error(), -1)
	}

	fmt.Println("Your billing info has been saved.")
	return nil
}

func redeemCouponCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	var err error
	code := cliCtx.Args().First()

	if code == "" {
		code, err = prompts.CouponCode()
		if err != nil {
			return err
		}
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Billing)
	if err != nil {
		return err
	}

	body := &bModels.DiscountCreateRequest{
		Code: bModels.CouponCode(code),
	}

	me := cliCtx.Bool("me")

	if !me && teamID != nil {
		body.TeamID = teamID
	}

	spin := prompts.NewSpinner("Applying coupon")
	spin.Start()
	defer spin.Stop()
	params := discount.NewPostDiscountsParamsWithContext(ctx)
	params.SetBody(body)

	_, err = client.Billing.Discount.PostDiscounts(params, nil)

	if err != nil {
		switch e := err.(type) {
		case *discount.PostDiscountsBadRequest:
			msg := strings.Join(e.Payload.Messages, "\n")
			return cli.NewExitError("Failed to redeem coupon code: "+msg, -1)
		case *discount.PostDiscountsConflict:
			return cli.NewExitError("This coupon code has been used before.", -1)
		case *discount.PostDiscountsInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return cli.NewExitError("Failed to redeem coupon code: "+err.Error(), -1)
		}
	}

	fmt.Println("Coupon credit has been applied to your account.")
	return nil
}

// loadUserID retrieves the user id based on the session.
func loadUserID(ctx context.Context) (*manifold.ID, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return nil, cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	return &s.User().ID, nil
}

// creditCardInput prompts for the user credit card information and returns
// the user id and the Stripe payment source token.
func creditCardInput(ctx context.Context) (string, error) {
	token, err := prompts.CreditCard()
	if err != nil {
		return "", cli.NewExitError("Failed to tokenize credit card: "+err.Error(), -1)
	}

	return token.ID, nil
}
