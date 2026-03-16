# Notion CLI — task runner
# Usage: just <recipe>

default:
    @just --list

# Download the latest Notion OpenAPI spec
update-spec:
    curl -s https://developers.notion.com/openapi.json -o spec/notion-openapi.json
    @echo "Spec updated: $(jq -r '.info.title + " v" + .info.version' spec/notion-openapi.json)"
    @echo "Endpoints: $(jq '.paths | keys | length' spec/notion-openapi.json)"

# Generate CLI code from OpenAPI spec
generate:
    go run ./gen/generate.go

# Build the binary
build:
    go build -o notion .

# Run tests
test:
    go test ./...

# Update golden test files after intentional spec/template changes
update-golden:
    UPDATE_GOLDEN=1 go test ./gen/... -run TestGolden

# Build + run tests
check: build test

# Run token benchmark (corpus → count → compare → regression check)
bench: build
    @echo "==> Collecting API outputs..."
    bash bench/corpus.sh
    @echo "==> Counting tokens..."
    bench/.venv/bin/python bench/count_tokens.py
    @echo "==> Comparing human vs agent..."
    bench/.venv/bin/python bench/compare.py
    @echo "==> Checking regression..."
    bench/.venv/bin/python bench/check_regression.py

# Update benchmark baseline after intentional improvement
bench-update-baseline:
    cp bench/results/tokens.csv bench/results/baseline.csv
    @echo "Baseline updated."
