package gen

import (
	"bytes"
	"strings"
	"testing"
)

func TestToPascal(t *testing.T) {
	cases := []struct{ in, want string }{
		{"retrieve-a-page", "RetrieveAPage"},
		{"get-self", "GetSelf"},
		{"post-search", "PostSearch"},
		{"patch-page", "PatchPage"},
		{"page_id", "PageId"},
	}
	for _, c := range cases {
		if got := toPascal(c.in); got != c.want {
			t.Errorf("toPascal(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestToKebab(t *testing.T) {
	cases := []struct{ in, want string }{
		{"page_id", "page-id"},
		{"filter_properties", "filter-properties"},
		{"block_id", "block-id"},
	}
	for _, c := range cases {
		if got := toKebab(c.in); got != c.want {
			t.Errorf("toKebab(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRenderCommandFile_GetSelf(t *testing.T) {
	// Build a minimal operation (GET /v1/users/me — no params, no body)
	op := &Operation{
		OperationID: "get-self",
		Summary:     "Retrieve your token's bot user",
		Tags:        []string{"Users"},
		Method:      "GET",
		Path:        "/v1/users/me",
	}

	var buf bytes.Buffer
	err := RenderCommandFile(&buf, "Users", []*Operation{op})
	if err != nil {
		t.Fatalf("RenderCommandFile error: %v", err)
	}

	out := buf.String()

	// Must contain the function name
	if !strings.Contains(out, "newGetSelfCmd") {
		t.Error("expected newGetSelfCmd function")
	}
	// Must reference the correct path
	if !strings.Contains(out, `"/v1/users/me"`) {
		t.Error("expected path /v1/users/me")
	}
	// Must use GET
	if !strings.Contains(out, "c.Get(path)") {
		t.Error("expected c.Get(path)")
	}
	// Must NOT have body flag
	if strings.Contains(out, `"body"`) {
		t.Error("GET command should not have --body flag")
	}
}

func TestRenderCommandFile_PostWithBody(t *testing.T) {
	op := &Operation{
		OperationID: "post-search",
		Summary:     "Search by title",
		Tags:        []string{"Search"},
		Method:      "POST",
		Path:        "/v1/search",
		RequestBody: &ReqBody{Required: false},
	}

	var buf bytes.Buffer
	err := RenderCommandFile(&buf, "Search", []*Operation{op})
	if err != nil {
		t.Fatalf("RenderCommandFile error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "newPostSearchCmd") {
		t.Error("expected newPostSearchCmd function")
	}
	if !strings.Contains(out, `"body"`) {
		t.Error("POST command should have --body flag")
	}
	if !strings.Contains(out, "c.Post(path") {
		t.Error("expected c.Post(path...)")
	}
}

func TestRenderCommandFile_PatchWithPathParam(t *testing.T) {
	op := &Operation{
		OperationID: "patch-page",
		Summary:     "Update page",
		Tags:        []string{"Pages"},
		Method:      "PATCH",
		Path:        "/v1/pages/{page_id}",
		Parameters: []Param{
			{Name: "page_id", In: "path", Required: true},
		},
		RequestBody: &ReqBody{Required: true},
	}

	var buf bytes.Buffer
	err := RenderCommandFile(&buf, "Pages", []*Operation{op})
	if err != nil {
		t.Fatalf("RenderCommandFile error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "newPatchPageCmd") {
		t.Error("expected newPatchPageCmd function")
	}
	// Path param replaced at runtime
	if !strings.Contains(out, `strings.ReplaceAll(path, "{page_id}", args[0])`) {
		t.Error("expected path param substitution")
	}
	// ArbitraryArgs used for flexibility with path params
	if !strings.Contains(out, "cobra.ArbitraryArgs") {
		t.Error("expected cobra.ArbitraryArgs")
	}
}

func TestRenderRootFile(t *testing.T) {
	byTag := map[string][]*Operation{
		"Users": {
			{OperationID: "get-self", Method: "GET", Path: "/v1/users/me", Tags: []string{"Users"}},
		},
	}

	var buf bytes.Buffer
	err := RenderRootFile(&buf, byTag)
	if err != nil {
		t.Fatalf("RenderRootFile error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "func AddTo") {
		t.Error("expected AddTo function")
	}
	if !strings.Contains(out, `"users"`) {
		t.Error("expected users group slug")
	}
	if !strings.Contains(out, "newGetSelfCmd") {
		t.Error("expected newGetSelfCmd reference in root")
	}
}
