# Notion CLI — Cookbook Human Mode

> Guide de validation de la Phase A (CLI Human).
> Toutes les commandes sont en mode human (défaut) — gum + jq requis.
>
> **Page de test** : `32445368-aa31-80da-ad1a-f9d92a4737a0`
> **User** : `76e85c6a-8d5c-45f3-b578-191e9e676daa` (Jeremie Alcaraz)

---

## Setup

```bash
# Build
just build

# Variables
PAGE_ID="32445368-aa31-80da-ad1a-f9d92a4737a0"
USER_ID="76e85c6a-8d5c-45f3-b578-191e9e676daa"
```

---

## HA01 — Détection gum

```bash
# Avec gum installé → aucun warning
./notion users get-self --dry-run

# Simuler gum absent → tip: install gum...
PATH=$(echo $PATH | tr ':' '\n' | grep -v homebrew | tr '\n' ':') ./notion users get-self --dry-run

# Forcer le mode sans gum
./notion users get-self --no-gum --dry-run
```

**Attendu** : warning discret sur stderr si gum absent, rien sinon.

---

## HA02 — JSON coloré via jq

```bash
# Mode human → JSON coloré (clés bleues, strings vertes)
./notion users get-self

# Pipe → JSON brut (pas de couleur)
./notion users get-self | cat

# Forcer JSON brut
./notion users get-self -f json

# Sans gum → JSON indenté non coloré
./notion users get-self --no-gum
```

**Attendu** : couleurs jq dans le terminal, disparaissent en pipe.

---

## HA03 — Spinner pendant les requêtes

```bash
# Spinner animé ⣾ pendant l'appel API
./notion users get-self

# Plusieurs appels enchaînés — spinner à chaque fois
./notion users get-users
./notion pages retrieve-a-page $PAGE_ID
```

**Attendu** : spinner `⣾ GET /v1/users/me` s'affiche puis disparaît proprement.

---

## HA04 — Wizard path-param manquant

```bash
# Sans argument → prompt "page_id: " avec gum input
./notion pages retrieve-a-page

# Avec argument → direct, pas de prompt
./notion pages retrieve-a-page $PAGE_ID

# Sans gum → erreur claire
./notion pages retrieve-a-page --no-gum
```

**Attendu** : prompt interactif si arg manquant, erreur `missing argument: page_id` sans gum.

---

## HA05 — Wizard body champ par champ

```bash
# Sans --body → wizard champ par champ
./notion search post-search

# Avec --body → direct, pas de wizard
./notion search post-search --body '{"query": "CLI"}'

# --help-body → exemple JSON sans lancer le wizard
./notion pages patch-page --help-body
./notion comments create-a-comment --help-body
./notion databases create-database --help-body
```

**Attendu** : wizard gum input pour chaque champ connu, skip si vide.

---

## HA06 — Confirmation avant DELETE

```bash
# Créer un bloc temporaire
TEMP_BLOCK=$(./notion blocks patch-block-children $PAGE_ID \
  --body '{"children":[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Bloc temporaire à supprimer"}}]}}]}' \
  --field results 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])")

echo "Bloc créé : $TEMP_BLOCK"

# Supprimer avec confirmation gum
./notion blocks delete-a-block $TEMP_BLOCK

# Supprimer sans confirmation (bypass)
./notion blocks delete-a-block $TEMP_BLOCK --yes

# Dry-run → pas de confirmation, affiche la requête
./notion blocks delete-a-block $TEMP_BLOCK --dry-run

# Sans gum → pas de confirmation (exécute direct)
./notion blocks delete-a-block $TEMP_BLOCK --no-gum --yes
```

**Attendu** : `gum confirm "Delete <id>? This cannot be undone."` en TTY. `--yes` et `--dry-run` bypasse.

---

## HA07 — Listes en tables gum

```bash
# Users → table Name / Type / ID
./notion users get-users

# Blocks d'une page → table Type / ID / Created
./notion blocks get-block-children $PAGE_ID

# Comments → table ID / Created / Text
./notion comments list-comments --block-id $PAGE_ID

# Recherche → table Object / ID / Created
./notion search post-search --body '{"query": "CLI"}'

# Forcer JSON brut (bypass table)
./notion users get-users -f json

# Sans gum → JSON indenté
./notion users get-users --no-gum
```

**Attendu** : tableau `gum table` avec bordure arrondie et compteur `N result(s)`.

---

## HA08 — Auth interactif

```bash
# Login interactif → prompt token masqué avec gum input --password
./notion auth login

# Login via pipe (inchangé)
echo "$NOTION_TOKEN" | ./notion auth login --with-token

# Status
./notion auth status

# Doctor
./notion auth doctor

# Switch de profil avec gum choose (si plusieurs profils)
./notion auth switch
```

**Attendu** : `gum input --password` masque la saisie du token. `gum choose` pour la sélection de profil.

---

## Flags globaux à valider

```bash
# --no-gum : désactive tout gum (spinner, colors, wizards, tables)
./notion users get-users --no-gum

# --dry-run : affiche la requête HTTP sans l'exécuter
./notion pages retrieve-a-page $PAGE_ID --dry-run
./notion blocks delete-a-block fake-id --dry-run

# --field : extrait un seul champ top-level
./notion users get-self --field name
./notion users get-self --field id

# -f json : force JSON brut même en TTY
./notion users get-users -f json
```

---

## Checklist de validation

Coche chaque item après test :

- [ ] HA01 — Warning gum absent sur stderr, rien si présent
- [ ] HA02 — JSON coloré en TTY, brut en pipe
- [ ] HA03 — Spinner visible pendant les requêtes
- [ ] HA04 — Prompt gum si arg manquant, erreur claire sans gum
- [ ] HA05 — Wizard body champ par champ sans `--body`
- [ ] HA06 — Confirm gum avant DELETE, `--yes` bypass
- [ ] HA07 — Tables gum pour les listes, `-f json` bypass
- [ ] HA08 — Token masqué à l'auth, `gum choose` pour switch
- [ ] Flags globaux (`--no-gum`, `--dry-run`, `--field`, `-f json`) fonctionnels
