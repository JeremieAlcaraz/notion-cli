# Tasks Plan — Notion CLI généré depuis OpenAPI

## Objectif

Remplacer le code hardcodé par une CLI générée automatiquement depuis la spec OpenAPI officielle de Notion.
À chaque mise à jour de la spec, on régénère la CLI.

## Spec source

- URL officielle : `https://developers.notion.com/openapi.json`
- Version MCP (plus récente, 2025-09-03) : `https://raw.githubusercontent.com/makenotion/notion-mcp-server/main/scripts/notion-openapi.json`
- Format : OpenAPI 3.1.0, ~801KB

## Architecture cible

```
notion-cli/
├── spec/
│   └── notion-openapi.json          # Spec snapshottée (source de vérité)
├── gen/
│   ├── generate.go                  # Script de génération (go generate)
│   ├── parser.go                    # Parse la spec OpenAPI
│   ├── templates/
│   │   ├── command.go.tmpl          # Template pour une commande cobra
│   │   ├── client.go.tmpl           # Template pour les appels HTTP
│   │   └── types.go.tmpl            # Template pour les structs Go
│   └── gen_test.go                  # Tests du générateur
├── internal/
│   ├── client/client.go             # Client HTTP (gardé, amélioré)
│   ├── config/config.go             # Auth (gardé)
│   └── render/render.go             # Output (gardé)
├── cmd/
│   └── generated/                   # ← CODE GÉNÉRÉ (ne pas éditer manuellement)
│       ├── pages.go
│       ├── databases.go
│       ├── blocks.go
│       ├── comments.go
│       ├── users.go
│       ├── search.go
│       └── file_uploads.go
├── main.go
└── justfile                         # just generate, just build, just update-spec
```

## Phases

- **Phase 1** — Fondations (T01–T03) : outillage, snapshot spec, justfile
- **Phase 2** — Générateur (T04–T07) : parser OpenAPI, templates, génération des commandes
- **Phase 3** — Migration (T08–T10) : brancher le code généré, supprimer le code hardcodé, tests
- **Phase 4** — Évolutivité (T11–T12) : CI auto-update spec, version header dynamique

---

## Phase 1 — Fondations

- [x] T01 Mettre à jour le header de version API
  Depends on: -
  Changes: internal/client/client.go
  Benefits: Corriger immédiatement les breaking changes (archived→in_trash, position, etc.)
  Tests: go build ./... && grep "2026-03-11" internal/client/client.go
  Commit: fix(client): bump Notion-Version header to 2026-03-11

- [x] T02 Snapshot la spec OpenAPI dans le repo
  Depends on: T01
  Changes: spec/notion-openapi.json, justfile (cible update-spec)
  Benefits: Source de vérité versionnée, reproductibilité offline
  Tests: jq '.info.version' spec/notion-openapi.json && just update-spec (doit overwriter sans erreur)
  Commit: chore(spec): add Notion OpenAPI 3.1.0 snapshot + just update-spec

- [x] T03 Créer le justfile avec les cibles essentielles
  Depends on: T02
  Changes: justfile
  Benefits: Interface unique pour generate/build/test/update-spec
  Tests: just affiche les cibles ; just build produit un binaire valide
  Commit: chore(just): add justfile with generate/build/test/update-spec targets

## Phase 2 — Générateur

- [x] T04 Créer le parser OpenAPI (gen/parser.go)
  Depends on: T02
  Changes: gen/parser.go, gen/go.mod si module séparé
  Benefits: Lit la spec et produit une représentation Go des endpoints (méthode, path, params, body, réponses)
  Tests: go test ./gen/... -run TestParser → vérifie qu'on extrait bien N endpoints depuis la spec snapshot
  Commit: feat(gen): add OpenAPI 3.1.0 parser for Notion spec

- [x] T05 Créer les templates Go (gen/templates/)
  Depends on: T04
  Changes: gen/templates/command.go.tmpl, client_method.go.tmpl, types.go.tmpl
  Benefits: Chaque endpoint devient une commande Cobra + une méthode client + des structs typés
  Tests: go test ./gen/... -run TestTemplateRender → golden file test sur un endpoint simple (GET /v1/users/me)
  Commit: feat(gen): add cobra command + client + types templates

- [x] T06 Créer le générateur principal (gen/generate.go)
  Depends on: T05
  Changes: gen/generate.go, cmd/generated/ (premier output)
  Benefits: `go generate ./...` ou `just generate` produit cmd/generated/*.go compilable
  Tests: just generate && go build ./... → zéro erreur de compilation
  Commit: feat(gen): wire generator — just generate produces compilable cmd/generated/

- [x] T07 Tests du générateur (gen/gen_test.go)
  Depends on: T06
  Changes: gen/gen_test.go
  Benefits: Garantit que si la spec change, le générateur détecte les breaking changes et échoue proprement
  Tests: go test ./gen/... -v → tous verts
  Commit: test(gen): add generator tests with golden files

## Phase 3 — Migration

- [ ] T08 Brancher cmd/generated/ dans root.go
  Depends on: T06
  Changes: cmd/root.go
  Benefits: Les commandes générées sont accessibles via le binaire
  Tests: go run . --help affiche les commandes générées ; go run . page view --help
  Commit: feat(cmd): register generated commands in root

- [ ] T09 Supprimer le code hardcodé (cmd/*.go manuels)
  Depends on: T08
  Changes: Suppression de cmd/page.go, cmd/db.go, cmd/block.go, cmd/user.go, cmd/comment.go, cmd/search.go, cmd/file.go, cmd/api.go
  Benefits: Single source of truth — plus de duplication entre spec et code
  Tests: go build ./... && go test ./... → aucune régression
  Commit: refactor(cmd): remove hand-written commands, all commands now generated

- [ ] T10 Mettre à jour README + DESIGN.md
  Depends on: T09
  Changes: README.md, DESIGN.md
  Benefits: Documentation alignée avec la nouvelle architecture
  Tests: Lecture humaine — les instructions d'installation et de contribution sont correctes
  Commit: docs: update README and DESIGN for OpenAPI-driven architecture

## Phase 4 — Évolutivité

- [ ] T11 GitHub Action : auto-update spec + PR
  Depends on: T10
  Changes: .github/workflows/update-spec.yml
  Benefits: Détecte automatiquement les nouvelles versions de la spec Notion et ouvre une PR
  Tests: Déclencher manuellement le workflow via gh workflow run update-spec.yml
  Commit: ci: add weekly spec auto-update workflow

- [ ] T12 Ajouter la commande `notion api` générique (escape hatch)
  Depends on: T08
  Changes: cmd/api.go (réécrit proprement, non généré)
  Benefits: Permet d'appeler n'importe quel endpoint non encore couvert sans recompiler
  Tests: notion api GET /v1/users/me → retourne le JSON de l'utilisateur courant
  Commit: feat(cmd): restore generic api escape hatch command
