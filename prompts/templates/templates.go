package templates

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

const (
	Active   = `▸ {{.Name | blue | bold }}{{ if .Title }} ({{ .Title }}){{end}}`
	Inactive = `  {{.Name | blue }}{{ if .Title }} ({{ .Title }}){{end}}`
	Selected = `{{"✔" | green }} %s: {{.Name | blue}}{{ if .Title }} ({{ .Title }}){{end}}`

	// TODO: remove legacy format
	active   = `▸ {{.Body.Label | blue | bold }} ({{ .Body.Name }})`
	inactive = `  {{.Body.Label | blue }} ({{ .Body.Name }})`
	selected = `{{"✔" | green }} %s: {{.Body.Label | blue}} ({{ .Body.Name }})`
)

var TplProvider = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: fmt.Sprintf(Selected, "Provider"),
}

var TplProduct = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
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

var TplPlan = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
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

var TplResource = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   `▸ {{ if .Project }}{{ .Project | bold }}/{{end}}{{ .Name | blue | bold }} ({{ .Title }})`,
	Inactive: `  {{ if .Project }}{{ .Project }}/{{end}}{{ .Name | blue }} ({{ .Title }})`,
	Selected: `{{"✔" | green }} Resource: {{ if .Project }}{{ .Project }}/{{end}}{{ .Name | blue }} ({{ .Title }})`,
}

var TplProject = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   `▸ {{.Name | blue | bold }}{{ if .Title }} ({{ .Title }}){{end}}`,
	Inactive: `  {{.Name | blue }}{{if .Title }} ({{ .Title }}){{end}}`,
	Selected: fmt.Sprintf(Selected, "Project"),
}

var TplTeam = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: Selected, // Selected label can vary
}
