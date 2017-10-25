package templates

import (
	"fmt"
	"text/template"

	"github.com/manifoldco/promptui"
)

const (
	active   = `▸ {{ .Name | blue | bold }}{{ if .Title }} ({{ .Title }}){{end}}`
	inactive = `  {{ .Name | blue }}{{ if .Title }} ({{ .Title }}){{end}}`
	Selected = `{{ "✔" | green }} %s: {{ .Name | blue}}{{ if .Title }} ({{ .Title }}){{end}}`
	success  = `{{ "✔" | green }} {{ .Label }}: {{ .Value }}`
	failure  = `{{ "✗" | red }} {{ .Label }}: {{ .Value }}`
)

type input struct {
	Label string
	Value string
}

var (
	tplSuccess *template.Template
	tplFailure *template.Template
)

func init() {
	tplSuccess = template.Must(template.New("").Funcs(funcMap()).Parse(success))
	tplFailure = template.Must(template.New("").Funcs(funcMap()).Parse(failure))
}

var TplProvider = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(Selected, "Provider"),
}

var TplProduct = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   active,
	Inactive: inactive,
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
	Active:   active,
	Inactive: inactive,
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

var TplRegion = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   `▸ {{ .Name | blue | bold }} ({{ .Platform }}::{{ .Location }})`,
	Inactive: `  {{ .Name | blue }} ({{ .Platform }}::{{ .Location }})`,
	Selected: `{{"✔" | green }} Region: {{ .Name }} ({{ .Platform }}::{{ .Location }})`,
}

var TplResource = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   `▸ {{ if .Project }}{{ .Project | bold }}/{{ end }}{{ .Name | blue | bold }} ({{ .Title }})`,
	Inactive: `  {{ if .Project }}{{ .Project }}/{{ end }}{{ .Name | blue }} ({{ .Title }})`,
	Selected: `{{"✔" | green }} Resource: {{ if .Project }}{{ .Project }}/{{ end }}{{ .Name | blue }} ({{ .Title }})`,
}

var TplProject = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   active,
	Inactive: inactive,
	Selected: fmt.Sprintf(Selected, "Project"),
}

var TplTeam = &promptui.SelectTemplates{
	FuncMap:  funcMap(),
	Active:   active,
	Inactive: inactive,
	Selected: Selected, // Selected label can vary
}
