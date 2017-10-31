package prompts

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/promptui"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/token"

	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts/templates"

	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

const (
	namePattern   = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"
	couponPattern = "^[0-9A-Z]{1,128}$"
	codePattern   = "^[0-9abcdefghjkmnpqrtuvwxyz]{16}$"
)

// NumberMask is the character used to mask number inputs
const NumberMask = '#'

var errBad = errors.New("Bad Value")

// Title prompts the user to provide a Title value
func Title(field, defaultValue string, autoSelect, allowEmpty bool) (string, error) {
	field = strings.Title(field)

	validate := func(input string) error {
		if allowEmpty && len(input) == 0 {
			return nil
		}

		t := manifold.Name(input)
		if err := t.Validate(nil); err != nil {
			return fmt.Errorf("Please provide a valid %s title", field)
		}

		return nil
	}

	label := fmt.Sprintf("New %s Title", field)

	if autoSelect {
		err := validate(defaultValue)
		if err != nil {
			fmt.Println(templates.PromptFailure(label, defaultValue))
		} else {
			fmt.Println(templates.PromptSuccess(label, defaultValue))
		}
		return defaultValue, err
	}

	p := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	return p.Run()
}

// Name prompts the user to provide a Name value
func Name(field, defaultValue string, autoSelect, allowEmpty bool) (string, error) {
	field = strings.Title(field)

	validate := func(input string) error {
		if len(input) == 0 {
			return fmt.Errorf("Please provide a %s name", field)
		}

		l := manifold.Label(input)
		if err := l.Validate(nil); err != nil {
			return fmt.Errorf("Please provide a valid %s name", field)
		}

		return nil
	}

	label := fmt.Sprintf("New %s Name", field)

	if autoSelect {
		err := validate(defaultValue)
		if err != nil {
			fmt.Println(templates.PromptFailure(label, defaultValue))
		} else {
			fmt.Println(templates.PromptSuccess(label, defaultValue))
		}

		return defaultValue, err
	}

	p := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	return p.Run()
}

// ResourceTitle prompts the user to provide a resource title or to accept empty
// to let the system generate one.
func ResourceTitle(defaultValue string, autoSelect bool) (string, error) {
	return Title("resource", defaultValue, autoSelect, true)
}

// ResourceName prompts the user to provide a label name
func ResourceName(defaultValue string, autoSelect bool) (string, error) {
	return Name("resource", defaultValue, autoSelect, false)
}

// TeamTitle prompts the user to enter a new Team title
func TeamTitle(defaultValue string, autoSelect bool) (string, error) {
	return Title("team", defaultValue, autoSelect, false)
}

// ProjectTitle prompts the user to enter a new project title
func ProjectTitle(defaultValue string, autoSelect bool) (string, error) {
	return Title("project", defaultValue, autoSelect, false)
}

// TokenDescription prompts the user to enter a token description
func TokenDescription() (string, error) {
	p := promptui.Prompt{
		Label:   "Token Description",
		Default: "",
	}
	return p.Run()
}

// ProjectDescription prompts the user to enter a project description
func ProjectDescription(defaultValue string, autoSelect bool) (string, error) {
	label := "Project Description"

	if autoSelect && defaultValue != "" {
		fmt.Println(templates.PromptSuccess(label, defaultValue))
		return defaultValue, nil
	}

	p := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	return p.Run()
}

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

			return errors.New("Please enter a valid email address")
		},
	}

	if defaultValue != "" {
		p.Default = defaultValue
	}

	return p.Run()
}

// FullName prompts the user to input a person's name
func FullName(defaultValue string) (string, error) {
	p := promptui.Prompt{
		Label: "Name",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, namePattern) {
				return nil
			}
			return errors.New("Please enter a valid name")
		},
	}
	if defaultValue != "" {
		p.Default = defaultValue
	}

	return p.Run()
}

// CouponCode prompts the user to input an alphanumeric coupon code.
func CouponCode() (string, error) {
	p := promptui.Prompt{
		Label: "Code",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, couponPattern) {
				return nil
			}
			return errors.New("Please enter a valid code")
		},
	}

	return p.Run()
}

// EmailVerificationCode prompts the user to input a person's name
func EmailVerificationCode(defaultValue string) (string, error) {
	p := promptui.Prompt{
		Label: "E-mail Verification Code",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, codePattern) {
				return nil
			}
			return errors.New("Please enter a valid e-mail verification code")
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
				return errors.New("Passwords must be greater than 8 characters")
			}

			return nil
		},
	}

	return prompt.Run()
}

// Confirm is a confirmation prompt
func Confirm(msg string) (string, error) {
	p := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
	}

	return p.Run()
}

// HandleSelectError returns a cli error if the error is not an EOF or
// Interrupt
func HandleSelectError(err error, generic string) error {
	if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
		return err
	}

	return errs.NewErrorExitError(generic, err)
}

func isCard(raw string) error {
	if govalidator.StringLength(raw, "12", "19") && govalidator.IsNumeric(raw) {
		return nil
	}

	return errBad
}

func isExpiry(raw string) error {
	if govalidator.StringLength(raw, "5", "5") {
		return nil
	}

	return errBad
}

func isCVV(raw string) error {
	if govalidator.StringLength(raw, "3", "4") && govalidator.IsNumeric(raw) {
		return nil
	}

	return errBad
}

// CreditCard handles receiving and tokenizing payment information
func CreditCard() (*stripe.Token, error) {
	rCrd, err := (&promptui.Prompt{
		Label:    "💳  Card Number",
		Validate: isCard,
	}).Run()
	if err != nil {
		return nil, err
	}

	rExp, err := (&promptui.Prompt{
		Label:    "📅  Expiry (MM/YY)",
		Validate: isExpiry,
	}).Run()
	if err != nil {
		return nil, err
	}

	rCVV, err := (&promptui.Prompt{
		Label:    "🔒  CVV",
		Mask:     NumberMask,
		Validate: isCVV,
	}).Run()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(rExp, "/")
	year, month := "20"+parts[1], parts[0]
	tkn, err := token.New(&stripe.TokenParams{Card: &stripe.CardParams{
		Number: rCrd,
		Month:  month,
		Year:   year,
		CVC:    rCVV,
	}})
	if err != nil {
		return nil, errs.NewStripeError(err)
	}

	return tkn, nil
}

var configKeyRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,1000}$`)

// CredentialAlias asks the user to supply a new alias for the given cred
func CredentialAlias(creds []*mModels.Credential) (*mModels.Credential, string, string, error) {
	c, originalName, err := SelectCredential(creds)
	if err != nil {
		return nil, "", "", err
	}

	fmt.Printf("Creating alias for `%s`\n", originalName)

	p := promptui.Prompt{
		Label: "Alias Name",
		Validate: func(input string) error {
			if !configKeyRegexp.MatchString(input) {
				return errors.New("Please provide a valid alias")
			}
			return nil
		},
	}
	newName, err := p.Run()
	if err != nil {
		return nil, "", "", err
	}

	return c, originalName, newName, nil
}
