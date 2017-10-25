package templates

import (
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
