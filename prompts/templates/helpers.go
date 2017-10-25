package templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	money "github.com/rhymond/go-money"
)

func funcMap() template.FuncMap {
	fn := promptui.FuncMap
	fn["price"] = price
	fn["title"] = title
	return fn
}

func price(price int) string {
	if price == 0 {
		return "Free"
	}
	return money.New(int64(price), "USD").Display() + "/month"
}

func title(v interface{}) string {
	val := fmt.Sprintf("%v", v)
	return strings.Title(val)
}

func PromptSuccess(tpls *promptui.SelectTemplates, data interface{}) string {
	tpl, err := template.New("").Funcs(funcMap()).Parse(tpls.Selected)
	if err != nil {
		return fmt.Sprintf("%+v", data)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		return fmt.Sprintf("%+v", data)
	}
	return buf.String()
}

func PromptFailure(label, data interface{}) string {
	txt := fmt.Sprintf(failure, label)
	tpl, err := template.New("").Funcs(funcMap()).Parse(txt)
	if err != nil {
		return fmt.Sprintf("%s: %+v", label, data)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		return fmt.Sprintf("%+v", data)
	}
	return buf.String()
}
