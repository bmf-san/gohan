# CLI Reference

> 日本語版: [cli.ja.md](cli.ja.md)

---

## Commands

| Command | Description |
|---|---|
| `gohan build` | Build the site (incremental by default) |
| `gohan build --full` | Force a full rebuild |
| `gohan build --dry-run` | Simulate a build without writing files |
| `gohan new post --slug=<s> --title=<t>` | Create a new post skeleton |
| `gohan new page --slug=<s> --title=<t>` | Create a new page skeleton |
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

---

## `gohan new`

Creates a new content skeleton with pre-filled Front Matter.

**Subcommands**

| Subcommand | Description |
|---|---|
| `gohan new post --slug=<s> [--title=<t>]` | Create `content/posts/<slug>.md` |
| `gohan new page --slug=<s> [--title=<t>]` | Create `content/pages/<slug>.md` |

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
