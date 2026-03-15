# Test Results & Reliability Recommendations

> End-to-end battery run on 2026-03-15 against the Notion API v2026-03-11.
> Integration: `cli-demo-test` (bot), all capabilities enabled.
> Test page: `Demo Page CLI — ✅ testée` (32445368-aa31-80da-ad1a-f9d92a4737a0)

---

## Battery Results

### Users (3/3 ✅)

| Command | Operation | Result |
|---|---|---|
| `users get-self` | R | ✅ Bot user retrieved |
| `users get-users` | R | ✅ 16 workspace users listed |
| `users get-user <id>` | R | ✅ User retrieved by ID |

### Pages (7/7 ✅)

| Command | Operation | Result |
|---|---|---|
| `pages retrieve-a-page <id>` | R | ✅ Page + properties retrieved |
| `pages retrieve-a-page-property <id> title` | R | ✅ Property value retrieved |
| `pages retrieve-page-markdown <id>` | R | ✅ Markdown content returned |
| `pages patch-page <id> --body '{...}'` | U | ✅ Title updated |
| `pages post-page --body '{...}'` | C | ✅ Sub-page created under parent |
| `pages move-page <id> --body '{...}'` | U | ✅ Page moved to new parent |
| `pages update-page-markdown <id> --body '{...}'` | U | ✅ Content inserted via `insert_content` type |

**Note on `update-page-markdown`:** the spec lists 3 operation types:
- `insert_content` — insert new content (optionally after a selection) ✅ tested
- `replace_content_range` — replace a range of existing content
- `replace_content` — replace all content (deprecated per API errors)

### Blocks (5/5 ✅)

| Command | Operation | Result |
|---|---|---|
| `blocks retrieve-a-block <id>` | R | ✅ Block retrieved with type |
| `blocks get-block-children <id>` | R | ✅ Child blocks listed |
| `blocks patch-block-children <id> --body '{...}'` | C | ✅ heading_2, callout, bullet created |
| `blocks update-a-block <id> --body '{...}'` | U | ✅ Block content updated |
| `blocks delete-a-block <id>` | D | ✅ `in_trash: true` confirmed (v2026-03-11) |

### Comments (3/3 ✅)

| Command | Operation | Result |
|---|---|---|
| `comments create-a-comment --body '{...}'` | C | ✅ Comment created on page |
| `comments list-comments --block-id <id>` | R | ✅ Comments listed (requires `--block-id` flag) |
| `comments retrieve-comment <id>` | R | ✅ Single comment retrieved |

### Search (1/1 ✅)

| Command | Operation | Result |
|---|---|---|
| `search post-search --body '{"query":"..."}'` | R | ✅ Results returned |

**Note:** filter value changed in v2025-09-03: use `"data_source"` instead of `"database"` in
`filter.value`.

### Databases (3/3 ✅)

| Command | Operation | Result |
|---|---|---|
| `databases create-database --body '{...}'` | C | ✅ Database created with properties |
| `databases retrieve-database <id>` | R | ✅ Schema + properties retrieved |
| `databases update-database <id> --body '{...}'` | U | ✅ Title updated |

### Data Sources (5/5 ✅)

| Command | Operation | Result | Note |
|---|---|---|---|
| `data-sources retrieve-a-data-source <id>` | R | ✅ | Use `data_source_id`, not `database_id` |
| `data-sources update-a-data-source <id> --body '{...}'` | U | ✅ | |
| `data-sources post-database-query <id> --body '{...}'` | R | ✅ | |
| `data-sources list-data-source-templates <id>` | R | ✅ | |
| `data-sources create-a-database --body '{...}'` | C | ✅ | Parent must be `database_id`, not `page_id` |

**Key distinction (new in API v2025-09-03):**
- A **Database** is the container (has an icon, title, lives under a page)
- A **Data Source** is the table of data that lives *inside* a database
- `data_source_id ≠ database_id` — use `POST /v1/search` with `filter.value: "data_source"` to list them

### File Uploads (4/4 ✅ + 1 ℹ️)

| Command | Operation | Result | Note |
|---|---|---|---|
| `file-uploads list-file-uploads` | R | ✅ | |
| `file-uploads create-file --body '{...}'` | C | ✅ Upload slot created (status: pending) |
| `file-uploads retrieve-file-upload <id>` | R | ✅ Status + metadata retrieved |
| `file-uploads upload-file <path> [--page-id <id>]` | C | ✅ Full flow: create→send→attach |
| `file-uploads complete-file-upload` | — | ℹ️ Only needed for files >20MB (multi-part) |

**Upload flow for files <20MB:**
```
upload-file <path>           # handles create + send internally
                             # status becomes "uploaded" automatically
```

