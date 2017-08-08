package main

import (
	"reflect"
	"strings"

	"github.com/urfave/cli"
	"gopkg.in/oleiade/reflections.v1"

	"github.com/manifoldco/manifold-cli/dirprefs"
)

// chain allows easy sequential calling of BeforeFuncs and AfterFuncs.
// chain will exit on the first error seen.
func chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {

		for _, f := range funcs {
			err := f(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// loadDirPrefs loads argument values from the .torus.json file
func loadDirPrefs(ctx *cli.Context) error {
	d, err := dirprefs.Load(true)
	if err != nil {
		return err
	}

	return reflectArgs(ctx, d, "json")
}

func reflectArgs(ctx *cli.Context, i interface{}, tagName string) error {
	// tagged field names match the argument names
	tags, err := reflections.Tags(i, tagName)
	if err != nil {
		return err
	}

	flags := make(map[string]bool)
	for _, flagName := range ctx.FlagNames() {
		// This value is already set via arguments or env vars. skip it.
		if isSet(ctx, flagName) {
			continue
		}

		flags[flagName] = true
	}

	for fieldName, tag := range tags {
		name := strings.SplitN(tag, ",", 2)[0] // remove omitempty if its there
		if _, ok := flags[name]; ok {
			field, err := reflections.GetField(i, fieldName)
			if err != nil {
				return err
			}

			if f, ok := field.(string); ok && f != "" {
				ctx.Set(name, field.(string))
			}
		}
	}

	return nil
}

func isSet(ctx *cli.Context, name string) bool {
	value := ctx.Generic(name)
	if value != nil {
		v := reflect.Indirect(reflect.ValueOf(value))
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			return v.Len() != 0
		}

		return true
	}

	return false
}
