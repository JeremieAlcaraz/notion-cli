# Tasks Plan — Notion CLI : Human & Agent Modes

## Objectif

Deux modes d'interface pour le même binaire `notion` :
- **`--human` (ou défaut)** : sortie riche avec `gum` + `jq` coloré, interactions guidées, UX terminal soignée
- **`--agent`** : sortie JSON minifiée, zéro décoration, messages d'erreur ultra-courts — optimisé pour consommation par LLM

## Architecture cible

```
notion [--agent] <resource> <operation> [flags]
```

Le flag global `--agent` (ou variable `NOTION_AGENT=1`) bascule tout le rendu en mode AI.
En mode human (défaut), chaque commande peut lancer un wizard `gum` si aucun argument n'est fourni.

## Spec source

- Spec snapshottée : `spec/notion-openapi.json`
- Binaire `gum` (Charmbracelet) — à installer via Homebrew ou détection automatique

## Phases

- **Phase A — CLI Human** (HA01–HA09) : gum, jq coloré, wizards interactifs + validation
- **Phase B — CLI Agent** (AB01–AB05) : mode minifié, JSON compact, erreurs codées
- **Phase C — Benchmark** (BM01–BM03) : mesures réelles de tokens avant/après

---

## Phase A — CLI Human (HA01–HA08)

### HA01 — Détection et installation de `gum`

- [x] HA01 Vérifier la présence de gum et afficher un message d'onboarding
  Depends on: -
  Changes: internal/tui/gum.go (nouveau package), cmd/root.go
  Benefits: Le mode human dégrade gracieusement si gum est absent (fallback texte brut)
  Tests: `./notion --help` sans gum → warning discret ; avec gum → pas de warning
  Commit: feat: add gum availability check with graceful fallback

### HA02 — Rendu JSON coloré avec jq

- [x] HA02 Piper la sortie JSON dans jq --color-output en mode human
  Depends on: HA01
  Changes: internal/render/render.go (OutputField), internal/tui/jq.go
  Benefits: Toute réponse API est lisible instantanément, sans install manuelle de jq
  Tests: `./notion users get-self` → JSON coloré ; `./notion users get-self -f json` → JSON brut
  Commit: feat: pipe JSON output through jq in human mode

### HA03 — Spinner gum pendant les requêtes HTTP

- [x] HA03 Afficher un spinner gum pendant chaque appel API
  Depends on: HA01
  Changes: internal/client/client.go (do()), internal/tui/spinner.go
  Benefits: Feedback visuel immédiat — l'utilisateur sait que la CLI travaille
  Tests: `./notion databases retrieve-database <id>` → spinner s'affiche puis disparaît
  Commit: feat: show gum spinner during HTTP requests

### HA04 — Wizard interactif pour les commandes sans argument

- [x] HA04 Lancer un wizard gum input/filter si un path-param est manquant
  Depends on: HA01
  Changes: gen/templates/command.go.tmpl (wizard hook), internal/tui/wizard.go
  Benefits: `./notion pages retrieve-a-page` sans ID → gum input "Page ID :" plutôt qu'une erreur
  Tests: `./notion pages retrieve-a-page` (sans arg) → prompt interactif ; avec arg → direct
  Commit: feat: interactive gum wizard for missing path params

### HA05 — Wizard pour la saisie du body JSON champ par champ

- [x] HA05 Proposer une saisie guidée des champs du body si --body est absent
  Depends on: HA04
  Changes: internal/tui/body_wizard.go, gen/parser.go (BodyProp déjà présents), gen/templates/command.go.tmpl
  Benefits: Plus besoin de mémoriser le format JSON — la CLI guide champ par champ
  Tests: `./notion comments create-a-comment` sans --body → wizard avec champs rich_text, parent
  Commit: feat: guided body wizard for JSON fields via gum

### HA06 — Confirmation gum avant les opérations destructives

- [x] HA06 Demander confirmation gum confirm avant DELETE et in_trash=true
  Depends on: HA01
  Changes: gen/templates/command.go.tmpl (detect DELETE), internal/tui/confirm.go
  Benefits: Protection contre les suppressions accidentelles en mode interactif
  Tests: `./notion blocks delete-a-block <id>` → "Supprimer ce bloc ? [y/N]" ; --yes bypasse
  Commit: feat: gum confirm prompt before destructive operations

### HA07 — Formatage de la réponse : tables et résumés adaptés par ressource

- [ ] HA07 Formater les listes (users, databases, pages) en table gum
  Depends on: HA02
  Changes: internal/render/table.go (nouveau), internal/render/render.go
  Benefits: `./notion users get-users` → table avec colonnes Name / Type / ID plutôt que JSON brut
  Tests: `./notion users get-users` → table ; `./notion users get-users -f json` → JSON brut
  Commit: feat: render list responses as gum tables in human mode

### HA08 — Commande notion auth login interactive avec gum

- [ ] HA08 Wizard d'authentification guidé (gum input pour le token)
  Depends on: HA01
  Changes: cmd/auth.go (déjà présent ?), internal/tui/auth_wizard.go
  Benefits: `./notion auth login` sans flag → prompt gum masqué pour le token
  Tests: `./notion auth login` → prompt token masqué ; token stocké dans config
  Commit: feat: interactive gum auth login wizard

