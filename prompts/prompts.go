package prompts

import (
	"github.com/asaskevich/govalidator"

	"github.com/manifoldco/torus-cli/promptui"
)

// Email prompts the user to provide an email *or* accepted the default
// email value
func Email(defaultValue string) (string, error) {
	p := promptui.Prompt{
		Label: "Email",
		Validate: func(input string) error {
			valid := govalidator.IsEmail(input)
			if valid {
				return nil
			}

			return promptui.NewValidationError("Please enter a valid email address")
		},
	}

	if defaultValue != "" {
		p.Default = defaultValue
	}

	return p.Run()
}

// PasswordMask is the character used to mask password inputs
const PasswordMask = '●'

// Password prompts the user to input a password value
func Password() (string, error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  PasswordMask,
		Validate: func(input string) error {
			if len(input) < 8 {
				return promptui.NewValidationError(
					"Passwords must be greater than 8 characters")
			}

			return nil
		},
	}

	return prompt.Run()
}
