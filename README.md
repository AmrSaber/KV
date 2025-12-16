# KV - Your Personal Command-Line Key-Value Store

[![Latest Release](https://img.shields.io/github/v/release/AmrSaber/kv?logo=github)](https://github.com/AmrSaber/kv/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-00ADD8?logo=go)](https://go.dev/)
[![Built with SQLite](https://img.shields.io/badge/Database-SQLite-003B57?logo=sqlite)](https://www.sqlite.org/)

KV is a lightweight, feature-rich key-value store that lives right in your terminal. Think of it as a personal database for all those bits of information you need to store and retrieve quickly—configuration snippets, API keys, temporary notes, or any data you want at your fingertips.

Unlike traditional databases, KV is designed for simplicity and speed. No servers to manage, no complex setup—just store a value, retrieve it when you need it, and move on with your day.

**Inspired by [Charm's Skate](https://github.com/charmbracelet/skate)**, KV extends the personal key-value store concept with encryption, version control, TTL management, and enhanced security features.

### KV vs Skate

| Feature                      | Skate | KV                |
| ---------------------------- | ----- | ----------------- |
| Basic Key-Value Storage      | ✅    | ✅                |
| Multiple Databases           | ✅    | ✅ (via prefixes) |
| Binary Data                  | ✅    | ✅                |
| **AES-256 Encryption**       | ❌    | ✅                |
| **Value Visibility Control** | ❌    | ✅                |
| **Version History & Revert** | ❌    | ✅                |
| **Auto-Expiration (TTL)**    | ❌    | ✅                |
| **Soft Deletes**             | ❌    | ✅                |
| **Multi-Key Operations**     | ❌    | ✅                |
| **JSON/YAML Output**         | ❌    | ✅                |

## Why KV?

- **Local & Fast**: All data stored locally in SQLite—no network calls, no dependencies
- **Secure**: Built-in AES-256-GCM encryption for sensitive data
- **Smart Expiration**: Set TTLs on keys for automatic cleanup
- **Version Control**: Complete history tracking with the ability to revert changes
- **Developer-Friendly**: JSON/YAML output, shell completion, and intuitive commands

## Table of Contents

- [KV vs Skate](#kv-vs-skate)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Capabilities](#core-capabilities)
- [Command Reference](#command-reference)
  - [Basic Key-Value Operations](#working-with-basic-key-value-operations)
  - [Managing Encrypted Values](#managing-encrypted-values)
  - [Managing Value Visibility (Hide/Show)](#managing-value-visibility-hideshow)
  - [Time-to-Live (TTL) Management](#time-to-live-ttl-management)
  - [Version Control & History](#version-control--history)
  - [Output Formats](#output-formats)
  - [Batch Operations & Multiple Keys](#batch-operations--multiple-keys)
  - [Backup & Restore](#backup--restore)
  - [Utility Commands](#utility-commands)
- [Configuration](#configuration)
- [Data Storage](#data-storage)
- [Tips & Tricks](#tips--tricks)
- [Use Cases](#use-cases)
- [Contributing](#contributing)

## Installation

Choose the installation method that works best for your platform:

### MacOS

```bash
brew install AmrSaber/tap/kv
```

### Linux

**Homebrew:**

```bash
brew install AmrSaber/tap/kv
```

**Snap (Ubuntu, Fedora, Arch, openSUSE, etc.):**

```bash
sudo snap install kv-cli
```

> **Note:** The snap package is named `kv-cli` (`kv` is reserved). After installation, the command will be `kv-cli`. To use just `kv`, create an alias:
>
> ```bash
> sudo snap alias kv-cli.kv-cli kv
> ```

**Arch Linux (AUR):**

```bash
yay -S kv-bin
# or
paru -S kv-bin
```

### Windows

**Scoop:**

```bash
scoop bucket add amrsaber https://github.com/AmrSaber/scoop-bucket
scoop install kv
```

### Any Platform (Go)

```bash
go install github.com/AmrSaber/kv@latest
```

### Enable Shell Completion (Optional but Recommended)

> **Note:** Package manager installations (Homebrew, Snap, Scoop, AUR) automatically include shell completion. If you installed via `go install`, use the following manual completion setup.

KV provides intelligent auto-completion for commands, flags, and most importantly, **relevant keys** for each command context.

**Bash:**

```bash
echo 'eval "$(kv completion bash)"' >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**

```bash
echo 'eval "$(kv completion zsh)"' >> ~/.zshrc
source ~/.zshrc
```

**Fish:**

```bash
echo 'kv completion fish | source' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

Once enabled, you can tab-complete key names when using commands like `get`, `delete`, `lock`, and more!

## Quick Start

```bash
# Store some values
kv set api-key "sk-1234567890"
kv set database-url "postgres://localhost/mydb"

# Retrieve a value
kv get api-key
# Output: sk-1234567890

# List all keys - see the beautiful table output
kv list
# ┌──────────────┬────────────────────────────┬─────────────────────┐
# │ KEY          │ VALUE                      │ TIMESTAMP           │
# ├──────────────┼────────────────────────────┼─────────────────────┤
# │ api-key      │ sk-1234567890              │ 2025-10-20 21:29:02 │
# │ database-url │ postgres://localhost/mydb  │ 2025-10-20 21:29:05 │
# └──────────────┴────────────────────────────┴─────────────────────┘

# Store with expiration (auto-deletes after 1 hour)
kv set temp-token "xyz789" --expires-after 1h

# Encrypt sensitive data with a password
kv set github-token "ghp_secret" --password "mypass"
```

## Core Capabilities

### Encryption & Security

Lock individual values with password protection using military-grade AES-256-GCM encryption. Perfect for API keys, credentials, and sensitive configuration. Hide sensitive values from list output (without encryption) for privacy.

### Time-to-Live (TTL)

Set automatic expiration on keys. Great for temporary tokens, session data, or anything that shouldn't stick around forever.

### Complete History Tracking

Every change is versioned. Made a mistake? Revert to any previous value. Need to see what changed? Browse the full history.

### Flexible Output

View data as beautiful terminal tables, machine-readable JSON, or structured YAML—whatever fits your workflow.

### Batch & Multi-Key Operations

Work with multiple keys at once — either by specifying them explicitly or using prefix matching. Operations are transactional: all keys succeed or none do.

---

## Command Reference

KV organizes commands into intuitive groups. Here are the main workflows and capabilities with real output examples.

> **Note:** These examples showcase common usage patterns. Each command has additional options and flags available—use `kv <command> --help` to see all available options.

### Working with Basic Key-Value Operations

```bash
# Store and retrieve values
kv set database-url "postgres://localhost/mydb"
kv get database-url
# Output: postgres://localhost/mydb

# List all keys - displays a beautiful table
kv list
# Output:
# ┌────────────────┬───────────────────────────┬─────────────────────┐
# │ KEY            │ VALUE                     │ TIMESTAMP           │
# ├────────────────┼───────────────────────────┼─────────────────────┤
# │ api-key        │ sk-1234567890abcdef       │ 2025-10-20 21:29:02 │
# │ config.db.host │ localhost                 │ 2025-10-20 21:29:02 │
# │ config.db.port │ 5432                      │ 2025-10-20 21:29:02 │
# │ database-url   │ postgres://localhost/mydb │ 2025-10-20 21:28:55 │
# └────────────────┴───────────────────────────┴─────────────────────┘

# List keys with a specific prefix
kv list config
# Output:
# ┌────────────────┬───────────┬─────────────────────┐
# │ KEY            │ VALUE     │ TIMESTAMP           │
# ├────────────────┼───────────┼─────────────────────┤
# │ config.db.host │ localhost │ 2025-10-20 21:29:02 │
# │ config.db.port │ 5432      │ 2025-10-20 21:29:02 │
# └────────────────┴───────────┴─────────────────────┘

# Delete a key (soft delete - keeps history)
kv delete old-setting

# Permanently remove including history
kv delete cached-data --prune
```

### Managing Encrypted Values

> **Security Note:** KV uses AES-256-GCM encryption with PBKDF2 key derivation (10,000 iterations). Passwords are never stored—they're only used to encrypt/decrypt your data. If you lose a password, the encrypted data cannot be recovered. Keep your passwords safe!

```bash
# Store an encrypted value directly
kv set github-token "ghp_secret123" --password "secure123"

# Retrieve encrypted value
kv get github-token --password "secure123"
# Output: ghp_secret123

# Lock an existing plain-text value
kv lock api-key --password "mypass"

# List shows locked values as [Locked]
kv list
# Output:
# ┌──────────────┬──────────┬─────────────────────┐
# │ KEY          │ VALUE    │ TIMESTAMP           │
# ├──────────────┼──────────┼─────────────────────┤
# │ api-key      │ [Locked] │ 2025-10-20 21:30:02 │
# │ github-token │ [Locked] │ 2025-10-20 21:29:29 │
# └──────────────┴──────────┴─────────────────────┘

# Unlock a locked value back to plain text
kv unlock api-key --password "mypass"

# Lock multiple keys at once
kv lock config --prefix --password "mypass"
```

**Warning:** Passing passwords directly on the command line (e.g., `--password "mypass"`) will save them in your shell history, making them visible to anyone with access to your terminal history file. For better security, use environment variables or command substitution:

```bash
# Using an environment variable
kv set api-key "secret" --password "$KV_PASSWORD"

# Using command substitution (e.g., from a password manager)
kv get api-key --password "$(pass show kv/master)"

# Or read from a file
kv lock sensitive-data --password "$(cat ~/.kv-password)"
```

### Managing Value Visibility (Hide/Show)

> **Privacy Note:** Hiding values is not encryption—it only controls visibility in output. Hidden values show as `[Hidden]` in lists but remain accessible via `get`. For true security, use encryption with `lock` instead.

```bash
# Hide sensitive values from list output
kv set api-key "sk-1234567890"
kv hide api-key

# List shows hidden values as [Hidden]
kv list
# Output:
# ┌─────────┬──────────┬─────────────────────┐
# │ KEY     │ VALUE    │ TIMESTAMP           │
# ├─────────┼──────────┼─────────────────────┤
# │ api-key │ [Hidden] │ 2025-10-20 21:29:02 │
# └─────────┴──────────┴─────────────────────┘

# Hidden values are still accessible via get
kv get api-key
# Output: sk-1234567890

# Show a hidden value again
kv show api-key

# Hide multiple keys at once
kv hide api-key db-password secret-token

# Hide all keys with a prefix
kv hide secrets --prefix

# Note: Locked keys always show as [Locked] (takes precedence over hidden state)
```

### Time-to-Live (TTL) Management

```bash
# Set a value with automatic expiration
kv set session-token "abc123xyz" --expires-after 1h

# List automatically shows "EXPIRES AT" column when any key has expiration
kv list
# Output:
# ┌───────────────┬───────────┬─────────────────────┬─────────────────────┐
# │ KEY           │ VALUE     │ TIMESTAMP           │ EXPIRES AT          │
# ├───────────────┼───────────┼─────────────────────┼─────────────────────┤
# │ database-url  │ postgres… │ 2025-10-20 21:28:55 │ -                   │
# │ session-token │ abc123xyz │ 2025-10-20 21:29:25 │ 2025-10-20 22:29:25 │
# └───────────────┴───────────┴─────────────────────┴─────────────────────┘

# Check how long until expiration
kv ttl session-token
# Output: 59m56s (expires at 2025-10-20 22:29:25)

# Get expiration as timestamp only
kv ttl session-token --date
# Output: 2025-10-20 22:29:25

# Set expiration on an existing key
kv expire temp-data --after 30m

# Remove expiration from a key
kv expire session-token --never
```

### Version Control & History

```bash
# Update a key to create history
kv set api-key "sk-1234567890abcdef"
kv set api-key "sk-updated-version"

# View complete history for a key
kv history list api-key
# Output:
# ┌───────┬─────────────────────┬─────────────────────┐
# │ INDEX │ VALUE               │ TIMESTAMP           │
# ├───────┼─────────────────────┼─────────────────────┤
# │ 1     │ sk-1234567890abcdef │ 2025-10-20 21:29:02 │
# │ -     │ sk-updated-version  │ 2025-10-20 21:29:37 │
# └───────┴─────────────────────┴─────────────────────┘
# Note: Index "-" indicates the current/latest value

# Revert to previous value (1 step back by default)
kv history revert api-key
# Output: sk-1234567890abcdef

# Revert multiple steps back
kv history revert api-endpoint --steps 3

# Interactively select from history
kv history select my-config

# Clear history for a key (keeps current value)
kv history prune old-key

# Clear history for all keys with a prefix
kv history prune temp --prefix
```

### Output Formats

```bash
# Get machine-readable JSON output
kv list --output json
# Output:
# [
#   {
#     "key": "api-key",
#     "value": "sk-1234567890abcdef",
#     "timestamp": "2025-10-20T20:29:02Z"
#   },
#   {
#     "key": "github-token",
#     "isLocked": true,
#     "timestamp": "2025-10-20T20:29:29Z"
#   }
# ]

# List only keys, hide values (adds "LOCKED" column for encrypted keys)
kv list --no-values
# Output:
# ┌──────────────┬─────────────────────┬────────┐
# │ KEY          │ TIMESTAMP           │ LOCKED │
# ├──────────────┼─────────────────────┼────────┤
# │ api-key      │ 2025-10-20 21:29:02 │ -      │
# │ github-token │ 2025-10-20 21:29:29 │ Yes    │
# └──────────────┴─────────────────────┴────────┘

# YAML output is also available
kv list --output yaml
```

### Batch Operations & Multiple Keys

```bash
# Operate on multiple keys at once
kv delete old-key temp-data cache-value
kv hide api-key db-password auth-token
kv lock secret1 secret2 secret3 --password "mypass"
kv expire session1 session2 session3 --after 1h

# Or use prefix matching for batch operations
kv delete cache --prefix
kv hide secrets --prefix
kv lock config --prefix --password "mypass"
kv unlock secrets --prefix --password "mypass"

# Unlock all keys at once
kv unlock --all --password "mypass"

# Note: Multi-key operations are transactional — if any key fails,
# none of the changes are applied (all-or-nothing behavior)
```

### Backup & Restore

> **Note:** Export creates a complete snapshot of your database including all keys, values, encryption, hidden state, TTL settings, and full history. Import completely replaces your current database with the imported one, creating a backup of your current database first.

```bash
# Export database to a file
kv db export backup.db
# Output: Database exported to: backup.db

# Export to absolute path
kv db export /path/to/backups/kv-backup-2025-10-20.db

# Overwrite existing backup file
kv db export backup.db --force

# Export to stdout (useful for piping)
kv db export - > backup.db
# Or simply:
kv db export > backup.db

# Import database from a file (replaces current database)
kv db import backup.db
# Output:
# Current database backed up to: /path/to/kv.db.backup
# Database imported successfully

# Import from stdin
cat backup.db | kv db import -
# Or simply:
kv db import < backup.db

# Restore from automatic backup (created during import)
kv db restore
# Output: Database restored from: /path/to/kv.db.backup
# Note: This restores the backup created during the last import operation

# Practical examples:

# Create daily backups
kv db export ~/backups/kv-$(date +%Y-%m-%d).db

# Transfer database between machines
# On source machine:
kv db export - | ssh user@remote 'kv db import -'

# Compress backup
kv db export - | gzip > kv-backup.db.gz
# Restore from compressed backup
gunzip -c kv-backup.db.gz | kv db import -
```

**What gets preserved in export/import:**
- All keys and values (plain text, encrypted, and hidden)
- Password-encrypted keys (with their encryption intact)
- Hidden/visible state
- TTL and expiration settings
- Complete version history for all keys
- All configuration and metadata

**Safety features:**
- Import automatically creates a backup at `<db-path>.backup` before replacing
- You can restore from this backup anytime using `kv db restore`
- Import validates the file is a valid database before proceeding
- Export fails if file exists (use `--force` to override)
- Export validates destination directory exists
- Restore validates the backup is a valid database before restoring
- Restore preserves the backup file after restoration

### Utility Commands

```bash
# Clear all data (keeps configuration)
kv implode
# Warning: This permanently deletes all keys and history

# Generate shell completion (see Installation section for setup)
kv completion bash > /etc/bash_completion.d/kv
```

---

**For detailed information about any command, including all available options and flags:**

```bash
kv --help                          # View all commands
kv <command> --help                # View specific command help
kv history <subcommand> --help     # View history subcommand help
```

All commands have comprehensive help text built into the CLI.

---

## Configuration

KV stores its configuration in a YAML file at your system's standard config location:

- **Linux**: `~/.config/kv/config.yaml`
- **macOS**: `~/Library/Application Support/kv/config.yaml`
- **Windows**: `%APPDATA%\kv\config.yaml`

### Available Settings

```yaml
# How long to keep deleted keys in history (days)
prune-history-after-days: 30

# Maximum history entries to maintain per key
history-length: 15
```

Both settings have sensible defaults.

---

## Data Storage

Your key-value data is stored locally in a SQLite database at:

- **Linux**: `~/.local/share/kv/kv.db`
- **macOS**: `~/Library/Application Support/kv/kv.db`
- **Windows**: `%LOCALAPPDATA%\kv\kv.db`

The database uses WAL (Write-Ahead Logging) mode for better performance and reliability. All data remains completely local—no network calls, no cloud sync, no telemetry.

---

## Tips & Tricks

### Namespace Your Keys

Use dots or slashes to organize related keys:

```bash
kv set app.db.host "localhost"
kv set app.db.port "5432"
kv set app.db.name "myapp"

# List all database config
kv list app.db
```

### Work with Complex Data

Store JSON, multi-line text, or any structured data:

```bash
# Store JSON configuration
kv set app.config '{
  "database": "postgres://localhost/db",
  "port": 8080,
  "debug": true
}'

# Retrieve and pipe to jq
kv get app.config | jq '.database'
# Output: "postgres://localhost/db"
```

### Combine with Shell Scripts

Use KV in your automation scripts:

```bash
#!/bin/bash
# Store build timestamp
kv set last-build "$(date)" --expires-after 24h

# Retrieve API key for deployment
API_KEY=$(kv get deploy-key --password "$MASTER_PASS")
curl -H "Authorization: Bearer $API_KEY" https://api.example.com/deploy
```

### Quick Temporary Storage

Perfect for sharing data between terminal sessions:

```bash
# Terminal 1
kv set clipboard "some long command or text"

# Terminal 2 (even different window/tab)
kv get clipboard
```

### Audit Changes with History

Track configuration changes over time:

```bash
# See all changes to production config
kv history list prod.api.endpoint

# Find out when something changed
kv history list db.password | grep "2025-10-15"
```

### Batch Cleanup

Clean up temporary data efficiently:

```bash
# Set expiration on all temp keys
kv list temp --output json | jq -r '.[].key' | while read key; do
  kv expire "$key" --after 1h
done

# Or just delete them all
kv delete temp --prefix
```

---

## Use Cases

**Development**

- Store API keys, database URLs, and service endpoints locally
- Manage environment-specific configurations without `.env` files
- Quick access to frequently-used test data or tokens

**DevOps & System Administration**

- Temporarily store credentials during deployment or maintenance
- Share configuration snippets between terminal sessions
- Track configuration changes with built-in version control

**Scripting & Automation**

- Inter-script communication and data passing
- Store script state that persists between runs
- Cache expensive computation results with automatic expiration

**Personal Productivity**

- Keep track of license keys and access codes
- Store frequently-used snippets and commands
- Maintain a personal knowledge base of settings and configurations

**Security & Secrets Management**

- Encrypted storage for sensitive data with password protection
- Time-limited access tokens that auto-expire
- Local-only storage - no network exposure

---

## Contributing

Contributions are welcome! If you'd like to contribute:

1. **Fork** the repository at [github.com/AmrSaber/kv](https://github.com/AmrSaber/kv)
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Bug Reports & Feature Requests

Please report bugs and request features at:
**https://github.com/AmrSaber/kv/issues**

When reporting bugs, please include:

- Your OS and Go version
- Steps to reproduce the issue
- Expected vs actual behavior
