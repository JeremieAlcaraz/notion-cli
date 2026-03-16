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

- **Phase A — CLI Human** (HA01–HA09) : gum, jq coloré, wizards interactifs + validation ✅
- **Phase B — CLI Agent** (AB01–AB05) : mode minifié, JSON compact, erreurs codées ✅
- **Phase C — Benchmark** (BM01–BM03) : mesures réelles de tokens avant/après ✅
- **Phase D — Output avancé** (OA01–OA06) : --fields sur listes, --strip-meta, --summary, NDJSON, Markdown blocs
- **Phase E — Token intelligence** (TI01–TI04) : historique réponses, dry-run enrichi, tips automatiques
- **Phase F — Qualité & pertinence** (QP01–QP05) : mesure de la fidélité des données malgré la réduction de tokens

---

## Phase A — CLI Human (HA01–HA09) ✅

### HA01 — Détection et installation de `gum`
- [x] HA01 Vérifier la présence de gum et afficher un message d'onboarding

### HA02 — Rendu JSON coloré avec jq
- [x] HA02 Piper la sortie JSON dans jq --color-output en mode human

### HA03 — Spinner gum pendant les requêtes HTTP
- [x] HA03 Afficher un spinner gum pendant chaque appel API

### HA04 — Wizard interactif pour les commandes sans argument
- [x] HA04 Lancer un wizard gum input/filter si un path-param est manquant

### HA05 — Wizard pour la saisie du body JSON champ par champ
- [x] HA05 Proposer une saisie guidée des champs du body si --body est absent

### HA06 — Confirmation gum avant les opérations destructives
- [x] HA06 Demander confirmation gum confirm avant DELETE et in_trash=true

### HA07 — Formatage de la réponse : tables et résumés adaptés par ressource
- [x] HA07 Formater les listes (users, databases, pages) en table gum

### HA08 — Commande notion auth login interactive avec gum
- [x] HA08 Wizard d'authentification guidé (gum input pour le token)

### HA09 — Mini cookbook Human + validation manuelle
- [x] HA09 Écrire un mini cookbook dédié au mode human et attendre validation

---

## Phase B — CLI Agent (AB01–AB05) ✅

- [x] AB01 Ajouter le flag --agent (ou env NOTION_AGENT=1)
- [x] AB02 Sortie JSON minifiée (zéro whitespace, zéro couleur)
- [x] AB03 Messages d'erreur ultra-courts `ERR:<code>:<message>`
- [x] AB04 Supprimer spinners, confirmations et wizards en mode agent
- [x] AB05 Flag --fields (liste CSV) pour filtrer les clés de la réponse

---

## Phase C — Benchmark (BM01–BM03) ✅

- [x] BM01 Corpus de référence + script tiktoken
- [x] BM02 Rapport comparatif human vs agent
- [x] BM03 `just bench` + garde de régression

---

## Phase D — Output avancé (OA01–OA06)

Objectif : réduire les tokens sans perte de données fonctionnelles, via des transformations de sortie ciblées.

### OA01 — --fields sur les listes (results[])

