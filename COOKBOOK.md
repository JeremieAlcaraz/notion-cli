# Notion CLI — Cookbook

> Guide complet copy-paste pour tester toutes les ressources de l'API Notion via la CLI.
>
> **Page de test** : `32445368-aa31-80da-ad1a-f9d92a4737a0`
> **Bot** : `cli-demo-test` (ID `32445368-aa31-81c6-a900-0027e23ceb66`)

---

## Sommaire

1. [Users](#1-users)
2. [Pages](#2-pages)
3. [Blocks](#3-blocks)
4. [Comments](#4-comments)
5. [Databases](#5-databases)
6. [Data Sources](#6-data-sources)
7. [Search](#7-search)
8. [File Uploads](#8-file-uploads)
9. [API générique](#9-api-générique-escape-hatch)

---

## Variables à définir

Avant de commencer, exporte ces variables dans ton shell :

```bash
PAGE_ID="32445368-aa31-80da-ad1a-f9d92a4737a0"
BOT_ID="32445368-aa31-81c6-a900-0027e23ceb66"
USER_ID="76e85c6a-8d5c-45f3-b578-191e9e676daa"          # Jeremie Alcaraz
DATA_SOURCE_ID="cf056b30-f41c-426b-9c80-d250cc583d0f"    # DB pour data_sources
```

---

## 1. Users

### Lire le bot courant (GET /v1/users/me)
```bash
./notion users get-self
```

### Extraire juste le nom
```bash
./notion users get-self --field name
```

### Lister tous les membres du workspace (GET /v1/users)
```bash
./notion users get-users
```

### Récupérer un utilisateur par ID (GET /v1/users/{id})
```bash
./notion users get-user $USER_ID
```

---

## 2. Pages

### Récupérer une page (GET /v1/pages/{id})
```bash
./notion pages retrieve-a-page $PAGE_ID
```

### Récupérer une propriété spécifique (GET /v1/pages/{id}/properties/{prop_id})
```bash
./notion pages retrieve-a-page-property $PAGE_ID title
```

### Lire le contenu en Markdown (GET /v1/pages/{id}/markdown)
```bash
./notion pages retrieve-page-markdown $PAGE_ID
```

### Mettre à jour le titre (PATCH /v1/pages/{id})
```bash
./notion pages patch-page $PAGE_ID --body '{
  "properties": {
    "title": {
      "title": [{"text": {"content": "Demo Page CLI — mise à jour via CLI ✅"}}]
    }
  }
}'
```

### Créer une sous-page (POST /v1/pages)
```bash
NEW_PAGE_ID=$(./notion pages post-page --body '{
  "parent": {"page_id": "'"$PAGE_ID"'"},
  "properties": {
    "title": {
      "title": [{"text": {"content": "Sous-page créée via CLI"}}]
    }
  }
}' --field id)

echo "Nouvelle page : $NEW_PAGE_ID"
```

### Déplacer la page vers un nouveau parent (PATCH /v1/pages/{id}/move)
```bash
# Déplacer la sous-page vers la même page (ou une autre)
./notion pages move-page $NEW_PAGE_ID --body '{
  "parent": {"page_id": "'"$PAGE_ID"'"}
}'
```

### Insérer du contenu Markdown (POST /v1/pages/{id}/markdown)
```bash
./notion pages update-page-markdown $PAGE_ID --body '{
  "type": "insert_content",
  "insert_content": {
    "content": "## Section ajoutée via CLI\n\nContenu inséré depuis le terminal avec `update-page-markdown`."
  }
}'
```

### Mettre à la corbeille (PATCH /v1/pages/{id})
```bash
./notion pages patch-page $NEW_PAGE_ID --body '{"in_trash": true}'
```

---

## 3. Blocks

### Récupérer un bloc (GET /v1/blocks/{id})
```bash
BLOCK_ID="32445368-aa31-8029-ba6d-feb8558afc6d"
./notion blocks retrieve-a-block $BLOCK_ID
```

### Lister les enfants d'un bloc (GET /v1/blocks/{id}/children)
```bash
./notion blocks get-block-children $PAGE_ID
```

### Créer des blocs enfants (PATCH /v1/blocks/{id}/children)
```bash
./notion blocks patch-block-children $PAGE_ID --body '{
  "children": [
    {
      "object": "block",
      "type": "heading_2",
      "heading_2": {
        "rich_text": [{"type": "text", "text": {"content": "Section créée via CLI"}}]
      }
    },
    {
      "object": "block",
      "type": "paragraph",
      "paragraph": {
        "rich_text": [{"type": "text", "text": {"content": "Paragraphe ajouté depuis le terminal."}}]
      }
    },
    {
      "object": "block",
      "type": "callout",
      "callout": {
        "rich_text": [{"type": "text", "text": {"content": "💡 Callout créé via la CLI Notion"}}],
        "icon": {"type": "emoji", "emoji": "💡"}
      }
    },
    {
      "object": "block",
      "type": "bulleted_list_item",
      "bulleted_list_item": {
        "rich_text": [{"type": "text", "text": {"content": "Item de liste"}}]
      }
    },
    {
      "object": "block",
      "type": "to_do",
      "to_do": {
        "rich_text": [{"type": "text", "text": {"content": "Tâche à faire"}}],
        "checked": false
      }
    }
  ]
}'
```

### Embed une vidéo YouTube dans un bloc
```bash
./notion blocks patch-block-children $PAGE_ID --body '{
  "children": [
    {
      "object": "block",
      "type": "video",
      "video": {
        "type": "external",
        "external": {
          "url": "https://www.youtube.com/watch?v=nQ3YqEVeqHg"
        }
      }
    }
  ]
}'
```

### Mettre à jour le contenu d'un bloc (PATCH /v1/blocks/{id})
```bash
./notion blocks update-a-block $BLOCK_ID --body '{
  "paragraph": {
    "rich_text": [{"type": "text", "text": {"content": "Contenu mis à jour via CLI ✅"}}]
  }
}'
```

### Supprimer un bloc (DELETE /v1/blocks/{id})
```bash
# Crée d'abord un bloc temporaire pour le supprimer
TEMP_BLOCK_ID=$(./notion blocks patch-block-children $PAGE_ID --body '{
  "children": [{
    "object": "block",
    "type": "paragraph",
    "paragraph": {"rich_text": [{"type": "text", "text": {"content": "Bloc temporaire — sera supprimé"}}]}
  }]
}' | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['results'][0]['id'])")

echo "Bloc temporaire : $TEMP_BLOCK_ID"
./notion blocks delete-a-block $TEMP_BLOCK_ID
```

### Prévisualiser une suppression sans l'exécuter (--dry-run)
```bash
./notion blocks delete-a-block $BLOCK_ID --dry-run
```

---

## 4. Comments

### Lister les commentaires d'une page (GET /v1/comments)
```bash
./notion comments list-comments --block-id $PAGE_ID
```

### Récupérer un commentaire par ID (GET /v1/comments/{id})
```bash
COMMENT_ID="32445368-aa31-8125-914a-001d5bbccfc1"
./notion comments retrieve-comment $COMMENT_ID
```

### Créer un commentaire (POST /v1/comments)
```bash
./notion comments create-a-comment --body '{
  "parent": {"page_id": "'"$PAGE_ID"'"},
  "rich_text": [{"type": "text", "text": {"content": "Commentaire ajouté via la CLI ✅"}}]
}'
```

---

## 5. Databases

### Créer une database (POST /v1/databases)
```bash
DB_ID=$(./notion databases create-database --body '{
  "parent": {"type": "page_id", "page_id": "'"$PAGE_ID"'"},
  "title": [{"type": "text", "text": {"content": "Base de données CLI"}}],
  "properties": {
    "Nom": {"title": {}},
    "Statut": {
      "select": {
        "options": [
          {"name": "À faire", "color": "red"},
          {"name": "En cours", "color": "yellow"},
          {"name": "Terminé", "color": "green"}
        ]
      }
    },
    "Priorité": {"number": {"format": "number"}},
    "Date": {"date": {}}
  }
}' --field id)

echo "Database créée : $DB_ID"
```

### Récupérer une database (GET /v1/databases/{id})
```bash
./notion databases retrieve-database $DB_ID
```

### Mettre à jour la database (PATCH /v1/databases/{id})
```bash
./notion databases update-database $DB_ID --body '{
  "title": [{"type": "text", "text": {"content": "Base de données CLI ✅ modifiée"}}]
}'
```

---

## 6. Data Sources

> **Important** : `data_source_id ≠ database_id`.
> Pour lister tes data sources : `./notion search post-search --body '{"filter":{"value":"data_source","property":"object"}}'`

```bash
DS_ID="cf056b30-f41c-426b-9c80-d250cc583d0f"
```

### Récupérer une data source (GET /v1/data_sources/{id})
```bash
./notion data-sources retrieve-a-data-source $DS_ID
```

### Mettre à jour une data source (PATCH /v1/data_sources/{id})
```bash
./notion data-sources update-a-data-source $DS_ID --body '{
  "title": [{"type": "text", "text": {"content": "DS mise à jour via CLI ✅"}}]
}'
```

### Requêter les entrées d'une data source (POST /v1/data_sources/{id}/query)
```bash
./notion data-sources post-database-query $DS_ID --body '{
  "page_size": 10
}'
```

### Lister les templates d'une data source (GET /v1/data_sources/{id}/templates)
```bash
./notion data-sources list-data-source-templates $DS_ID
```

### Créer une data source (POST /v1/data_sources)
```bash
./notion data-sources create-a-database --body '{
  "parent": {"type": "database_id", "database_id": "'"$DS_ID"'"},
  "title": [{"type": "text", "text": {"content": "Nouvelle DS via CLI"}}],
  "properties": {
    "Nom": {"title": {}},
    "Tags": {"multi_select": {"options": []}}
  }
}'
```

---

## 7. Search

### Recherche globale (POST /v1/search)
```bash
./notion search post-search --body '{"query": "CLI"}'
```

### Filtrer sur les pages uniquement
```bash
./notion search post-search --body '{
  "query": "Demo",
  "filter": {"value": "page", "property": "object"}
}'
```

### Filtrer sur les data sources uniquement
```bash
./notion search post-search --body '{
  "filter": {"value": "data_source", "property": "object"}
}'
```

### Via stdin (évite les guillemets)
```bash
echo '{"query": "CLI", "filter": {"value": "page", "property": "object"}}' \
  | ./notion search post-search --body -
```

### Via fichier JSON
```bash
cat > /tmp/search.json << 'EOF'
{
  "query": "Demo",
  "filter": {"value": "page", "property": "object"},
  "page_size": 5
}
EOF
./notion search post-search --body @/tmp/search.json
```

---

## 8. File Uploads

### Uploader une image (flow complet automatique)
```bash
./notion file-uploads upload-file img-to-send/img.jpg
```

### Uploader et attacher directement à la page de test
```bash
./notion file-uploads upload-file img-to-send/img.jpg --page-id $PAGE_ID
```

### Créer un slot manuellement (POST /v1/file_uploads)
```bash
UPLOAD_ID=$(./notion file-uploads create-file --body '{
  "name": "img.jpg",
  "content_type": "image/jpeg"
}' --field id)

echo "Upload slot : $UPLOAD_ID"
```

### Lister les uploads récents (GET /v1/file_uploads)
```bash
./notion file-uploads list-file-uploads
```

### Filtrer par statut
```bash
./notion file-uploads list-file-uploads --status uploaded
./notion file-uploads list-file-uploads --status pending
```

### Récupérer le statut d'un upload (GET /v1/file_uploads/{id})
```bash
./notion file-uploads retrieve-file-upload $UPLOAD_ID
```

### Injecter une image uploadée dans un bloc
```bash
# 1. Uploader l'image et récupérer son ID
IMG_UPLOAD_ID=$(./notion file-uploads upload-file img-to-send/img.jpg --field id 2>/dev/null || \
  ./notion file-uploads upload-file img-to-send/img.jpg | grep '"id"' | head -1 | tr -d ' "id:,')

# 2. L'injecter comme bloc image dans la page
./notion blocks patch-block-children $PAGE_ID --body '{
  "children": [{
    "object": "block",
    "type": "image",
    "image": {
      "type": "file_upload",
      "file_upload": {"id": "'"$IMG_UPLOAD_ID"'"}
    }
  }]
}'
```

### Grille d'images 2×3 (column_list → column → image)
```bash
# Uploader 6 fois la même image pour simuler une galerie
for i in 1 2 3 4 5 6; do
  IDS[$i]=$(./notion file-uploads upload-file img-to-send/img.jpg 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
  echo "Image $i : ${IDS[$i]}"
done

# Injecter en grille 2 colonnes × 3 images par colonne
./notion blocks patch-block-children $PAGE_ID --body '{
  "children": [{
    "object": "block",
    "type": "column_list",
    "column_list": {
      "children": [
        {
          "object": "block",
          "type": "column",
          "column": {"children": [
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[1]}"'"}}},
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[2]}"'"}}},
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[3]}"'"}}}
          ]}
        },
        {
          "object": "block",
          "type": "column",
          "column": {"children": [
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[4]}"'"}}},
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[5]}"'"}}},
            {"object":"block","type":"image","image":{"type":"file_upload","file_upload":{"id":"'"${IDS[6]}"'"}}}
          ]}
        }
      ]
    }
  }]
}'
```

---

## 9. API générique (escape hatch)

Pour tout endpoint non couvert par les commandes dédiées.

### Syntaxe
```bash
./notion api <METHOD> <path> [--body '<json>']
```

### Exemples

```bash
# GET — lire le bot courant
./notion api GET /v1/users/me

# POST — recherche
./notion api POST /v1/search --body '{"query":"test"}'

# POST via stdin
echo '{"query":"CLI"}' | ./notion api POST /v1/search

# PATCH — mettre à la corbeille
./notion api PATCH /v1/pages/$PAGE_ID --body '{"in_trash": false}'

# DELETE — supprimer un bloc
./notion api DELETE /v1/blocks/<block_id>
```

---

## Flags globaux utiles

| Flag | Description | Exemple |
|---|---|---|
| `--dry-run` | Affiche la requête HTTP sans l'exécuter | `./notion blocks delete-a-block <id> --dry-run` |
| `--field <key>` | Extrait un seul champ top-level du JSON | `./notion users get-self --field name` |
| `--body @file.json` | Lit le body depuis un fichier | `./notion search post-search --body @query.json` |
| `--body -` | Lit le body depuis stdin | `echo '{}' \| ./notion search post-search --body -` |
| `--help-body` | Affiche un exemple JSON du body attendu | `./notion pages patch-page --help-body` |
| `--debug` | Affiche les détails HTTP (headers, status) | `./notion users get-self --debug` |
| `-f json` | Force le format de sortie JSON | `./notion users get-self -f json` |
