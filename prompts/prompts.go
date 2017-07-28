package prompts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/torus-cli/promptui"
	"github.com/rhymond/go-money"

	"github.com/manifoldco/manifold-cli/errs"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
)

const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"

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
func SelectPlan(plans []*cModels.Plan, label string) (int, string, error) {
	line := func(p *cModels.Plan) string {
		return fmt.Sprintf("%s (%s) - %s", p.Body.Name, p.Body.Label, getPlanCost(p))
	}

	var idx int
	if label != "" {
		found := false
		for i, p := range plans {
			if string(p.Body.Label) == label {
				idx = i
				found = true
				break
			}
		}

		p := plans[idx]
		if !found {
			fmt.Println(promptui.FailedValue("Plan", label))
			return 0, "", errs.ErrPlanNotFound
		}

		fmt.Println(promptui.SuccessfulValue("Plan", line(p)))
		return idx, label, nil
	}

	sort.Sort(plansSortByCost(plans))
	labels := make([]string, len(plans))
	for i, p := range plans {
		labels[i] = line(p)
	}

	prompt := promptui.Select{
		Label: "Select Plan",
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

// SelectCreateAppName prompts the user to select or create an application
// name, a -1 idx will be returned if the app name requires creation.
func SelectCreateAppName(names []string, label string) (int, string, error) {
	labels := make([]string, len(names))
	for i, n := range names {
		labels[i] = n

		if label != "" && labels[i] == label {
			fmt.Println(promptui.SuccessfulValue("App Name", label))
			return i, label, nil
		}
	}

	if label != "" {
		prompt := promptui.Prompt{
			Label:   "App Name",
			Default: label,
		}

		value, err := prompt.Run()
		return -1, value, err
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
func ResourceName(defaultValue string) (string, error) {
	p := promptui.Prompt{
		Label:   "Resource Name (one will be generated if left blank)",
		Default: defaultValue,
		Validate: func(input string) error {
			if len(input) == 0 {
				return nil
			}

			n := manifold.Name(input)
			if err := n.Validate(nil); err != nil {
				return promptui.NewValidationError("Please provide a valid App Name")
			}

			return nil
		},
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
