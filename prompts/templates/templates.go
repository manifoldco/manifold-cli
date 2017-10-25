package templates

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

const (
	Active   = `▸ {{.Name | blue | bold }}{{ if .Title }} ({{ .Title }}){{end}}`
	Inactive = `  {{.Name | blue }}{{ if .Title }} ({{ .Title }}){{end}}`
	Selected = `{{"✔" | green }} %s: {{.Name | blue}}{{ if .Title }} ({{ .Title }}){{end}}`
)

var TplProvider = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: fmt.Sprintf(Selected, "Provider"),
}

var TplProduct = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: fmt.Sprintf(Selected, "Product"),
	Details: `
Product:	{{ .Name | blue}} ({{ .Title }})
Tagline:	{{ .Tagline }}
Features:
{{- range $i, $el := .Features }}
{{- if lt $i 3 }}
  {{ $el -}}
{{- end -}}
{{- end -}}`,
}

var TplPlan = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: fmt.Sprintf(Selected, "Plan"),
	Details: `
Plan:	{{ .Name | blue}} ({{ .Title }})
Price:	{{ .Cost | price }}
{{- range $i, $el := .Features }}
{{- if lt $i 3 }}
{{ $el.Name | title }}:	{{ $el.Description -}}
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
	Active:   `▸ {{ .Name | blue | bold }}{{ if .Title }} ({{ .Title }}){{end}}`,
	Inactive: `  {{ .Name | blue }}{{ if .Title }} ({{ .Title }}){{end}}`,
	Selected: fmt.Sprintf(Selected, "Project"),
}

var TplTeam = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   Active,
	Inactive: Inactive,
	Selected: Selected, // Selected label can vary
}
