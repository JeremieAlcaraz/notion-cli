# tasks-plan.md — notion-cli fork setup & GoReleaser

## Context

Fork de [4ier/notion-cli](https://github.com/4ier/notion-cli) par **JeremieAlcaraz**.

**Objectifs du fork :**
- Refonte via OpenAPI generator (génération des commandes depuis `spec/notion-openapi.json`)
- Mode **agent** : réduction max des tokens (output JSON compact, pas de couleur, headers minimaux)
- Mode **human** : UI interactive avec `gum` (selects, confirms, spinners)
- Distribution via Homebrew (binaire autonome, dépendances embarquées ou gérées par brew)

---

## Phase 1 — Repo cleanup & GitHub setup

- [ ] T01 Nettoyer le .gitignore et supprimer les fichiers parasites
  Depends on: -
  Changes: `.gitignore`, suppression de `notion` (binaire local) et `notion-cli` (ancien binaire)
  Benefits: repo propre, aucun binaire tracké
  Tests: `git status` ne montre aucun fichier non-voulu ; `git ls-files | grep -E '^notion$|^notion-cli$'` retourne vide
  Commit: chore: remove committed binaries and clean gitignore

- [ ] T02 Renommer le remote origin en upstream
  Depends on: T01
  Changes: `.git/config` (local uniquement)
  Benefits: garde la référence au repo original 4ier sans ambiguïté
  Tests: `git remote -v` affiche `upstream https://github.com/4ier/notion-cli`
  Commit: - (config locale, pas de commit)

- [ ] T03 Créer le repo GitHub `notion-cli` sous JeremieAlcaraz et le setter comme origin
  Depends on: T02
  Changes: `.git/config` (local uniquement)
  Benefits: le fork a son propre remote GitHub
  Tests: `git remote -v` affiche `origin https://github.com/JeremieAlcaraz/notion-cli` + `upstream`
  Commit: - (config locale, pas de commit)

- [ ] T04 Mettre à jour le module Go path
  Depends on: T03
  Changes: `go.mod`, `cmd/*.go`, `main.go` (import paths `github.com/4ier/notion-cli` → `github.com/JeremieAlcaraz/notion-cli`)
  Benefits: le module Go pointe sur le bon repo
  Tests: `go build ./...` passe sans erreur
  Commit: chore: update Go module path to JeremieAlcaraz/notion-cli

- [ ] T05 Réécrire le README.md
  Depends on: T04
  Changes: `README.md`
  Benefits: crédits clairs au créateur original, objectifs du fork documentés, badges corrects
  Tests: lecture humaine ; lien upstream et lien fork sont valides
  Commit: docs: rewrite README with fork credits and project goals

---

## Phase 2 — GoReleaser + Homebrew tap

- [ ] T06 Mettre à jour .goreleaser.yaml aux coordonnées du fork
  Depends on: T05
  Changes: `.goreleaser.yaml` (owner, module path, maintainer)
  Benefits: les builds et releases pointent sur JeremieAlcaraz, plus sur 4ier
  Tests: `goreleaser check` passe
  Commit: chore(goreleaser): update project owner to JeremieAlcaraz

- [ ] T07 Créer le repo `homebrew-tap` sur GitHub avec la formula scaffold
  Depends on: T06
  Changes: nouveau repo GitHub `homebrew-tap` + `Formula/notion-cli.rb` minimal
  Benefits: `brew install JeremieAlcaraz/tap/notion-cli` possible après la première release
  Tests: `gh repo view JeremieAlcaraz/homebrew-tap` répond OK
  Commit: feat(tap): initial homebrew formula scaffold

- [ ] T08 Wirer GoReleaser sur le nouveau tap + gérer la dépendance gum
  Depends on: T07
  Changes: `.goreleaser.yaml` (section `brews` : owner, tap repo, `dependencies: [gum]`)
  Benefits: GoReleaser met à jour la formula automatiquement à chaque release ; brew installe gum automatiquement
  Tests: `goreleaser check` passe ; section brews bien configurée
  Commit: chore(goreleaser): wire brew tap with gum dependency

- [ ] T09 Mettre à jour le GitHub Actions workflow release
  Depends on: T08
  Changes: `.github/workflows/release.yml` (secrets, owner, token name)
  Benefits: la CI crée les releases et met à jour le tap automatiquement
  Tests: YAML valide syntaxiquement
  Commit: ci: update release workflow for fork

- [ ] T10 Push main vers origin + créer la première release v0.1.0-fork
  Depends on: T09
  Changes: push de `main` + push de `feature/openapi-driven-generation` + tag `v0.1.0-fork`
  Benefits: première release fonctionnelle disponible sur brew
  Tests: `gh release view v0.1.0-fork` montre les assets binaires
  Commit: - (tag + push, pas un commit)

---

## Phase 3 — Mode agent & mode human (futures tâches)

> Détaillées après validation de la Phase 2.

- [ ] T11 [FUTURE] Commande `notion skill install` avec go:embed
  Depends on: T10
  Changes: `cmd/skill.go`, `skills/notion-cli/SKILL.md` embarqué via `go:embed`, `.goreleaser.yaml` (message post_install brew)
  Benefits: le skill est toujours en sync avec la version installée ; fonctionne brew/go install/binaire direct
  Tests: `notion skill install --claude` crée `~/.claude/skills/notion-cli/SKILL.md` ; `notion skill install --codex` idem pour codex
  Commit: feat(skill): add `notion skill install` command with embedded SKILL.md

- [ ] T12 [FUTURE] Flag `--agent` : output JSON compact, no-color, headers supprimés
  Depends on: T10
  Note: réduit drastiquement les tokens consommés par les LLMs

- [ ] T13 [FUTURE] Flag `--human` : UI gum (search interactif, db query avec select, confirms)
  Depends on: T10
  Note: `gum` géré comme dépendance brew dans la formula (installé auto)

- [ ] T14 [FUTURE] Refonte OpenAPI : génération des commandes depuis spec/notion-openapi.json
  Depends on: T12, T13

---

## Notes

### Dépendances binaire brew

Le binaire Go est **statiquement compilé** (CGO_ENABLED=0) → aucune dépendance runtime.

| Outil | Statut | Gestion |
|-------|--------|---------|
| `gum` | Mode human (optionnel dans un premier temps) | `depends_on "charmbracelet/tap/gum"` dans la formula |
| `jq` | Usage scripts utilisateurs, pas dans le binaire | Documenté comme optionnel |
| Tout le reste | Compilé dans le binaire | Rien à faire |

### Branches

- `main` → branche principale du fork
- `feature/openapi-driven-generation` → branche de dev active (continuée après T10)
- Anciennes branches `remotes/origin/*` → deviennent `remotes/upstream/*` après T02
