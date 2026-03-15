package gen

import (
	"bytes"
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

// NeedsJSON returns true if any operation in this tag needs encoding/json.
func (d TemplateData) NeedsJSON() bool {
	for _, op := range d.Operations {
		if op.HasBody() {
			return true
		}
	}
	return false
}

// NeedsFmt returns true if any operation needs fmt (body error formatting).
func (d TemplateData) NeedsFmt() bool {
	return d.NeedsJSON()
}

// NeedsStrings returns true if any operation has path or query params.
func (d TemplateData) NeedsStrings() bool {
	for _, op := range d.Operations {
		if len(op.PathParams()) > 0 || len(op.QueryParams()) > 0 {
			return true
		}
	}
	return false
}

// RootTemplateData is passed to root.go.tmpl.
type RootTemplateData struct {
	ByTag map[string][]*Operation
}

var tmplFuncs = template.FuncMap{
	"needsNilBody": func(method string) bool {
		m := strings.ToUpper(method)
		return m == "POST" || m == "PATCH" || m == "PUT"
	},
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

// RenderCommandFileBytes renders command.go.tmpl and returns the result as bytes.
func RenderCommandFileBytes(tag string, ops []*Operation) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderCommandFile(&buf, tag, ops); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderRootFile renders root.go.tmpl into w.
func RenderRootFile(w io.Writer, byTag map[string][]*Operation) error {
	tmpl, err := loadTemplate("templates/root.go.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, RootTemplateData{ByTag: byTag})
}

// RenderRootFileBytes renders root.go.tmpl and returns the result as bytes.
func RenderRootFileBytes(byTag map[string][]*Operation) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderRootFile(&buf, byTag); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
