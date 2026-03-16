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
│   │   └── root.go.tmpl             # Template pour le registre des commandes
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
│       ├── file_uploads.go
│       ├── file_upload_manual.go    # ← Manuel (multipart, non généré)
│       └── helpers.go               # ← Manuel (token, non généré)
├── main.go
└── justfile                         # just generate, just build, just update-spec
```

## Phases

- **Phase 1** — Fondations (T01–T03) : outillage, snapshot spec, justfile ✅
- **Phase 2** — Générateur (T04–T07) : parser OpenAPI, templates, génération des commandes ✅
- **Phase 3** — Migration (T08–T10) : brancher le code généré, supprimer le code hardcodé, docs ✅
- **Phase 4** — Fiabilité (T13–T19) : améliorations issues de la batterie de tests ✅
- **Phase 5** — Évolutivité (T11–T12) : CI auto-update spec, escape hatch api

---

## Phase 1 — Fondations

- [x] T01 Mettre à jour le header de version API
  Depends on: -
  Changes: internal/client/client.go
  Benefits: Corriger immédiatement les breaking changes (archived→in_trash, position, etc.)
  Tests: go build ./... && grep "2026-03-11" internal/client/client.go
  Commit: fix: bump Notion-Version header to 2026-03-11

- [x] T02 Snapshot la spec OpenAPI dans le repo
  Depends on: T01
  Changes: spec/notion-openapi.json, justfile (cible update-spec)
  Benefits: Source de vérité versionnée, reproductibilité offline
  Tests: jq '.info.version' spec/notion-openapi.json && just update-spec
  Commit: dev: add OpenAPI spec snapshot and justfile

- [x] T03 Créer le justfile avec les cibles essentielles
  Depends on: T02
  Changes: justfile
  Benefits: Interface unique pour generate/build/test/update-spec
  Tests: just affiche les cibles ; just build produit un binaire valide
  Commit: (inclus dans T02)

## Phase 2 — Générateur

- [x] T04 Créer le parser OpenAPI (gen/parser.go)
  Depends on: T02
  Changes: gen/parser.go, gen/parser_test.go
  Benefits: Lit la spec et produit une représentation Go structurée des endpoints
  Tests: go test ./gen/... -run TestParser → 7/7 verts, 32 opérations parsées
  Commit: feat: add OpenAPI 3.1 parser for Notion spec

- [x] T05 Créer les templates Go (gen/templates/)
  Depends on: T04
  Changes: gen/templates/command.go.tmpl, gen/templates/root.go.tmpl, gen/renderer.go
  Benefits: Chaque endpoint devient une commande Cobra typée
  Tests: go test ./gen/... -run TestTemplate → 12/12 verts
  Commit: feat: add cobra command templates for code generation

- [x] T06 Créer le générateur principal (gen/generate.go)
  Depends on: T05
  Changes: gen/generate.go, cmd/generated/ (premier output), internal/render/render.go
  Benefits: just generate produit cmd/generated/*.go compilable
  Tests: just generate && go build ./... → zéro erreur
  Commit: feat: wire generator — just generate produces cmd/generated

- [x] T07 Tests du générateur avec golden files
  Depends on: T06
  Changes: gen/golden_test.go, gen/testdata/
  Benefits: Tout changement de spec ou template est détecté explicitement
  Tests: go test ./gen/... -v → 17/17 verts
  Commit: test: add golden file tests for generator

## Phase 3 — Migration

- [x] T08 Brancher cmd/generated/ dans root.go
  Depends on: T06
  Changes: cmd/root.go
  Benefits: Les 32 commandes générées sont accessibles via le binaire
  Tests: just build && ./notion --help affiche les groupes générés
  Commit: feat: register generated commands in root

- [x] T09 Supprimer le code hardcodé (cmd/*.go manuels)
  Depends on: T08
  Changes: Suppression de cmd/page.go, cmd/db.go, cmd/block.go, cmd/user.go,
           cmd/comment.go, cmd/search.go, cmd/file.go + nettoyage root.go
  Benefits: Single source of truth — plus de duplication
  Tests: go build ./... && ./notion --help → plus de doublons
  Commit: refactor: remove hand-written commands, all generated

- [x] T10 Documenter résultats de tests et recommandations
  Depends on: T09
  Changes: TESTING.md, DESIGN.md
  Benefits: Traçabilité des décisions et des limitations connues
  Tests: Lecture humaine
  Commit: docs: add test results and reliability recommendations

## Phase 4 — Fiabilité (issues R1–R7)

- [x] T13 R1 — Améliorer les messages d'erreur data-sources
  Depends on: T09
  Changes: internal/client/client.go (errorHint), cmd/generated/data_sources.go (hint ajouté)
  Benefits: L'utilisateur comprend immédiatement qu'il faut un data_source_id ≠ database_id
  Tests: ./notion data-sources retrieve-a-data-source <database_id> → message d'erreur explicite
  Commit: fix: improve error hint for data-sources ID confusion

- [x] T14 R2 — Supporter --body @file.json et --body - (stdin)
  Depends on: T09
  Changes: cmd/generated/helpers.go (resolveBody helper), gen/templates/command.go.tmpl
  Benefits: Fini les escaping hell en shell — on passe un fichier JSON directement
  Tests: echo '{"query":"test"}' | ./notion search post-search --body - → fonctionne
  Commit: feat: support --body @file and --body - for stdin input

- [x] T15 R3 — Mécanisme "manually implemented" dans le générateur
  Depends on: T06
  Changes: gen/generate.go (liste d'exclusion), gen/templates/command.go.tmpl
  Benefits: just generate ne peut plus écraser file_upload_manual.go ni d'autres overrides
  Tests: just generate && ls cmd/generated/file_upload_manual.go → toujours présent et intact
  Commit: feat: add manual-override exclusion list to generator

- [x] T16 R4 — Flag --field pour extraire un champ top-level
  Depends on: T09
  Changes: internal/render/render.go (Output accepte --field), cmd/root.go (flag global)
  Benefits: notion file-uploads create-file --body '{...}' --field id → juste l'ID
  Tests: ./notion users get-self --field name → "cli-demo-test"
  Commit: feat: add --field flag to extract top-level JSON field

- [x] T17 R5 — Clarifier complete-file-upload (>20MB only)
  Depends on: T09
  Changes: cmd/generated/file_upload_manual.go (Short + Long améliorés)
  Benefits: L'utilisateur ne perd plus de temps à tester une commande qui ne s'applique pas
  Tests: ./notion file-uploads complete-file-upload --help → description claire
  Commit: docs: clarify complete-file-upload is for multi-part uploads only

- [x] T18 R6 — Flag --dry-run global
  Depends on: T09
  Changes: internal/client/client.go (dry-run mode), cmd/root.go (flag global)
  Benefits: Voir la requête HTTP avant de l'exécuter — essentiel pour le debug
  Tests: ./notion blocks delete-a-block <id> --dry-run → affiche DELETE /v1/blocks/<id> sans appeler l'API
  Commit: feat: add --dry-run flag to preview HTTP requests

- [x] T19 R7 — Flag --help-body avec exemple depuis la spec
  Depends on: T06
  Changes: gen/templates/command.go.tmpl (--help-body flag), gen/parser.go (exemples extraits)
  Benefits: Chaque commande est auto-documentée — plus de trial & error sur le format du body
  Tests: ./notion pages update-page-markdown <id> --help-body → affiche un exemple JSON valide
  Commit: feat: add --help-body flag showing example request body from spec

## Phase 5 — Évolutivité

- [ ] T11 GitHub Action : auto-update spec + PR
  Depends on: T19
  Changes: .github/workflows/update-spec.yml
  Benefits: Détecte automatiquement les nouvelles versions de la spec Notion et ouvre une PR
  Tests: Déclencher manuellement via gh workflow run update-spec.yml
  Commit: ci: add weekly spec auto-update workflow

- [ ] T12 Ajouter la commande notion api générique (escape hatch)
  Depends on: T08
  Changes: cmd/api.go (déjà présent, non supprimé)
  Benefits: Appeler n'importe quel endpoint non couvert sans recompiler
  Tests: ./notion api GET /v1/users/me → JSON de l'utilisateur courant
  Commit: feat: restore generic api escape hatch command
