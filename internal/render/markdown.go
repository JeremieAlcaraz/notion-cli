package render

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BlocksToMarkdown converts a Notion blocks list response to Markdown.
// Supports: heading_1/2/3, paragraph, bulleted_list_item, numbered_list_item,
// numbered_list_item, code, quote, callout, to_do, divider, toggle,
// image, video, child_page, child_database, column_list.
func BlocksToMarkdown(data []byte) (string, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	results, ok := resp["results"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no results field in response")
	}

	var sb strings.Builder
	numberedIdx := 0 // track numbered list continuity

	for _, item := range results {
		block, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		blockType, _ := block["type"].(string)

		// Reset numbered list index on non-numbered blocks
		if blockType != "numbered_list_item" {
			numberedIdx = 0
		}

		line := blockToMarkdown(block, blockType, &numberedIdx)
		if line != "" {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

func blockToMarkdown(block map[string]interface{}, blockType string, numberedIdx *int) string {
	switch blockType {
	case "heading_1":
		return "# " + richText(block, "heading_1")
	case "heading_2":
		return "## " + richText(block, "heading_2")
	case "heading_3":
		return "### " + richText(block, "heading_3")

	case "paragraph":
		text := richText(block, "paragraph")
		if text == "" {
			return "" // blank line between paragraphs
		}
		return text

	case "bulleted_list_item":
		return "- " + richText(block, "bulleted_list_item")

	case "numbered_list_item":
		*numberedIdx++
		return fmt.Sprintf("%d. %s", *numberedIdx, richText(block, "numbered_list_item"))

	case "to_do":
		inner, _ := block["to_do"].(map[string]interface{})
		checked, _ := inner["checked"].(bool)
		box := "[ ]"
		if checked {
			box = "[x]"
		}
		return "- " + box + " " + richText(block, "to_do")

	case "code":
		inner, _ := block["code"].(map[string]interface{})
		lang, _ := inner["language"].(string)
		text := richText(block, "code")
		return fmt.Sprintf("```%s\n%s\n```", lang, text)

	case "quote":
		text := richText(block, "quote")
		// Prefix each line with >
		lines := strings.Split(text, "\n")
		for i, l := range lines {
			lines[i] = "> " + l
		}
		return strings.Join(lines, "\n")

	case "callout":
		inner, _ := block["callout"].(map[string]interface{})
		emoji := ""
		if icon, ok := inner["icon"].(map[string]interface{}); ok {
			emoji, _ = icon["emoji"].(string)
		}
		text := richText(block, "callout")
		if emoji != "" {
			return fmt.Sprintf("> %s **%s**", emoji, text)
		}
		return "> **" + text + "**"

	case "toggle":
		text := richText(block, "toggle")
		return "<details><summary>" + text + "</summary></details>"

	case "divider":
		return "---"

	case "image":
		inner, _ := block["image"].(map[string]interface{})
		caption := captionText(inner)
		url := mediaURL(inner)
		if caption != "" {
			return fmt.Sprintf("![%s](%s)", caption, url)
		}
		return fmt.Sprintf("![](%s)", url)

	case "video":
		inner, _ := block["video"].(map[string]interface{})
		url := mediaURL(inner)
		caption := captionText(inner)
		if caption != "" {
			return fmt.Sprintf("[▶ %s](%s)", caption, url)
		}
		return fmt.Sprintf("[▶ video](%s)", url)

	case "child_page":
		inner, _ := block["child_page"].(map[string]interface{})
		title, _ := inner["title"].(string)
		id, _ := block["id"].(string)
		if title == "" {
			title = id
		}
		return fmt.Sprintf("📄 [%s](notion://%s)", title, id)

	case "child_database":
		inner, _ := block["child_database"].(map[string]interface{})
		title, _ := inner["title"].(string)
		id, _ := block["id"].(string)
		if title == "" {
			title = id
		}
		return fmt.Sprintf("🗃 [%s](notion://%s)", title, id)

	case "column_list":
		// Columns require child fetching — indicate presence only
		return "<!-- column_list: fetch children for full content -->"

	default:
		// Unknown block type — emit a comment so nothing is silently lost
		return fmt.Sprintf("<!-- unsupported block: %s -->", blockType)
	}
}

// richText extracts plain text from a block's rich_text array,
// applying basic Markdown annotations (bold, italic, code, strikethrough).
func richText(block map[string]interface{}, key string) string {
	inner, ok := block[key].(map[string]interface{})
	if !ok {
		return ""
	}
	arr, ok := inner["rich_text"].([]interface{})
	if !ok {
		return ""
	}
	return richTextArray(arr)
}

func richTextArray(arr []interface{}) string {
	var sb strings.Builder
	for _, item := range arr {
		seg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		plain, _ := seg["plain_text"].(string)
		if plain == "" {
			continue
		}
		ann, _ := seg["annotations"].(map[string]interface{})
		text := applyAnnotations(plain, ann)
		sb.WriteString(text)
	}
	return sb.String()
}

func applyAnnotations(text string, ann map[string]interface{}) string {
	if ann == nil {
		return text
	}
	// Apply in reverse nesting order: code first (no nesting inside code)
	if b, _ := ann["code"].(bool); b {
		return "`" + text + "`"
	}
	if b, _ := ann["bold"].(bool); b {
		text = "**" + text + "**"
	}
	if b, _ := ann["italic"].(bool); b {
		text = "_" + text + "_"
	}
	if b, _ := ann["strikethrough"].(bool); b {
		text = "~~" + text + "~~"
	}
	return text
}

func captionText(inner map[string]interface{}) string {
	arr, ok := inner["caption"].([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}
	return richTextArray(arr)
}

func mediaURL(inner map[string]interface{}) string {
	// Hosted file
	if f, ok := inner["file"].(map[string]interface{}); ok {
		if url, ok := f["url"].(string); ok {
			return url
		}
	}
	// External URL
	if e, ok := inner["external"].(map[string]interface{}); ok {
		if url, ok := e["url"].(string); ok {
			return url
		}
	}
	return ""
}
