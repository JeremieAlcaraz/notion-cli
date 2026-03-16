package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/4ier/notion-cli/internal/mode"
	"github.com/4ier/notion-cli/internal/tui"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// OutputFields is like OutputField but also supports --fields (CSV list of keys).
// Priority: --fields > --field > auto render.
func OutputFields(data []byte, format, field, fields string) error {
	if fields != "" {
		keys := strings.Split(fields, ",")
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		result := make(map[string]interface{}, len(keys))
		for _, k := range keys {
			k = strings.TrimSpace(k)
			if v, ok := obj[k]; ok {
				result[k] = v
			}
		}
		// Always compact in agent mode, indented otherwise
		var out []byte
		var err error
		if mode.IsAgent() {
			out, err = json.Marshal(result)
		} else {
			out, err = json.MarshalIndent(result, "", "  ")
		}
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	return OutputField(data, format, field)
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Output prints raw API response bytes in the requested format.
// Used by generated commands that receive []byte from client.Get/Post/Patch/Delete.
func Output(data []byte, format string) error {
	return OutputField(data, format, "")
}

// OutputField is like Output but extracts a single top-level field when field != "".
func OutputField(data []byte, format, field string) error {
	// If --field is set, extract the top-level field and print its value raw.
	if field != "" {
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		val, ok := obj[field]
		if !ok {
			return fmt.Errorf("field %q not found in response", field)
		}
		switch v := val.(type) {
		case string:
			fmt.Println(v)
		default:
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(out))
		}
		return nil
	}

	// Agent mode: compact JSON, no color, no tables
	if mode.IsAgent() {
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			fmt.Print(string(data))
			return nil
		}
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	}

	// Auto mode: try smart rendering first
	if format == "" {
		// Lists → gum table
		if isList(data) && RenderList(data) {
			return nil
		}
		// Single objects → jq colored JSON
		if IsTTY() {
			var v interface{}
			if err := json.Unmarshal(data, &v); err == nil {
				out, err := json.MarshalIndent(v, "", "  ")
				if err == nil && tui.ColorJSON(out) {
					return nil
				}
			}
		}
		format = "json"
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return nil
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// JSON outputs data as formatted JSON.
func JSON(data interface{}) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// Title prints a styled title.
func Title(icon, text string) {
	bold := color.New(color.Bold)
	bold.Printf("%s %s\n", icon, text)
}

// Subtitle prints a dimmed subtitle.
func Subtitle(text string) {
	dim := color.New(color.Faint)
	dim.Println(text)
}

// Separator prints a horizontal line.
func Separator() {
	fmt.Println(strings.Repeat("━", 40))
}

// Field prints a key-value pair.
func Field(key, value string) {
	keyColor := color.New(color.FgCyan)
	keyColor.Printf("%-16s", key+":")
	fmt.Println(value)
}

// Table prints rows in aligned columns.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No results.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Cap max width
	for i := range widths {
		if widths[i] > 60 {
			widths[i] = 60
		}
	}

	// Print header
	headerColor := color.New(color.Bold, color.FgWhite)
	for i, h := range headers {
		headerColor.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i := range headers {
		fmt.Print(strings.Repeat("─", widths[i]) + "  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				// Truncate if needed
				if len(cell) > widths[i] {
					cell = cell[:widths[i]-1] + "…"
				}
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}

// ExtractTitle extracts a readable title from a Notion page or database object.
func ExtractTitle(obj map[string]interface{}) string {
	// Database title
	if titleArr, ok := obj["title"].([]interface{}); ok {
		return extractPlainText(titleArr)
	}

	// Page title (in properties)
	if props, ok := obj["properties"].(map[string]interface{}); ok {
		for _, v := range props {
			prop, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			if prop["type"] == "title" {
				if titleArr, ok := prop["title"].([]interface{}); ok {
					return extractPlainText(titleArr)
				}
			}
		}
	}

	return "(untitled)"
}

func extractPlainText(richText []interface{}) string {
	var parts []string
	for _, t := range richText {
		if m, ok := t.(map[string]interface{}); ok {
			if pt, ok := m["plain_text"].(string); ok {
				parts = append(parts, pt)
			}
		}
	}
	text := strings.Join(parts, "")
	if text == "" {
		return "(untitled)"
	}
	return text
}
