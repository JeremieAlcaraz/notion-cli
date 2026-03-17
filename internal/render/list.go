package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/JeremieAlcaraz/notion-cli/internal/tui"
)

// isList returns true if the JSON data is a Notion list response.
func isList(data []byte) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return false
	}
	return obj["object"] == "list"
}

// RenderList renders a Notion list response as a gum table.
// Returns false if rendering was not possible (fallback to JSON).
func RenderList(data []byte) bool {
	if !tui.IsAvailable() || !IsTTY() {
		return false
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return false
	}

	results, ok := obj["results"].([]interface{})
	if !ok || len(results) == 0 {
		fmt.Fprintln(os.Stderr, "(no results)")
		return true
	}

	// Determine columns and row extractor based on object type
	cols, extractor := columnsFor(results)

	// Build CSV
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(cols)
	for _, item := range results {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		_ = w.Write(extractor(m))
	}
	w.Flush()

	// Print count hint
	total := len(results)
	hasMore, _ := obj["has_more"].(bool)
	suffix := ""
	if hasMore {
		suffix = "+"
	}
	fmt.Fprintf(os.Stderr, "  %d%s result(s)\n\n", total, suffix)

	// Pipe through gum table --print
	cmd := exec.Command("gum", "table", "--print", "--border", "rounded")
	cmd.Stdin = &buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

// columnsFor returns column headers and a row extractor for a slice of results.
func columnsFor(results []interface{}) ([]string, func(map[string]interface{}) []string) {
	if len(results) == 0 {
		return nil, nil
	}
	first, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	objType, _ := first["object"].(string)

	switch objType {
	case "user":
		cols := []string{"Name", "Type", "ID"}
		return cols, func(m map[string]interface{}) []string {
			return []string{
				str(m["name"]),
				str(m["type"]),
				str(m["id"]),
			}
		}
	case "page":
		cols := []string{"Title", "ID", "Created", "In Trash"}
		return cols, func(m map[string]interface{}) []string {
			return []string{
				ExtractTitle(m),
				str(m["id"]),
				formatTime(str(m["created_time"])),
				boolStr(m["in_trash"]),
			}
		}
	case "database":
		cols := []string{"Title", "ID", "Created"}
		return cols, func(m map[string]interface{}) []string {
			return []string{
				ExtractTitle(m),
				str(m["id"]),
				formatTime(str(m["created_time"])),
			}
		}
	case "block":
		cols := []string{"Type", "ID", "Created"}
		return cols, func(m map[string]interface{}) []string {
			return []string{
				str(m["type"]),
				str(m["id"]),
				formatTime(str(m["created_time"])),
			}
		}
	case "comment":
		cols := []string{"ID", "Created", "Text"}
		return cols, func(m map[string]interface{}) []string {
			return []string{
				str(m["id"]),
				formatTime(str(m["created_time"])),
				extractRichText(m["rich_text"]),
			}
		}
	default:
		// Generic fallback: id + object + created_time if present
		cols := []string{"Object", "ID"}
		if _, ok := first["created_time"]; ok {
			cols = append(cols, "Created")
		}
		return cols, func(m map[string]interface{}) []string {
			row := []string{str(m["object"]), str(m["id"])}
			if _, ok := m["created_time"]; ok {
				row = append(row, formatTime(str(m["created_time"])))
			}
			return row
		}
	}
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func boolStr(v interface{}) string {
	if b, ok := v.(bool); ok && b {
		return "yes"
	}
	return "no"
}

func formatTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04")
}

func extractRichText(v interface{}) string {
	arr, ok := v.([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if pt, ok := m["plain_text"].(string); ok {
			parts = append(parts, pt)
		}
	}
	text := strings.Join(parts, "")
	if len(text) > 60 {
		return text[:57] + "…"
	}
	return text
}