- [ ] OA01 Appliquer --fields aux items de results[] quand la réponse est une liste
  Depends on: AB05
  Changes: internal/render/render.go (OutputFields gère object==list), internal/render/list.go
  Benefits: `--fields id,title` sur une liste → -97% tokens (800 → 25 tokens par item)
  Tests:
    `./notion --agent databases list-databases --fields id,title`
    → `[{"id":"abc","title":"Projects"},...]` (pas d'objet list wrapper)
    `./notion --agent users list-all-users --fields id,name,type`
    → tableau compact
  Commit: feat: apply --fields to list results items

### OA02 — --strip-meta : suppression des champs bruités

- [ ] OA02 Supprimer les champs non-fonctionnels des réponses Notion
  Depends on: -
  Changes: internal/render/strip.go (nouveau), internal/render/render.go, cmd/root.go (flag global)
  Fields supprimés: `request_id`, `created_by`, `last_edited_by`, `cover`, `icon`, `public_url`, `in_trash`, `archived` (si false)
  Benefits: ~30% réduction sur chaque objet, zéro perte fonctionnelle pour un agent
  Tests:
    `./notion --agent pages retrieve-a-page <id> --strip-meta`
    → pas de `request_id`, `cover`, `icon`, `created_by` dans la sortie
    `./notion --agent pages retrieve-a-page <id>` (sans flag)
    → sortie complète inchangée
  Commit: feat: --strip-meta flag to remove noisy Notion metadata fields

### OA03 — --summary : vue minimale d'un objet

- [ ] OA03 Retourner uniquement id, title, url, type, parent_id pour pages/databases/blocks
  Depends on: OA02
  Changes: internal/render/summary.go (nouveau), cmd/root.go (flag global)
  Benefits: ~95% réduction (400 → 20 tokens) — suffisant pour navigation et référencement
  Tests:
    `./notion --agent pages retrieve-a-page <id> --summary`
    → `{"id":"...","title":"My Page","url":"...","type":"page","parent_id":"..."}`
    Exactement 5 clés, pas plus
  Commit: feat: --summary flag for minimal object representation

### OA04 — --format ndjson pour les listes

- [ ] OA04 Émettre une ligne JSON par item de results[] au lieu d'un objet list
  Depends on: -
  Changes: internal/render/render.go (branche ndjson), internal/render/list.go
  Benefits: Streamable, grep-able, head -n 1-able ; compatible jq -s pour recomposer
  Tests:
    `./notion --agent users list-all-users --format ndjson | head -1`
    → une ligne JSON valide
    `./notion --agent users list-all-users --format ndjson | wc -l`
    → N (nombre d'utilisateurs)
  Commit: feat: --format ndjson outputs one JSON line per list item

### OA05 — --format md pour les blocs (Markdown natif)

- [ ] OA05 Convertir les blocs Notion en Markdown lisible
  Depends on: -
  Changes: internal/render/markdown.go (nouveau)
  Blocs supportés: heading_1/2/3, paragraph, bulleted_list_item, numbered_list_item,
    toggle, code, quote, divider, callout, to_do, image (alt text)
  Benefits: -93% tokens (300 → 20 tokens pour 3 blocs) ; directement injectable dans un prompt
  Tests:
    `./notion blocks retrieve-block-children <page-id> --format md`
    → sortie Markdown valide (vérifiable avec un parser MD)
    heading_2 "Introduction" → `## Introduction`
    paragraph → texte brut
    bulleted_list_item → `- item`
    code block → ``` fence avec langage
  Commit: feat: --format md renders Notion blocks as Markdown

### OA06 — --all : pagination automatique

- [ ] OA06 Boucler automatiquement sur les pages suivantes si has_more=true
  Depends on: -
  Changes: internal/client/client.go (paginate helper), cmd/root.go (flag global)
  Benefits: 1 appel agent au lieu de N ; zéro logique de boucle côté agent
  Tests:
    `./notion --agent users list-all-users --all`
    → `has_more: false` dans la réponse finale, tous les users présents
    Sans --all : comportement actuel inchangé
  Commit: feat: --all flag for automatic pagination

---

## Phase E — Token intelligence (TI01–TI04)

Objectif : permettre à l'agent d'estimer le coût d'une requête AVANT de l'exécuter, et enregistrer l'historique pour des recommandations informées.

### TI01 — Historique local des réponses

- [ ] TI01 Enregistrer bytes+tokens de chaque réponse dans ~/.cache/notion-cli/history.json
  Depends on: BM01
  Changes: internal/cache/history.go (nouveau), internal/client/client.go (hook post-response)
  Format: `{"method":"GET","path_pattern":"/v1/users","ts":1234567890,"bytes":467,"tokens":178}`
  TTL: 7 jours, max 500 entrées, rotation FIFO
  Benefits: Source de données pour les estimations dry-run — basée sur des appels réels, pas des moyennes fixes
  Tests:
    `./notion users get-self` puis `cat ~/.cache/notion-cli/history.json | jq .`
    → entrée avec bytes et tokens pour GET /v1/users/me
  Commit: feat: record response size history for token estimation

### TI02 — --dry-run enrichi avec estimation tokens

- [ ] TI02 Afficher estimation tokens dans --dry-run si historique disponible
  Depends on: TI01
  Changes: internal/client/client.go (dry-run output), internal/cache/history.go (lookup)
  Sortie enrichie:
    ```
    DRY RUN: GET /v1/users/me
    Estimated response: ~178 tokens (avg of 3 similar calls)
    ```
    Si pas d'historique: `no estimate available (run once without --dry-run first)`
  Benefits: L'agent peut décider si une requête est dans son budget de contexte avant de l'exécuter
  Tests:
    `./notion users get-self` (1x sans dry-run)
    `./notion --dry-run users get-self` → affiche estimation
  Commit: feat: enrich --dry-run with token count estimation

### TI03 — Estimation avec flags de filtrage appliqués

- [ ] TI03 Calculer l'estimation post-filtrage si --fields ou --strip-meta sont présents
  Depends on: TI02, OA01, OA02
  Changes: internal/cache/history.go (estimate with filter simulation)
  Logique: récupérer la dernière réponse brute du cache, appliquer le filtre, compter les tokens
  Sortie enrichie:
    ```
    DRY RUN: GET /v1/databases
    Estimated response (raw):   ~800 tokens
    Estimated with --fields id,title: ~25 tokens  (-97%)
    ```
  Benefits: L'agent choisit le bon set de flags avant d'exécuter
  Tests:
    `./notion --dry-run --agent databases list-databases --fields id,title`
    → affiche les deux estimations
  Commit: feat: dry-run estimation accounts for --fields and --strip-meta

### TI04 — Tips automatiques si réponse > seuil

- [ ] TI04 Suggérer des flags de réduction si la réponse dépasse 500 tokens
  Depends on: TI01, OA01, OA02, OA03
  Changes: internal/render/render.go (post-render tip), internal/cache/history.go
  Tip émis sur stderr uniquement en mode human (jamais en mode agent) :
    ```
    Tip: response was 1 200 tokens — try --fields id,title (~80 tokens) or --summary (~20 tokens)
    ```
  Benefits: Discovery progressive des optimisations sans RTFM
  Tests:
    Appeler une commande qui retourne une grosse réponse en mode human
    → tip affiché sur stderr
    Même commande en mode agent → aucun tip (stderr silencieux)
  Commit: feat: suggest token-reduction flags for large responses

---

## Phase F — Qualité & pertinence (QP01–QP05)

Objectif : **prouver** que la réduction de tokens ne dégrade pas la qualité des réponses pour un agent.
Principe : définir des tâches agent représentatives, mesurer le taux de succès avec et sans optimisations.

### QP01 — Définition des scénarios de test agent

- [ ] QP01 Écrire 10 scénarios agent avec critères de succès objectifs
  Depends on: -
  Changes: bench/quality/scenarios.yaml (nouveau)
  Format par scénario:
    ```yaml
    - id: SC01
      description: "Trouver l'ID d'une base de données par son titre"
      command: "notion --agent databases list-databases --fields id,title"
      success_criteria:
        - "La réponse contient le champ id pour chaque base"
        - "La réponse contient le champ title pour chaque base"
        - "L'agent peut identifier la bonne base sans appel supplémentaire"
      data_required: [id, title]
      data_not_required: [created_time, cover, icon, request_id]
    ```
  Scénarios couverts: lookup par titre, navigation parent→enfant, lecture contenu page,
    création page avec propriétés, mise à jour statut, listing users, recherche full-text,
    ajout commentaire, archivage page, upload fichier
  Benefits: Référentiel objectif — "optimisé" ne signifie pas "dégradé"
  Tests: Fichier YAML valide, 10 scénarios, chaque scénario a id+command+success_criteria
  Commit: dev: define 10 agent quality scenarios with success criteria

### QP02 — Script d'évaluation automatique des scénarios

- [ ] QP02 Exécuter chaque scénario et vérifier les critères de succès programmatiquement
  Depends on: QP01
  Changes: bench/quality/eval.py (nouveau)
  Logique:
    - Exécuter la commande du scénario
    - Parser la sortie JSON/NDJSON/MD
    - Vérifier que `data_required` est présent dans la sortie
    - Vérifier que la tâche décrite est réalisable avec les données reçues
    - Score: PASS / FAIL / PARTIAL par scénario
  Benefits: Détection automatique de régression de pertinence (pas seulement de tokens)
  Tests:
    `bench/.venv/bin/python bench/quality/eval.py`
    → tableau PASS/FAIL par scénario, score global X/10
  Commit: dev: add quality evaluation script for agent scenarios

### QP03 — Rapport qualité × efficacité

- [ ] QP03 Générer un rapport croisé tokens économisés vs taux de succès
  Depends on: QP02, BM02
  Changes: bench/quality/report.py (nouveau), bench/quality/results/report.md (généré)
  Format du rapport:
    | Scénario | Mode | Tokens | Taux succès | Verdict |
    |---|---|---|---|---|
    | SC01 list databases | raw | 800 | 100% | baseline |
    | SC01 list databases | --fields id,title | 25 | 100% | ✅ optimisé |
    | SC03 lire contenu page | raw | 3000 | 100% | baseline |
    | SC03 lire contenu page | --format md | 200 | 100% | ✅ optimisé |
    | SC03 lire contenu page | --fields id | 10 | 0% | ❌ trop agressif |
  Benefits: Visualise exactement où l'on économise des tokens SANS perdre en pertinence
  Tests:
    `bench/.venv/bin/python bench/quality/report.py`
    → report.md avec tableau complet, colonnes tokens + succès
  Commit: dev: add quality×efficiency cross report

### QP04 — Intégration dans `just bench`

- [ ] QP04 Ajouter `just bench-quality` qui exécute QP02 + QP03
  Depends on: QP02, QP03
  Changes: justfile (cible bench-quality), bench/quality/run_all.sh
  Sortie:
    ```
    ==> Running quality scenarios...
    SC01 PASS  SC02 PASS  SC03 PASS  ...
    Quality score: 10/10
    ==> Generating quality×efficiency report...
    Report saved to bench/quality/results/report.md
    ```
  Garde de régression : si score < 8/10 → exit 1
  Benefits: Toute PR qui dégrade la pertinence est bloquée, pas seulement celle qui augmente les tokens
  Tests: `just bench-quality` → passe avec score 10/10
  Commit: dev: add just bench-quality target with regression guard

### QP05 — Documentation de la matrice optimisation × pertinence

- [ ] QP05 Documenter quels flags sont sûrs, lesquels sont risqués, lesquels sont contextuels
  Depends on: QP03
  Changes: docs/agent-optimization-matrix.md (nouveau)
  Contenu:
    | Flag | Réduction tokens | Risque pertinence | Cas d'usage recommandé |
    |---|---|---|---|
    | --agent | ~30% | Aucun | Toujours en contexte agent |
    | --fields id,title | ~97% | Faible (navigation) | Lookup, listing |
    | --strip-meta | ~30% | Aucun | Toujours safe |
    | --summary | ~95% | Moyen (lecture contenu) | Navigation uniquement |
    | --format md | ~93% | Faible (blocs texte) | Lecture contenu page |
    | --format ndjson | 0% tokens, +ergonomie | Aucun | Streaming, grep |
    | --all | 0% tokens, -appels | Aucun | Toujours si has_more |
  Benefits: Guide décisionnel pour les agents et les développeurs qui intègrent la CLI
  Tests: Fichier présent, tableau complet, chaque flag documenté
  Commit: docs: add agent optimization matrix with safety ratings

---

## Résumé des phases

| Phase | Tasks | Status | Impact |
|---|---|---|---|
| A — CLI Human | HA01–HA09 | ✅ | UX terminal |
| B — CLI Agent | AB01–AB05 | ✅ | JSON compact |
| C — Benchmark | BM01–BM03 | ✅ | Mesures tokens |
| D — Output avancé | OA01–OA06 | ⏳ | -30% à -97% tokens |
| E — Token intelligence | TI01–TI04 | ⏳ | Estimation pré-exécution |
| F — Qualité & pertinence | QP01–QP05 | ⏳ | Valider sans dégradation |
