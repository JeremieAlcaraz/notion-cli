package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JeremieAlcaraz/notion-cli/internal/mode"
	"github.com/fatih/color"
)

// Summary extracts a compact summary from a Notion API response.
// In agent mode: compact JSON with just id, type, title, url, parent_id.
// In human mode: a readable table or key-value block.
func Summary(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Println(string(data))
		return nil
	}

	// List response → summarize each item
	if obj, ok := raw.(map[string]interface{}); ok {
		if obj["object"] == "list" {
			if items, ok := obj["results"].([]interface{}); ok {
				return summaryList(items)
			}
		}
		// Single object
		return summarySingle(obj)
	}

	// Bare array
	if items, ok := raw.([]interface{}); ok {
		return summaryList(items)
	}

	fmt.Println(string(data))
	return nil
}

type summaryItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
}

func extractSummary(obj map[string]interface{}) summaryItem {
	s := summaryItem{}

	if id, ok := obj["id"].(string); ok {
		s.ID = id
	}
	if t, ok := obj["object"].(string); ok {
		s.Type = t
	}
	s.Title = ExtractTitle(obj)
	if u, ok := obj["url"].(string); ok {
		s.URL = u
	}
	if p, ok := obj["parent"].(map[string]interface{}); ok {
		for _, key := range []string{"page_id", "database_id", "block_id", "workspace"} {
			if v, ok := p[key]; ok {
				switch val := v.(type) {
				case string:
					s.ParentID = val
				case bool:
					if val {
						s.ParentID = "workspace"
					}
				}
				break
			}
		}
	}
	return s
}

func summaryList(items []interface{}) error {
	summaries := make([]summaryItem, 0, len(items))
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok {
			summaries = append(summaries, extractSummary(obj))
		}
	}

	if mode.IsAgent() {
		out, _ := json.Marshal(summaries)
		fmt.Print(string(out))
		return nil
	}

	// Human: table
	headers := []string{"ID", "TYPE", "TITLE", "PARENT"}
	rows := make([][]string, 0, len(summaries))
	for _, s := range summaries {
		title := s.Title
		if len(title) > 40 {
			title = title[:39] + "…"
		}
		parent := s.ParentID
		if len(parent) > 20 {
			parent = parent[:19] + "…"
		}
		rows = append(rows, []string{s.ID, s.Type, title, parent})
	}
	printSummaryTable(headers, rows)
	return nil
}

func summarySingle(obj map[string]interface{}) error {
	s := extractSummary(obj)

	if mode.IsAgent() {
		out, _ := json.Marshal(s)
		fmt.Print(string(out))
		return nil
	}

	// Human: key-value block
	keyColor := color.New(color.FgCyan)
	keyColor.Printf("%-12s", "ID:")
	fmt.Println(s.ID)
	keyColor.Printf("%-12s", "Type:")
	fmt.Println(s.Type)
	keyColor.Printf("%-12s", "Title:")
	fmt.Println(s.Title)
	if s.URL != "" {
		keyColor.Printf("%-12s", "URL:")
		fmt.Println(s.URL)
	}
	if s.ParentID != "" {
		keyColor.Printf("%-12s", "Parent ID:")
		fmt.Println(s.ParentID)
	}
	return nil
}

func printSummaryTable(headers []string, rows [][]string) {
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
	maxWidths := []int{36, 12, 40, 36}
	for i := range widths {
		if i < len(maxWidths) && widths[i] > maxWidths[i] {
			widths[i] = maxWidths[i]
		}
	}

	headerColor := color.New(color.Bold, color.FgWhite)
	for i, h := range headers {
		headerColor.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()
	for i := range headers {
		fmt.Print(strings.Repeat("─", widths[i]) + "  ")
	}
	fmt.Println()
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if len(cell) > widths[i] {
					cell = cell[:widths[i]-1] + "…"
				}
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}
