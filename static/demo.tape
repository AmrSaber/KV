Output static/demo.mp4
Output static/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 500
Set Theme "Dracula"
Set Padding 10

# ── Hidden setup ────────────────────────────────────────────────────────────
Hide

# Use a throw-away data dir so real kv data is untouched
Type "export XDG_DATA_HOME=$(mktemp -d)"
Enter
Sleep 200ms

Type `echo "mypass" > /tmp/kv-demo-pass`
Enter
Sleep 200ms

Type "cd ~"
Enter
Sleep 200ms

Ctrl+L

Show
# ── Visible demo ────────────────────────────────────────────────────────────

# --- Basic set / get ---
Type "# Store and retrieve values"
Enter
Sleep 800ms

Type "kv set api-key sk-1234567890abcdef"
Enter
Sleep 1.5s

Type "kv get api-key"
Enter
Sleep 1.5s

Enter

# --- List ---
Type "kv set config.db.host localhost"
Enter
Sleep 500ms

Type "kv set config.db.port 5432"
Enter
Sleep 500ms

Type "kv list"
Enter
Sleep 3s

Enter

# --- Prefix filter ---
Type "# Filter by prefix"
Enter
Sleep 500ms

Type "kv list config"
Enter
Sleep 3s

Ctrl+L

# --- Encryption ---
Type "# Encrypt sensitive values"
Enter
Sleep 800ms

# bare --password triggers interactive prompt (confirms twice for writes)
Type "kv set github-token ghp_secret123 --password"
Enter
Sleep 500ms
Type "mypass"
Enter
Sleep 300ms
Type "mypass"
Enter
Sleep 1.5s

Type "kv list"
Enter
Sleep 3s

Enter

# get from file (scripting-friendly path)
Type `kv get github-token --password="$(cat /tmp/kv-demo-pass)"`
Enter
Sleep 2s

Enter

# get interactively (bare --password, prompts once for reads)
Type "kv get github-token --password"
Enter
Sleep 500ms
Type "mypass"
Enter
Sleep 2s

Ctrl+L

# --- Hide / Show ---
Type "# Hide values from list output"
Enter
Sleep 800ms

Type "kv hide api-key"
Enter
Sleep 1.5s

Type "kv list"
Enter
Sleep 3s

Enter

Type "# Hidden values are still accessible via get"
Enter
Sleep 800ms

Type "kv get api-key"
Enter
Sleep 2s

Ctrl+L

# --- TTL ---
Type "# Set automatic expiration"
Enter
Sleep 800ms

Type "kv set session-token abc123xyz --expires-after 1h"
Enter
Sleep 1.5s

Type "kv list"
Enter
Sleep 3s

Enter

Type "kv ttl session-token"
Enter
Sleep 2s

Ctrl+L

# --- Version history ---
Type "# Every change is versioned"
Enter
Sleep 800ms

Type "kv set config.db.port 5433"
Enter
Sleep 1s

Type "kv history list config.db.port"
Enter
Sleep 3s

Enter

Type "kv history revert config.db.port"
Enter
Sleep 1.5s

Type "kv get config.db.port"
Enter
Sleep 3s

# ── Hidden teardown ─────────────────────────────────────────────────────────
Hide

Type "rm -rf $XDG_DATA_HOME /tmp/kv-demo-pass"
Enter
Sleep 100ms
