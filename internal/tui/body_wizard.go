package tui

import (
	"fmt"
	"strings"

	"github.com/4ier/notion-cli/internal/mode"
)

// BodyField describes one field of a request body.
type BodyField struct {
	Name        string
	Type        string // "string", "boolean", "integer", "number", "array", "object", ""
	Description string
	Required    bool
}

// AskBody interactively prompts for each field of a JSON body using gum input.
// Returns a JSON string built from the user's answers.
// Skips fields the user leaves blank (unless required).
// Falls back to returning "" (caller must handle) if gum is unavailable.
func AskBody(operationID string, fields []BodyField) (string, error) {
	if mode.IsAgent() || !IsAvailable() || !isTTY() {
		return "", nil
	}

	pairs := []string{}

	for _, f := range fields {
		placeholder := placeholderFor(f)
		label := f.Name
		if f.Required {
			label += " (required)"
		}
		if f.Description != "" {
			label += " — " + f.Description
		}

		val, err := AskInput(label+": ", placeholder)
		if err != nil {
			// User cancelled this field — skip if not required, abort if required
			if f.Required {
				return "", fmt.Errorf("field %q is required", f.Name)
			}
			continue
		}
		if val == "" {
			if f.Required {
				return "", fmt.Errorf("field %q is required", f.Name)
			}
			continue
		}

		pairs = append(pairs, jsonPair(f.Name, f.Type, val))
	}

	if len(pairs) == 0 {
		return "{}", nil
	}
	return "{" + strings.Join(pairs, ",") + "}", nil
}

// placeholderFor returns a helpful placeholder string based on field type.
func placeholderFor(f BodyField) string {
	switch f.Type {
	case "string":
		return `"my value"`
	case "boolean":
		return "true or false"
	case "integer", "number":
		return "42"
	case "array":
		return `[{"type":"text","text":{"content":"…"}}]`
	default:
		return `{"key": "value"}`
	}
}

// jsonPair formats a key-value pair as JSON based on the field type.
// For string fields the value is quoted; for others it's used raw.
func jsonPair(name, typ, val string) string {
	switch typ {
	case "string":
		// Quote the value, escaping internal quotes
		escaped := strings.ReplaceAll(val, `"`, `\"`)
		return fmt.Sprintf(`%q:%q`, name, escaped)
	default:
		// boolean, number, array, object — use raw value
		return fmt.Sprintf(`%q:%s`, name, val)
	}
}
