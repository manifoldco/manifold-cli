package prompts

import "github.com/manifoldco/promptui"

var ProductSelect = &promptui.SelectTemplates{
	Active:   `▸ {{.Body.Name | bold}} ({{ .Body.Label | blue }})`,
	Inactive: `  {{.Body.Name }} ({{ .Body.Label | blue }})`,
	Details: `
{{ .Body.Name | bold }}:
{{ .Body.Tagline }}
{{- range $i, $el := .Body.ValueProps }}
{{- if lt $i 3 }}
- {{ $el.Header -}}
{{- end -}}
{{- end -}}`,
}
