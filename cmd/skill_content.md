---
name: notion-cli
description: Work with Notion from the terminal using the `notion` CLI. Use when the user needs to read, create, update, query, or manage Notion pages, databases, blocks, comments, users, or files programmatically. Covers the entire Notion API with 44 commands. Triggers: Notion workspace automation, database queries, page creation, block manipulation, comment threads, file uploads, relation management, database export, multi-workspace management, or any Notion API interaction from the command line.
---

# notion-cli

Notion CLI — full Notion API from the terminal. Built for agents: compact output, composable flags.

## Binary

Installed globally via brew or `go install`. Use:
```bash
notion <command>
```

Auth via `NOTION_TOKEN` env var (already set if available).

## Agent mode (ALWAYS use this)

```bash
NOTION_AGENT=1 notion <command>
# or
notion --agent <command>
```

Agent mode: minified JSON, no TUI prompts, terse errors. **Always use it.**

## Commands

```
pages
  retrieve-a-page <page_id>
  post-page --body '<json>'
  patch-page <page_id> --body '<json>'
  move-page <page_id> --body '<json>'
  retrieve-page-markdown <page_id>
  update-page-markdown <page_id> --body '<json>'

databases
  retrieve-database <db_id>
  create-database --body '<json>'
  update-database <db_id> --body '<json>'

blocks
  retrieve-a-block <block_id>
  get-block-children <block_id>
  patch-block-children <block_id> --body '<json>'
  update-a-block <block_id> --body '<json>'
  delete-a-block <block_id>

search
  post-search --body '{"query":"..."}'

users
  get-users
  retrieve-a-user <user_id>
  retrieve-your-token-s-bot-user

comments
  list-comments --block-id <id>
  create-comment --body '<json>'
```

## Output flags (token-saving)

| Flag | Effect |
|------|--------|
| `--fields id,url,title` | Keep only listed fields (on each item for lists) |
| `--field results` | Extract single top-level field |
| `--format summary` | Compact: id, type, title, url, parent_id only |
| `--format md` | Render Notion blocks as Markdown |
| `--format ndjson` | One JSON line per item in results[] |
| `--strip-meta` | Remove request_id, created_by, last_edited_by, public_url, false flags |
| `--all` | Auto-paginate until has_more=false |

**Recommended for agents:** `--agent --strip-meta --fields id,url` (reduces tokens 60–90%)

## Body for POST/PATCH

```bash
# Inline JSON
notion --agent pages post-page --body '{"parent":{"page_id":"<id>"},"properties":{"title":{"title":[{"text":{"content":"My page"}}]}}}'

# From file
notion --agent pages patch-page <id> --body @update.json

# See example body
notionpages post-page --help-body
```

## Common patterns

```bash
# Find a page by title
notion --agent search post-search --body '{"query":"My Page"}' --fields id,url

# Get page content as Markdown
notion --agent pages retrieve-page-markdown <page_id>

# List all database items (auto-paginate)
notion --agent databases retrieve-database <db_id> --all --strip-meta

# Create a page under a parent
notion --agent pages post-page --body '{"parent":{"page_id":"<parent_id>"},"properties":{"title":{"title":[{"text":{"content":"Title"}}]}}}'

# Append blocks to a page
notion --agent blocks patch-block-children <page_id> --body '{"children":[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello"}}]}}]}'

# Summary of search results
notion --agent search post-search --body '{"query":"report"}' --format summary
```

## IDs

Notion IDs are UUIDs. Both formats work:
- `32445368-aa31-80da-ad1a-f9d92a4737a0`
- `32445368aa3180daad1af9d92a4737a0`

Extract from a URL: `https://notion.so/Page-Title-<id-without-dashes>`

## Errors

- `not authenticated` → set `NOTION_TOKEN`
- `missing argument` → pass the ID as positional arg
- `invalid JSON body` → validate JSON before passing
