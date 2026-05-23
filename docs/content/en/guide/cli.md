---
title: "CLI Reference"
description: "Command reference for the gohan CLI."
slug: "cli"
categories:
  - guide
translation_key: "cli"
---

## Commands

| Command | Description |
|---|---|
| `gohan build` | Build the site (incremental by default) |
| `gohan build --full` | Force a full rebuild |
| `gohan build --dry-run` | Simulate a build without writing files |
| `gohan build --draft` | Include draft articles (`draft: true`) in the build |
| `gohan build --future` | Include articles whose `date` is in the future |
| `gohan build --stats` | Print per-phase timing and counts after the build |
| `gohan build --explain` | Print which files triggered a rebuild and why |
| `gohan init [--force] [<dir>]` | Scaffold a new gohan project (config + content + archetypes) |
| `gohan new [--type=post] [--title=<t>] <slug>` | Create a new post skeleton |
| `gohan new --type=page [--title=<t>] <slug>` | Create a new page skeleton |
| `gohan serve` | Start the live-reload development server |
| `gohan version` | Print version information |

---

## `gohan build`

Builds the site from `content/` into the configured output directory.

By default, only files that have changed since the last build are regenerated (incremental build). Use `--full` to rebuild everything.

**Flags**

| Flag | Description |
|---|---|
| `--full` | Skip diff detection and regenerate all pages |
| `--dry-run` | Print what would be generated without writing any files |
| `--draft` | Include articles with `draft: true` in their Front Matter. By default drafts are excluded. |
| `--future` | Include articles whose `date` is later than the current time. By default future-dated articles are excluded, allowing them to be "scheduled" by setting a future `date`. |
| `--stats` | Print a per-phase timing report (parse / diff / process / plugins / render / feeds / manifest) and total wall-clock time. |
| `--explain` | Print which content files triggered the rebuild. For a full rebuild it also prints the reason (e.g. `--full` flag, config hash change, missing manifest). |

---

---

## `gohan init`

Scaffolds a new gohan project under the given directory (or the current directory if omitted).

**Usage**

```sh
gohan init [--force] [<dir>]
```

Generates a minimal `config.yaml`, the standard `content/` and `archetypes/` folders, and a starter `README.md`. The command refuses to write into a non-empty directory unless `--force` is given. Existing files are never overwritten — `--force` only allows writing alongside them.

**Flags**

| Flag | Description |
|---|---|
| `--force` | Allow scaffolding into a non-empty directory. Existing files are preserved. |

---

## `gohan new`

Creates a new content skeleton with pre-filled Front Matter.

**Flags**

| Flag | Description |
|---|---|
| `--type` | Content type: `post` (default) or `page` |
| `--title` | Article title (defaults to slug converted to title case) |

**Usage**

| Usage | Description |
|---|---|
| `gohan new [--type=post] [--title=<t>] <slug>` | Create `content/posts/<slug>.md` |
| `gohan new --type=page [--title=<t>] <slug>` | Create `content/pages/<slug>.md` |

If `--title` is omitted, gohan derives a title from the slug (e.g. `my-post` → `My Post`).

---

## `gohan serve`

Starts a local HTTP development server with live reload.

- Default address: `http://127.0.0.1:1313`
- Watches `content/`, `themes/`, and `assets/` for changes
- Automatically rebuilds and reloads the browser on file changes

---

## `gohan version`

Prints the version, commit hash, and build date of the installed binary.
