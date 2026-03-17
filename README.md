<h1 align="center">
  notion-cli
</h1>

<p align="center">
  <b>Like <code>gh</code> for GitHub, but for Notion. For humans and AI agents.</b>
</p>

<p align="center">
  <a href="https://github.com/JeremieAlcaraz/notion-cli/releases"><img src="https://img.shields.io/github/v/release/JeremieAlcaraz/notion-cli?style=flat-square" alt="Release"></a>
  <a href="https://github.com/JeremieAlcaraz/notion-cli/actions"><img src="https://img.shields.io/github/actions/workflow/status/JeremieAlcaraz/notion-cli/test.yml?style=flat-square&label=tests" alt="Tests"></a>
  <a href="https://github.com/JeremieAlcaraz/notion-cli/blob/main/LICENSE"><img src="https://img.shields.io/github/license/JeremieAlcaraz/notion-cli?style=flat-square" alt="License"></a>
</p>

---

## Origin & Credits

This project started from [4ier/notion-cli](https://github.com/4ier/notion-cli) — an excellent foundation that covers the entire Notion API in a single binary. Rather than forking, I chose to create an independent repository to pursue a different direction.

All credit for the original design and implementation goes to [@4ier](https://github.com/4ier). If you're looking for a stable, production-ready Notion CLI, check out the original.

The new direction taken in this repo — OpenAPI-driven generation, agent mode, human mode — was developed with [Claude Code](https://claude.ai/claude-code) as the primary coding assistant. Most of the code in this fork was generated through AI-assisted development.

---

## What's different here

This repo takes the original as a starting point and extends it in three directions:

**1. OpenAPI-driven generation**
Commands are generated from the official `spec/notion-openapi.json` spec. Adding a new Notion endpoint means updating the spec, not writing Go by hand.

**2. Agent mode** — optimized for LLM token consumption
```sh
notion --agent pages retrieve-a-page <id>
```
Minified JSON output, no colors, no decorative headers. Designed to reduce token usage by 60–90% when used inside AI agents (Claude, Codex, etc.).

**3. Human mode** — interactive UI powered by `gum`
```sh
notion --human search
```
Interactive prompts, selects, confirmations. Built for terminal users who prefer a guided experience over flags.

---

## Install

### Homebrew (macOS/Linux)
```sh
brew install JeremieAlcaraz/tap/notion-cli
```

### Go
```sh
go install github.com/JeremieAlcaraz/notion-cli@latest
```

### Binary
Download from [GitHub Releases](https://github.com/JeremieAlcaraz/notion-cli/releases) — available for Linux, macOS, and Windows (amd64/arm64).

---

## Install AI skill

Once installed, you can register the CLI skill for your AI agent:

```sh
notion skill install           # auto-detects Claude, Codex, etc.
notion skill install --claude  # → ~/.claude/skills/notion-cli/
notion skill install --codex   # → ~/.codex/skills/notion-cli/
```

The skill is embedded in the binary and always matches the installed version.

---

## Quick Start

```sh
# Authenticate
export NOTION_TOKEN=ntn_xxxxx

# Search your workspace
notion search "meeting notes"

# Query a database
notion databases retrieve-database <db-id>

# Get page content as Markdown
notion pages retrieve-page-markdown <page-id>

# Raw API escape hatch
notion api GET /v1/users/me
```

---

## Commands

```
pages
  retrieve-a-page <page_id>
  post-page --body '<json>'
  patch-page <page_id> --body '<json>'
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

skill
  install [--claude] [--codex]
```

---

## Output flags

| Flag | Effect |
|------|--------|
| `--fields id,url,title` | Keep only listed fields |
| `--field results` | Extract single top-level field |
| `--format summary` | Compact: id, type, title, url, parent_id |
| `--format md` | Render Notion blocks as Markdown |
| `--format ndjson` | One JSON line per item in results[] |
| `--strip-meta` | Remove noisy Notion metadata |
| `--all` | Auto-paginate until has_more=false |
| `--agent` | Minified JSON, no colors, terse errors |

---

## License

[MIT](LICENSE)