### HA09 — Mini cookbook Human + validation manuelle

- [ ] HA09 Écrire un mini cookbook dédié au mode human et attendre validation
  Depends on: HA08
  Changes: COOKBOOK-HUMAN.md (nouveau)
  Benefits: Valider que l'UX human est cohérente et agréable avant d'attaquer la Phase B
  Tests: Validation manuelle par l'utilisateur — toutes les commandes du cookbook passent
  Commit: docs: add human mode cookbook for manual validation

---

## Phase B — CLI Agent (AB01–AB05)

### AB01 — Flag global `--agent` et variable NOTION_AGENT

- [ ] AB01 Ajouter le flag --agent (ou env NOTION_AGENT=1) qui bascule tout le rendu
  Depends on: -
  Changes: cmd/root.go, internal/render/render.go, internal/tui/mode.go (nouveau)
  Benefits: Un seul flag change tout le comportement ; compatible scripts et MCP servers
  Tests: `NOTION_AGENT=1 ./notion users get-self` → JSON minifié une ligne
  Commit: feat: add --agent flag and NOTION_AGENT env for AI mode

### AB02 — Sortie JSON minifiée (zéro whitespace, zéro couleur)

- [ ] AB02 En mode agent : json.Marshal compact sans indentation ni ANSI
  Depends on: AB01
  Changes: internal/render/render.go (branche agent dans OutputField)
  Benefits: Réduction ~40% des tokens par rapport au JSON indenté
  Tests: `./notion --agent users get-self | wc -c` < `./notion users get-self | wc -c`
  Commit: feat: minified JSON output in agent mode

### AB03 — Messages d'erreur ultra-courts en mode agent

- [ ] AB03 En mode agent : erreurs au format `ERR:<code>:<message_court>` sans hint
  Depends on: AB01
  Changes: internal/client/client.go (errorHint skipped), internal/render/errors.go
  Benefits: L'AI reçoit un code parsable, pas un paragraphe explicatif
  Tests: `./notion --agent pages retrieve-a-page invalid-id` → `ERR:object_not_found:page not found`
  Commit: feat: terse error format in agent mode

### AB04 — Suppression des spinners, confirmations et wizards en mode agent

- [ ] AB04 Désactiver toute sortie stderr décorative en mode agent
  Depends on: AB01, HA03, HA06
  Changes: internal/tui/gum.go (IsAgentMode guard), internal/tui/spinner.go, internal/tui/confirm.go
  Benefits: Zéro bruit parasite — stdout = données pures, stderr = vide
  Tests: `./notion --agent blocks delete-a-block <id> 2>/dev/null` → pas de confirm, exécution directe
  Commit: feat: suppress all decorative output in agent mode

### AB05 — Flag --fields (liste CSV) pour filtrer les clés de la réponse

- [ ] AB05 Ajouter --fields id,name,url pour retourner un sous-objet JSON minimal
  Depends on: AB02
  Changes: internal/render/render.go (OutputFields multi-key), cmd/root.go (flag global)
  Benefits: L'AI ne reçoit que ce dont elle a besoin — économie maximale de tokens
  Tests: `./notion --agent users get-self --fields id,name` → `{"id":"...","name":"..."}`
  Commit: feat: --fields flag to select response keys in agent mode

---

## Phase C — Benchmark (BM01–BM03)

Mesures réelles de tokens sur un corpus fixe de 10 appels API représentatifs.
Outil de comptage : `tiktoken` (cl100k_base, modèle GPT-4/Claude compatible) via script Python.

### BM01 — Corpus de référence et script de mesure

- [ ] BM01 Créer le corpus de benchmark et le script de comptage de tokens
  Depends on: -
  Changes: bench/corpus.sh (10 appels API fixés), bench/count_tokens.py (tiktoken), bench/README.md
  Benefits: Baseline reproductible — chaque PR peut comparer ses chiffres au même corpus
  Tests: `just bench` → tableau CSV `mode,command,bytes,tokens` dans bench/results/baseline.csv
  Commit: dev: add token benchmark corpus and counting script

### BM02 — Mesure comparative human vs agent sur le corpus

- [ ] BM02 Exécuter le corpus en mode human et agent, générer un rapport diff
  Depends on: BM01, AB02, AB05
  Changes: bench/compare.sh, bench/results/report.md (généré)
  Benefits: Chiffres réels : bytes économisés, tokens économisés, % de réduction par commande
  Tests: `just bench-compare` → report.md avec tableau avant/après pour les 10 commandes
  Commit: dev: add human vs agent benchmark comparison report

### BM03 — Intégration du benchmark dans le justfile + CI check

- [ ] BM03 Ajouter `just bench` dans le justfile et un check de régression (tokens ne doivent pas augmenter)
  Depends on: BM02
  Changes: justfile (cible bench, bench-compare), bench/check_regression.py
  Benefits: Toute modif du renderer qui augmente les tokens en mode agent est détectée automatiquement
  Tests: `just bench` passe ; modifier render.go pour ajouter un espace → `just bench` échoue
  Commit: dev: add bench target to justfile with regression guard
