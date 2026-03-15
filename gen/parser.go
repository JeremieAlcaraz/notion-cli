// Package gen parses the Notion OpenAPI spec and produces a structured
// representation used to generate CLI commands.
package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Spec is the top-level OpenAPI 3.1 document (only fields we need).
type Spec struct {
	Info  SpecInfo             `json:"info"`
	Paths map[string]PathItem  `json:"paths"`
}

type SpecInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// PathItem maps HTTP methods to their operation definitions.
type PathItem map[string]json.RawMessage

// Operation represents a single API operation (one HTTP method on one path).
type Operation struct {
	// From the spec
	OperationID string    `json:"operationId"`
	Summary     string    `json:"summary"`
	Tags        []string  `json:"tags"`
	Parameters  []Param   `json:"parameters"`
	RequestBody *ReqBody  `json:"requestBody"`

	// Injected by the parser
	Method string
	Path   string
}

// Tag returns the primary tag (e.g. "Pages", "Blocks") or "misc".
func (o *Operation) Tag() string {
	if len(o.Tags) > 0 {
		return o.Tags[0]
	}
	return "misc"
}

// PathParams returns only path parameters (e.g. {page_id}).
func (o *Operation) PathParams() []Param {
	var out []Param
	for _, p := range o.Parameters {
		if p.In == "path" {
			out = append(out, p)
		}
	}
	return out
}

// QueryParams returns only query parameters.
func (o *Operation) QueryParams() []Param {
	var out []Param
	for _, p := range o.Parameters {
		if p.In == "query" {
			out = append(out, p)
		}
	}
	return out
}

// HasBody returns true if the operation accepts a request body.
func (o *Operation) HasBody() bool {
	return o.RequestBody != nil
}

// CobraUse returns the cobra Use string derived from path + method.
// e.g. GET /v1/pages/{page_id} → "get-page"
func (o *Operation) CobraUse() string {
	if o.OperationID != "" {
		// operationId is already kebab-case in Notion's spec
		return o.OperationID
	}
	// fallback: method + last path segment
	seg := o.Path
	seg = strings.TrimPrefix(seg, "/v1/")
	seg = strings.ReplaceAll(seg, "/", "-")
	seg = strings.ReplaceAll(seg, "{", "")
	seg = strings.ReplaceAll(seg, "}", "")
	return strings.ToLower(o.Method) + "-" + seg
}

// Param represents a single parameter (path or query).
type Param struct {
	Name     string     `json:"name"`
	In       string     `json:"in"` // "path" | "query" | "header"
	Required bool       `json:"required"`
	Schema   ParamSchema `json:"schema"`

	// $ref parameters are skipped (e.g. notionVersion header)
	Ref string `json:"$ref"`
}

// IsSkipped returns true for $ref params we don't expose to users.
func (p *Param) IsSkipped() bool {
	return p.Ref != "" || p.In == "header"
}

// ParamSchema holds just enough schema info to generate a flag.
type ParamSchema struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Enum        []string    `json:"enum"`
}

// ReqBody signals that the operation accepts a JSON body.
type ReqBody struct {
	Required bool `json:"required"`
}

// ParseSpec reads and parses the OpenAPI JSON file at path.
func ParseSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	return &spec, nil
}

// Operations extracts all operations from the spec, sorted by tag then path.
// OAuth endpoints are excluded — they are not relevant to the CLI.
func (s *Spec) Operations() []*Operation {
	var ops []*Operation

	for path, item := range s.Paths {
		for method, raw := range item {
			// PathItem may contain non-method keys like "parameters" — skip them
			method = strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}

			var op Operation
			if err := json.Unmarshal(raw, &op); err != nil {
				continue
			}
			op.Method = method
			op.Path = path

			// Skip OAuth — not relevant for a token-based CLI
			if op.Tag() == "OAuth" {
				continue
			}

			// Filter out $ref / header params
			filtered := op.Parameters[:0]
			for _, p := range op.Parameters {
				if !p.IsSkipped() {
					filtered = append(filtered, p)
				}
			}
			op.Parameters = filtered

			ops = append(ops, &op)
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		ti, tj := ops[i].Tag(), ops[j].Tag()
		if ti != tj {
			return ti < tj
		}
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})

	return ops
}

// OperationsByTag groups operations by their primary tag.
func (s *Spec) OperationsByTag() map[string][]*Operation {
	groups := make(map[string][]*Operation)
	for _, op := range s.Operations() {
		tag := op.Tag()
		groups[tag] = append(groups[tag], op)
	}
	return groups
}

func isHTTPMethod(s string) bool {
	switch s {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}
