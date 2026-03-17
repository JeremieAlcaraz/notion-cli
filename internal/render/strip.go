package render

import "encoding/json"

var stripMeta bool

// SetStripMeta enables or disables automatic metadata stripping.
func SetStripMeta(v bool) { stripMeta = v }

// IsStripMeta returns true if --strip-meta is active.
func IsStripMeta() bool { return stripMeta }

// metaFields are Notion response fields stripped by --strip-meta:
// - request_id: Notion internal support ID, no semantic value for agents
// - created_by / last_edited_by: authorship rarely needed for decisions
// - public_url: almost always null
// Note: cover and icon are intentionally kept — agents may need them
// for page appearance tasks (e.g. "set an icon", "read the cover").
var metaFields = []string{
	"request_id",
	"created_by",
	"last_edited_by",
	"public_url",
}

// conditionalFalseFields are stripped only when their value is false/null.
var conditionalFalseFields = []string{
	"in_trash",
	"archived",
	"is_inline",
	"is_locked",
}

// StripMeta removes noisy metadata fields from a Notion response object.
// Works recursively on list responses (strips each item in results[]).
func StripMeta(data map[string]interface{}) map[string]interface{} {
	// Always remove unconditional meta fields
	for _, f := range metaFields {
		delete(data, f)
	}

	// Remove conditional fields only when false/null
	for _, f := range conditionalFalseFields {
		v, ok := data[f]
		if !ok {
			continue
		}
		switch val := v.(type) {
		case bool:
			if !val {
				delete(data, f)
			}
		case nil:
			delete(data, f)
		}
	}

	// Recurse into results[] for list responses
	if data["object"] == "list" {
		if items, ok := data["results"].([]interface{}); ok {
			for i, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					items[i] = StripMeta(m)
				}
			}
		}
	}

	return data
}

// StripMetaBytes applies StripMeta to raw JSON bytes.
// Returns original bytes if parsing fails.
func StripMetaBytes(data []byte) []byte {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data
	}
	stripped := StripMeta(obj)
	out, err := json.Marshal(stripped)
	if err != nil {
		return data
	}
	return out
}
