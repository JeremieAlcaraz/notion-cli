package gen

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateData is passed to command.go.tmpl for a single tag group.
type TemplateData struct {
	Tag        string
	Operations []*Operation
}

// RootTemplateData is passed to root.go.tmpl.
type RootTemplateData struct {
	ByTag map[string][]*Operation
}

var tmplFuncs = template.FuncMap{
	"pascal": toPascal,
	"kebab":  toKebab,
	"tagSlug": func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	},
	"httpMethod": func(method string) string {
		switch strings.ToUpper(method) {
		case "GET":
			return "Get"
		case "POST":
			return "Post"
		case "PATCH":
			return "Patch"
		case "DELETE":
			return "Delete"
		case "PUT":
			return "Put"
		default:
			return "Get"
		}
	},
	"not": func(b bool) bool { return !b },
	"gt": func(a, b int) bool { return a > b },
}

// RenderCommandFile renders command.go.tmpl for a single tag into w.
func RenderCommandFile(w io.Writer, tag string, ops []*Operation) error {
	tmpl, err := loadTemplate("templates/command.go.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, TemplateData{Tag: tag, Operations: ops})
}

// RenderRootFile renders root.go.tmpl into w.
func RenderRootFile(w io.Writer, byTag map[string][]*Operation) error {
	tmpl, err := loadTemplate("templates/root.go.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, RootTemplateData{ByTag: byTag})
}

func loadTemplate(name string) (*template.Template, error) {
	src, err := templateFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", name, err)
	}
	return template.New(name).Funcs(tmplFuncs).Parse(string(src))
}

// toPascal converts kebab-case or snake_case to PascalCase.
// "retrieve-a-page" → "RetrieveAPage"
func toPascal(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '-' || r == '_' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// toKebab converts snake_case to kebab-case.
// "page_id" → "page-id"
func toKebab(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}
