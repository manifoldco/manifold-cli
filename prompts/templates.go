package prompts

import (
	"fmt"
	"text/template"

	"github.com/manifoldco/promptui"
	money "github.com/rhymond/go-money"
)

const (
	active   = `▸ {{.Body.Name | bold}} ({{ .Body.Label | blue }})`
	inactive = `  {{.Body.Name }} ({{ .Body.Label | blue }})`
	selected = `{{"✔" | green }} %s: {{.Body.Name | bold}} ({{ .Body.Label | blue }})`
)

var funcMap template.FuncMap

func init() {
	funcMap = promptui.FuncMap
	funcMap["price"] = price
}

var ProductSelect = &promptui.SelectTemplates{
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(selected, "Product"),
	Details: `
{{ .Body.Name | bold }}:
{{ .Body.Tagline }}
{{- range $i, $el := .Body.ValueProps }}
{{- if lt $i 3 }}
 - {{ $el.Header -}}
{{- end -}}
{{- end -}}`,
}

var PlanSelect = &promptui.SelectTemplates{
	FuncMap:  funcMap,
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(selected, "Plan"),
	Details: `
{{ .Body.Name | bold }} - {{ .Body.Cost | price }}:
{{- range $i, $el := .Body.Features }}
{{- if lt $i 3 }}
 - {{ $el.Feature }}: {{ $el.Value -}}
{{- end -}}
{{- end -}}`,
}

func price(value *int64) string {
	price := *value
	if price == 0 {
		return "Free"
	}
	return money.New(price, "USD").Display()
}
