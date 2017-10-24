package prompts

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	money "github.com/rhymond/go-money"
)

const (
	active   = `▸ {{.Body.Label | blue | bold }} ({{ .Body.Name }})`
	inactive = `  {{.Body.Label | blue }} ({{ .Body.Name }})`
	selected = `{{"✔" | green }} %s: {{.Body.Label | blue}} ({{ .Body.Name }})`
)

var funcMap template.FuncMap

func init() {
	funcMap = promptui.FuncMap
	funcMap["price"] = price
	funcMap["title"] = title
}

var ProductSelect = &promptui.SelectTemplates{
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(selected, "Product"),
	Details: `
Product:	{{.Body.Label | blue}} ({{ .Body.Label }})
Tagline:	{{ .Body.Tagline }}
Features:
{{- range $i, $el := .Body.ValueProps }}
{{- if lt $i 3 }}
 {{ $el.Header -}}
{{- end -}}
{{- end -}}`,
}

var PlanSelect = &promptui.SelectTemplates{
	FuncMap:  funcMap,
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(selected, "Plan"),
	Details: `
Plan:	{{.Body.Label | blue}} ({{ .Body.Label }})
Price:	{{ .Body.Cost | price }}
{{- range $i, $el := .Body.Features }}
{{- if lt $i 3 }}
{{ $el.Feature | title }}:	{{ $el.Value -}}
{{- end -}}
{{- end -}}`,
}

func price(value *int64) string {
	price := *value
	if price == 0 {
		return "Free"
	}
	return money.New(price, "USD").Display() + "/month"
}

func title(v interface{}) string {
	val := fmt.Sprintf("%v", v)
	return strings.Title(val)
}
