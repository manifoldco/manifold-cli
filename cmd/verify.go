package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/identity/client/user"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
)

func init() {
	verifyCmd := cli.Command{
		Name:      "verify",
		ArgsUsage: "[code]",
		Usage:     "Verify an e-mail address with an e-mail verification code",
		Category:  "ADMINISTRATIVE",
		Action:    verifyEmailCode,
	}

	cmds = append(cmds, verifyCmd)
}

func verifyEmailCode(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	verificationCode, err := optionalArgCode(cliCtx, 0, "e-mail verification")
	if err != nil {
		return err
	}

	if verificationCode == "" {
		verificationCode, err = prompts.EmailVerificationCode("")
	}
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	if !s.Authenticated() {
		return errs.ErrMustLogin
	}

	client, err := api.New(api.Identity)
	if err != nil {
		return err
	}

	params := user.NewPostUsersVerifyParams()
	params.SetBody(&models.VerifyEmail{
		Body: &models.VerifyEmailBody{
			VerificationCode: &verificationCode,
		},
	})
	_, err = client.User.PostUsersVerify(params, nil)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to verify e-mail code: %s", err), -1)
	}

	analyticsClient, err := analytics.New(cfg, s)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("A problem occurred: %s", err), -1)
	}

	analyticsClient.Track(ctx, "E-mail Verified", nil)

	fmt.Printf("Thanks! Your e-mail address has been verified.\n")
	return nil
}
