package gen

import (
	"testing"
)

const specPath = "../spec/notion-openapi.json"

func TestParseSpec_LoadsWithoutError(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	if spec.Info.Title == "" {
		t.Error("expected non-empty spec title")
	}
}

func TestOperations_Count(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	ops := spec.Operations()
	// Spec has 26 paths × methods; minus OAuth (3 endpoints) = at least 20 ops
	if len(ops) < 20 {
		t.Errorf("expected at least 20 operations, got %d", len(ops))
	}
	t.Logf("total operations (excl. OAuth): %d", len(ops))
}

func TestOperations_NoOAuth(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	for _, op := range spec.Operations() {
		if op.Tag() == "OAuth" {
			t.Errorf("OAuth operation leaked through: %s %s", op.Method, op.Path)
		}
	}
}

func TestOperations_AllHaveMethodAndPath(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	for _, op := range spec.Operations() {
		if op.Method == "" {
			t.Errorf("operation %q has empty Method", op.OperationID)
		}
		if op.Path == "" {
			t.Errorf("operation %q has empty Path", op.OperationID)
		}
	}
}

func TestOperationsByTag_Groups(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	groups := spec.OperationsByTag()

	expectedTags := []string{"Pages", "Blocks", "Users", "Comments"}
	for _, tag := range expectedTags {
		if _, ok := groups[tag]; !ok {
			t.Errorf("expected tag %q not found in groups", tag)
		}
	}
	t.Logf("tags found: %v", tagKeys(groups))
}

func TestOperation_PathParams(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	for _, op := range spec.Operations() {
		if op.Path == "/v1/pages/{page_id}" && op.Method == "GET" {
			params := op.PathParams()
			if len(params) != 1 || params[0].Name != "page_id" {
				t.Errorf("expected path param page_id, got %+v", params)
			}
			return
		}
	}
	t.Error("GET /v1/pages/{page_id} not found")
}

func TestOperation_CobraUse(t *testing.T) {
	op := &Operation{
		OperationID: "retrieve-a-page",
		Method:      "GET",
		Path:        "/v1/pages/{page_id}",
		Tags:        []string{"Pages"},
	}
	if got := op.CobraUse(); got != "retrieve-a-page" {
		t.Errorf("CobraUse = %q, want %q", got, "retrieve-a-page")
	}
}

func tagKeys(m map[string][]*Operation) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
