package gen

import (
	"bytes"
	"go/format"
	"os"
	"testing"
)

// updateGolden=true regenerates golden files instead of comparing.
// Usage: UPDATE_GOLDEN=1 go test ./gen/... -run TestGolden
var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGolden_UsersFile(t *testing.T) {
	testGoldenFile(t, "Users", "testdata/users.go.golden")
}

func TestGolden_PagesFile(t *testing.T) {
	testGoldenFile(t, "Pages", "testdata/pages.go.golden")
}

func TestGolden_RegisterFile(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	raw, err := RenderRootFileBytes(spec.OperationsByTag())
	if err != nil {
		t.Fatalf("RenderRootFileBytes: %v", err)
	}

	got, err := format.Source(raw)
	if err != nil {
		t.Fatalf("gofmt: %v\nraw output:\n%s", err, raw)
	}

	goldenPath := "testdata/register.go.golden"
	if updateGolden {
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run UPDATE_GOLDEN=1 to create)", goldenPath, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("register.go differs from golden file\n"+
			"Run: UPDATE_GOLDEN=1 go test ./gen/... -run TestGolden\n"+
			"to update golden files after intentional spec/template changes")
	}
}

func TestGolden_Idempotent(t *testing.T) {
	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	byTag := spec.OperationsByTag()

	// Run generation twice and compare — must be identical
	run := func() []byte {
		raw, err := RenderCommandFileBytes("Users", byTag["Users"])
		if err != nil {
			t.Fatalf("RenderCommandFileBytes: %v", err)
		}
		out, err := format.Source(raw)
		if err != nil {
			t.Fatalf("gofmt: %v", err)
		}
		return out
	}

	first := run()
	second := run()

	if !bytes.Equal(first, second) {
		t.Error("generator is not idempotent: two runs produced different output")
	}
}

// testGoldenFile is a helper that renders a tag and compares to a golden file.
func testGoldenFile(t *testing.T, tag, goldenPath string) {
	t.Helper()

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	ops := spec.OperationsByTag()[tag]
	if len(ops) == 0 {
		t.Fatalf("no operations for tag %q", tag)
	}

	raw, err := RenderCommandFileBytes(tag, ops)
	if err != nil {
		t.Fatalf("RenderCommandFileBytes: %v", err)
	}

	got, err := format.Source(raw)
	if err != nil {
		t.Fatalf("gofmt: %v\nraw output:\n%s", err, raw)
	}

	if updateGolden {
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run UPDATE_GOLDEN=1 to create)", goldenPath, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("%s differs from golden file\n"+
			"Run: UPDATE_GOLDEN=1 go test ./gen/... -run TestGolden\n"+
			"to update golden files after intentional spec/template changes",
			goldenPath)
	}
}
