# KV - Your Personal Command-Line Key-Value Store

[![Latest Release](https://img.shields.io/github/v/release/AmrSaber/kv?logo=github)](https://github.com/AmrSaber/kv/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://go.dev/)
[![Built with SQLite](https://img.shields.io/badge/Database-SQLite-003B57?logo=sqlite)](https://www.sqlite.org/)

KV is a lightweight, feature-rich key-value store that lives right in your terminal. Think of it as a personal database for all those bits of information you need to store and retrieve quickly — configuration snippets, API keys, temporary notes, or any data you want at your fingertips. No servers to manage, no complex setup.

## Features

- **Local SQLite storage** — no server, no setup
- **AES-256-GCM encryption** — per-key password protection
- **TTL / auto-expiry** — keys expire on schedule
- **Full history & revert** — every change is versioned
- **Hide/show visibility** — control what appears in listings
- **Multiple databases** — separate key namespaces in their own files
- **`key@db` syntax** — target any database per key
- **Batch & prefix operations** — work with many keys at once
- **JSON / YAML / table output** — machine and human readable
- **Cross-database copy & move** — promote keys between databases
- **Transactional guarantees** — all-or-nothing within a single DB
- **Shell completion** — substring matching across all databases

## Benchmark
`kv` is super fast. Here are the results from [hyperfine](https://github.com/sharkdp/hyperfine) benchmark for `get` and `set` commands:

```
Benchmark 1: ./kv set bench-key "hello"
  Time (mean ± σ):       6.6 ms ±   0.6 ms    [User: 3.6 ms, System: 3.6 ms]
  Range (min … max):     5.3 ms …   8.4 ms    408 runs

Benchmark 2: ./kv get bench-key
  Time (mean ± σ):       6.6 ms ±   0.6 ms    [User: 3.4 ms, System: 3.8 ms]
  Range (min … max):     5.3 ms …   8.1 ms    400 runs

Summary
  ./kv get bench-key ran
    1.00 ± 0.13 times faster than ./kv set bench-key "hello"
```

You can run the benchmark yourself using `mise run benchmark`.

## Quick Start

```bash
kv set api-key "sk-1234567890"
kv set config.json '{"debug": true}' --expires-after 24h
kv get api-key
kv list
kv lock db-password --password        # encrypt with interactive prompt
```

See `kv --help` for all commands.

## Installation

| Platform                 | Command                                             |
| ------------------------ | --------------------------------------------------- |
| macOS / Linux (Homebrew) | `brew install AmrSaber/tap/kv`                      |
| Linux (Snap)             | `sudo snap install kv-cli`                          |
| Arch Linux (AUR)         | `yay -S kv-bin`                                     |
| Windows (Scoop)          | `scoop bucket add amrsaber ... && scoop install kv` |
| Any (Go)                 | `go install github.com/AmrSaber/kv@latest`          |

## Key Concepts

### Multiple Databases & Key Parsing

KV supports separate databases, each backed by its own SQLite file. By default everything goes into the `default` database. Databases are created automatically on first use.

Three ways to target a database:

```bash
KV_DB=work kv set api-key "sk-1234"    # KV_DB env variable
kv --db work set api-key "sk-1234"     # --db flag for one command (overrides env)
kv set config@work "value"             # key@db syntax (overrides --db and env)
kv set api-key "sk-1234"               # Uses 'default' DB
```

**The `@` shorthand** copies/moves a key under the same name to another database:

```bash
kv copy config@work @personal     # copies 'config' from work to personal
kv move config@work @personal     # moves 'config' from work to personal
```

**Keys with literal `@`**: Set `KV_NO_PARSE_KEYS=1` to bypass key parsing entirely. All keys are treated literally and `key@db` syntax is disabled. Only `--db` and `KV_DB` remain available for targeting databases.

```bash
export KV_NO_PARSE_KEYS=1
kv set "user@example.com" "value"
```

> Generally speaking, the use of keys that include '@' is discouraged, and `KV_NO_PARSE_KEYS` env variable is there as a work around to be used when absolutely necessary.

**DB management commands:**

| Command                             | Description                                         |
| ----------------------------------- | --------------------------------------------------- |
| `kv db ls`                          | List all registered databases and their directories |
| `kv db rm <name>`                   | Delete a database (creates backup unless `--prune`) |
| `kv db set name <old> <new>`        | Rename a database                                   |
| `kv db set directory <name> <path>` | Move a database to a custom directory               |

> The `default` database cannot be updated or deleted. All DB management commands create a backup before making changes.

**Cross-DB copy & move:**

- **Copy** preserves encryption and hidden state. TTL is **not** preserved across databases.
- **Move** preserves everything: value, encryption, hidden state, TTL, and full history.
- Cross-database operations are **not** transactional — each database is handled independently.

### Encryption & the `--password` Flag

KV uses AES-256-GCM encryption with PBKDF2 key derivation (10,000 iterations). Passwords are never stored — if you lose a password, the data cannot be recovered.

```bash
kv set github-token "ghp_secret" --password        # prompts interactively (recommended)
kv set github-token "ghp_secret" --password=mypass  # inline for scripting
kv get github-token --password                      # decrypt on retrieval
kv lock existing-key --password                     # encrypt a plain-text key
kv unlock existing-key --password                   # decrypt back to plain text
```

> **Password flag quirk:** Use `--password=value` (with `=`). `--password value` (with a space) treats the value as a positional argument due to Cobra's `NoOptDefVal` mechanism.

For write operations (`set`, `lock`), the interactive prompt asks twice to confirm. Locked keys show as `[Locked]` in lists.

### Storage & Transaction Model

- Each database is a standalone SQLite file.
- All single-DB operations run inside a transaction: **all keys succeed or none do** (atomicity applies to batch and prefix operations).
- Cross-DB operations (`copy`/`move` across databases) are **not** transactional — each DB is handled independently.
- History is implemented as rows with an `is_latest` flag. Soft deletes set the value to an empty string. Every change appends a new row; old rows are trimmed based on `history-length` configuration.

## Example Workflows

```bash
# Environment-specific configs
kv --db staging set api-key "sk-test"
kv --db prod set api-key "sk-live"
kv copy api-key@staging @prod          # promote config across environments

# Temp data with auto-cleanup
kv set deploy-token "tok_abc" --expires-after 2h

# Script integration
API_KEY=$(kv get api-key --password="$MASTER_PASS")
curl -H "Authorization: Bearer $API_KEY" https://api.example.com/deploy

# Cross-device transfer
kv db backup --stdout | ssh server 'kv db restore --stdin'

# Audit changes
kv set db-url "postgres://localhost/mydb"
kv set db-url "postgres://prod/mydb"
kv history list db-url                 # see every version
kv history revert db-url               # undo last change

# Machine-readable output
kv list --output json | jq '.[] | select(.isLocked) | .key'
```

## Shell Completion & Scripting

```bash
# Enable
echo 'eval "$(kv completion)"' >> ~/.bashrc   # or ~/.zshrc
source ~/.bashrc
```

Completions use substring matching and work across all -registered- databases — type any part of a key name to narrow suggestions. Keys from non-current databases are annotated with `@db`.

## Configuration

KV stores its configuration in a YAML file (see path via `kv info`). Config and data directories follow the XDG Base Directory Specification.

```yaml
prune-history-after-days: 30                    # how long to keep soft-deleted keys - default: 30
history-length: 15                              # max history entries per key - default: 15
dbs:                                            # custom DBs - default: empty (configs for 'default' db cannot be updated)
  work:                                         # DB name
    directory: /home/user/.local/share/kv       # DB directory
  personal:
    directory: /home/user/personal-kv
```

The `dbs` section is managed automatically — databases are registered when first accessed. Use `kv db set directory` to change a database's storage location. This handles creating a backup for the DB and moving it. For that reason, manually updating that section is highly discouraged.

## Data Storage

Each database is a standalone SQLite file. By default all files live under `$XDG_DATA_HOME/kv` (see `kv info`). Custom directories can be set per database. All data is local — no network calls, no telemetry.

## Contributing

For bug reports and feature requests, please [create a ticket](https://github.com/AmrSaber/kv/issues).

The project uses integration tests (and a few unit tests) that build the binary and run it against real SQLite databases, covering nearly all commands and workflows. Run them with `mise run test`.
