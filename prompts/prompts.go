package prompts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/torus-cli/promptui"
	"github.com/rhymond/go-money"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/token"

	"github.com/manifoldco/manifold-cli/errs"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	iModels "github.com/manifoldco/manifold-cli/generated/identity/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"

// NumberMask is the character used to mask number inputs
const NumberMask = '#'

var errBad = promptui.NewValidationError("Bad Value")

type plansSortByCost []*cModels.Plan

func (p plansSortByCost) Len() int {
	return len(p)
}

func (p plansSortByCost) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p plansSortByCost) Less(i, j int) bool {
	return *p[i].Body.Cost < *p[j].Body.Cost
}

type productsSortByName []*cModels.Product

func (p productsSortByName) Len() int {
	return len(p)
}

func (p productsSortByName) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p productsSortByName) Less(i, j int) bool {
	return strings.Compare(
		strings.ToLower(fmt.Sprintf("%s", p[i].Body.Name)),
		strings.ToLower(fmt.Sprintf("%s", p[j].Body.Name)),
	) < 0
}

// SelectProduct prompts the user to select a product from the given list.
func SelectProduct(products []*cModels.Product, label string) (int, string, error) {
	line := func(p *cModels.Product) string {
		return fmt.Sprintf("%s (%s)", p.Body.Name, p.Body.Label)
	}

	var idx int
	if label != "" {
		found := false
		for i, p := range products {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Product", label))
			return 0, "", errs.ErrProductNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Product", line(products[idx])))
		return idx, label, nil
	}

	sort.Sort(productsSortByName(products))
	labels := make([]string, len(products))
	for i, p := range products {
		labels[i] = line(p)
	}

	prompt := promptui.Select{
		Label: "Select Product",
		Items: labels,
	}

	return prompt.Run()
}

// SelectPlan prompts the user to select a plan from the given list.
func SelectPlan(plans []*cModels.Plan, label string, filterLabelTop bool) (int, string, error) {
	line := func(p *cModels.Plan) string {
		return fmt.Sprintf("%s (%s) - %s", p.Body.Name, p.Body.Label, getPlanCost(p))
	}

	var idx int
	var fp *cModels.Plan
	if label != "" {
		found := false
		for i, p := range plans {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		fp = plans[idx]
		if !found {
			fmt.Println(promptui.FailedValue("Plan", label))
			return 0, "", errs.ErrPlanNotFound
		}

		if !filterLabelTop {
			fmt.Println(promptui.SuccessfulValue("Plan", line(fp)))
			return idx, label, nil
		}
	}

	sort.Sort(plansSortByCost(plans))
	labels := make([]string, len(plans))

	var selectedIdx int
	for i, p := range plans {
		labels[i] = line(p)
		if p == fp {
			selectedIdx = i
		}
	}

	if filterLabelTop {
		labels[0], labels[selectedIdx] = labels[selectedIdx], labels[0]
	}

	prompt := promptui.Select{
		Label: "Select Plan",
		Items: labels,
	}

	return prompt.Run()
}

// SelectResource promps the user to select a provisioned resource from the given list
func SelectResource(resources []*mModels.Resource, label string) (int, string, error) {
	line := func(r *mModels.Resource) string {
		return fmt.Sprintf("%s (%s) %s", r.Body.Name, r.Body.Label, r.Body.AppName)
	}

	var idx int
	if label != "" {
		found := false
		for i, p := range resources {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		r := resources[idx]
		if !found {
			fmt.Println(promptui.FailedValue("Resource", label))
			return 0, "", errs.ErrResourceNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Resource", line(r)))
		return idx, label, nil
	}

	labels := make([]string, len(resources))
	for i, r := range resources {
		labels[i] = line(r)
	}

	prompt := promptui.Select{
		Label: "Select Resource",
		Items: labels,
	}

	return prompt.Run()
}

// SelectRegion prompts the user to select a region from the given list.
func SelectRegion(regions []*cModels.Region) (int, string, error) {
	line := func(r *cModels.Region) string {
		return fmt.Sprintf("%s (%s::%s)", r.Body.Name, *r.Body.Platform, *r.Body.Location)
	}

	labels := make([]string, len(regions))
	for i, r := range regions {
		labels[i] = line(r)
	}

	// TODO: Build "auto" resolve into promptui in case of only one item
	if len(regions) == 1 {
		fmt.Println(promptui.SuccessfulValue("Region", line(regions[0])))
		return 0, string(regions[0].Body.Name), nil
	}

	prompt := promptui.Select{
		Label: "Select Region",
		Items: labels,
	}

	return prompt.Run()
}

// SelectTeam prompts the user to select a team from the given list. -1 as the first return value
// indicates no team has been selected
func SelectTeam(teams []*iModels.Team, label string, includeNoTeam bool) (int, string, error) {
	line := func(t *iModels.Team) string {
		return fmt.Sprintf("%s (%s)", t.Body.Name, t.Body.Label)
	}

	var idx int
	if label != "" {
		found := false
		for i, t := range teams {
			if string(t.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		t := teams[idx]

		if !found {
			fmt.Println(promptui.FailedValue("Team", label))
			return 0, "", errs.ErrTeamNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Team", line(t)))
		return idx, label, nil
	}

	labels := make([]string, len(teams)+1)
	for i, t := range teams {
		labels[i] = line(t)
	}

	if includeNoTeam {
		labels = append([]string{"Don't use a team"}, labels...)
	}

	prompt := promptui.Select{
		Label: "Select Team",
		Items: labels,
	}

	teamIdx, name, err := prompt.Run()

	if includeNoTeam {
		return teamIdx - 1, name, err
	}

	return teamIdx, name, err
}

// SelectCreateAppName prompts the user to select or create an application
// name, a -1 idx will be returned if the app name requires creation.
func SelectCreateAppName(names []string, label string, filterToTop bool) (int, string, error) {
	labels := make([]string, len(names))

	var idx int
	for i, n := range names {
		labels[i] = n

		if label != "" && labels[i] == label {
			if !filterToTop {
				fmt.Println(promptui.SuccessfulValue("App Name", label))
				return i, label, nil
			}

			idx = i
		}
	}

	if label != "" && !filterToTop {
		prompt := promptui.Prompt{
			Label:   "App Name",
			Default: label,
		}

		value, err := prompt.Run()
		return -1, value, err
	}

	if filterToTop {
		labels[0], labels[idx] = labels[idx], labels[0]
	}

	prompt := promptui.SelectWithAdd{
		Label:    "App Name",
		Items:    labels,
		AddLabel: "Create a New App",
		Validate: func(input string) error {
			n := manifold.Name(input)
			if err := n.Validate(nil); err != nil {
				return promptui.NewValidationError("Please provide a valid App Name")
			}

			return nil
		},
	}

	return prompt.Run()
}

// ResourceName prompts the user to provide a resource name or to accept empty
// to let the system generate one.
func ResourceName(defaultValue string, autoSelect bool) (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return nil
		}

		n := manifold.Name(input)
		if err := n.Validate(nil); err != nil {
			return promptui.NewValidationError("Please provide a valid App Name")
		}

		return nil
	}

	label := "Resource Name (one will be generated if left blank)"

	if autoSelect {
		err := validate(defaultValue)
		if err != nil {
			fmt.Println(promptui.FailedValue(label, defaultValue))
		} else {
			fmt.Println(promptui.SuccessfulValue(label, defaultValue))
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

// ResourceLabel prompts the user to provide a label name
func ResourceLabel(defaultValue string, autoSelect bool) (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return promptui.NewValidationError("Please provide a resource label")
		}

		l := manifold.Label(input)
		if err := l.Validate(nil); err != nil {
			return promptui.NewValidationError("Please provide a valid resource label")
		}

		return nil
	}

	label := "Resource Label"

	if autoSelect {
		err := validate(defaultValue)
		if err != nil {
			fmt.Println(promptui.FailedValue(label, defaultValue))
		} else {
			fmt.Println(promptui.SuccessfulValue(label, defaultValue))
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

// TeamName prompts the user to enter a new Team name
func TeamName(defaultValue string, autoSelect bool) (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return promptui.NewValidationError("Please provide a valid team name")
		}

		l := manifold.Name(input)
		if err := l.Validate(nil); err != nil {
			return promptui.NewValidationError("Please provide a valid team name")
		}

		return nil
	}

	label := "Team Name"

	if autoSelect {
		err := validate(defaultValue)
		if err != nil {
			fmt.Println(promptui.FailedValue(label, defaultValue))
		} else {
			fmt.Println(promptui.SuccessfulValue(label, defaultValue))
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

// FullName prompts the user to input a person's name
func FullName(defaultValue string) (string, error) {
	p := promptui.Prompt{
		Label: "Name",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, namePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid name")
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

func getPlanCost(p *cModels.Plan) string {
	if p.Body.Cost == nil {
		return "Free!"
	}

	c := *p.Body.Cost
	if c == 0 {
		return "Free!"
	}

	return money.New(c, "USD").Display()
}

func isCard(raw string) error {
	if govalidator.StringLength(raw, "16", "16") && govalidator.IsNumeric(raw) {
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
	if govalidator.StringLength(raw, "3", "3") && govalidator.IsNumeric(raw) {
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
