// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"html/template"
	"reflect"

	"github.com/AnUnnamedProject/i18n"
)

// eq returns a boolean of arg1 == arg2.
func eq(x, y interface{}) bool {
	return x == y
	// reflect.DeepEqual(x, y)
}

func ne(x, y interface{}) bool {
	return !eq(x, y)
}

// meta returns the metas string.
func meta(context map[string]interface{}) template.HTML {
	if context["FrameworkMeta"] == nil {
		return ""
	}

	str := ""
	metas := context["FrameworkMeta"].(map[string]string)

	for key, value := range metas {
		str += "<meta name=\"" + template.HTMLEscapeString(key) + "\" content=\"" + template.HTMLEscapeString(value) + "\">"
	}

	return template.HTML(str)
}

func addJS(url string) template.HTML {
	return template.HTML(`<script type="text/javascript" src="` + url + `"></script>`)
}

func addCSS(url string) template.HTML {
	return template.HTML(`<link rel="stylesheet" href="` + url + `">`)
}

func i18n_translate(v ...interface{}) template.HTML {
	context := v[len(v)-1].(map[string]interface{})
	args := v[:len(v)-1]
	str := args[0].(string)
	args = args[1:]

	return template.HTML(context["i18n"].(*i18n.I18N).Print(str, args...))
}

func i18n_plural(v ...interface{}) template.HTML {
	context := v[len(v)-1].(map[string]interface{})
	args := v[:len(v)-1]
	lang := ""

	if reflect.ValueOf(args[0]).Type().String() == "string" {
		lang = args[0].(string)
		args = args[1:]
	}

	value := args[0].(int)
	args = args[1:]

	if len(args) < 3 {
		return template.HTML("")
	}

	return template.HTML(context["i18n"].(*i18n.I18N).Plural(value, args[0].(string), args[1].(string), args[2].(string), lang))
}

// GetTemplateFuncs returs the FuncMap with framework custom functions.
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"eq":          eq,
		"ne":          ne,
		"meta":        meta,
		"addJS":       addJS,
		"addCSS":      addCSS,
		"i18n":        i18n_translate,
		"i18n_plural": i18n_plural,
	}
}