**Upload flow for files >20MB:**
```
create-file --body '{...}'   # get upload_id
upload-file <path>           # send part 1 (with --part-number 1)
upload-file <path>           # send part N
complete-file-upload <id>    # finalize
```

**Bonus — tested live:** 3-column × 2-row image grid injected via `patch-block-children`
using `column_list` → `column` → `image` (file_upload) blocks. All 6 instances rendered
correctly in Notion.

---

## Summary

| Group | Commands | Passing |
|---|---|---|
| Users | 3 | ✅ 3/3 |
| Pages | 7 | ✅ 7/7 |
| Blocks | 5 | ✅ 5/5 |
| Comments | 3 | ✅ 3/3 |
| Search | 1 | ✅ 1/1 |
| Databases | 3 | ✅ 3/3 |
| Data sources | 5 | ✅ 5/5 |
| File uploads | 4+1 | ✅ 4/4 + 1 ℹ️ |
| **Total** | **31** | **✅ 31/31** |

---

## Reliability Issues Observed & Recommendations

These are friction points hit during the battery — ranked by impact.

---

### 🔴 R1 — IDs ambiguity: `database_id` vs `data_source_id`

**Observed:** `data-sources` commands fail silently with `object_not_found` when
passed a `database_id` instead of a `data_source_id`. The error message doesn't
explain the distinction.

**Root cause:** Since API v2025-09-03, Notion separates Databases (containers) from
Data Sources (tables within a database). They have different IDs.

**Recommendation:**
- Improve the error hint in `errorHint()` for `object_not_found` on `/v1/data_sources/`:
  *"For data-sources commands, use a data_source_id (not a database_id). Run: `notion search post-search --body '{\"filter\":{\"value\":\"data_source\",\"property\":\"object\"}}'` to list them."*
- Consider a `data-sources list` shortcut command.

---

### 🔴 R2 — `--body` JSON parsing: multiline strings are fragile in shell

**Observed:** Several commands failed with `jq parse error` because the `--body` flag
value contained shell-interpolated variables, newlines, or unescaped quotes.
Required multiple retries to get escaping right.

**Root cause:** Passing structured JSON via a CLI flag is inherently fragile.
Users must manually escape quotes and handle variable interpolation.

**Recommendation:**
- Support `--body @file.json` (read body from file, like `curl`)
- Support `--body -` (read body from stdin)
- Example: `echo '{"query":"x"}' | notion search post-search --body -`

---

### 🟠 R3 — `upload-file` command takes a `file_upload_id`, not a file path

**Observed:** The generated `upload-file` command expected an upload ID (from the spec),
but users expect to pass a file path. Required a manual override.

**Root cause:** The spec models the low-level API (`/send` endpoint takes an ID),
not the user workflow (pick a file, upload it end-to-end).

**Recommendation:**
The manual `upload-file <path>` implementation is the right UX. The generator should
have a mechanism to mark specific operationIds as "manually implemented" so they are
excluded from generation and never overwritten by `just generate`.

---

### 🟠 R4 — `jq` output parsing: generated `id` field conflicts with nested IDs

**Observed:** When piping `upload-file` output to `jq '.id'`, the bot's `created_by.id`
was returned instead of the file upload's top-level `id`, because both appear in the
JSON at different nesting levels.

**Root cause:** `render.Output()` dumps raw JSON — no normalization or top-level
field promotion.

**Recommendation:**
- `render.Output()` could print a `# id: xxx` header line in TTY mode for easy grabbing
- Or expose a `--field id` flag to extract a specific top-level field:
  `notion file-uploads create-file --body '{...}' --field id`

---

### 🟡 R5 — `complete-file-upload` is in the generated commands but misleading

**Observed:** `complete-file-upload` accepts a `--body` flag but the API rejects any body.
It also only applies to files >20MB, which is not obvious.

**Recommendation:**
- Remove from the generated group or add a clear `Short` description:
  *"Finalize a multi-part upload (files >20MB only). For single files, use upload-file."*
- The `upload-file` command already handles the complete flow for normal files.

---

### 🟡 R6 — No `--dry-run` flag for destructive operations

**Observed:** `delete-a-block` and `patch-page` execute immediately with no confirmation.
In scripts, a typo in the ID can delete the wrong block.

**Recommendation:**
- Add `--dry-run` global flag that prints the HTTP request without executing it
- Add `--yes` flag (already in DESIGN.md) to skip confirmation on destructive ops

---

### 🟢 R7 — `update-page-markdown` body format not discoverable

**Observed:** The error messages from Notion were helpful in finding the right format
(`insert_content`, `replace_content_range`), but required multiple trial-and-error attempts.

**Recommendation:**
- Add `--help-body` flag that prints an example body for the current command, derived
  from the OpenAPI spec's `requestBody.content.application/json.schema` examples.
- This would make every generated command self-documenting for its body format.
