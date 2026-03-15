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

# Build + run tests
check: build test
