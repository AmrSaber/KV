# Design for Separate DB

## Config
```yaml
# All old fields are the same, plus...
dbs:
  default:
    directory: <data-dir>/
  other-db:
    directory: /some/directory/
```

## Interface
```bash
# Global flag
kv --db other-db <command>

# Env variable override
KV_DB=other-db kv <command>
KV_DB=other-db kv set some-key some-value

# Access keys from other DBs
kv <command> <key>@<db>
kv set some-key@other-db some-value
kv hide some-key@other-db

# Set new directory for DB -> writes into config file -- create directory if it doesn't exist
# Backup DB and move it to the new directory
kv db set directory <db> <path>
kv db set directory some-db /some/path/for/db/

# Rename DB
# Old DB must exist in configs
# Backup old DB then rename file, leave backup as-is
kv db set name <old> <new>
kv db set name some-db other-db

# Delete DB
# Cannot delete default DB
# Backup old DB then rename file, leave backup as-is
# Remove DB from data directory and from config
kv db rm <db>
kv db rm some-db

# Moving keys between DBs
# Through rename
kv rename some-key some-key@other-db
kv rename some-key@other-db some-key
# Through copy
kv copy some-key some-key@other-db
kv copy some-key@other-db some-key
```

## Notes
- Prohibit keys with '@' in the name
  - Enforce in commands `set`, `copy`, `rename`
  - Migrate existing keys with `@` -> replace `@` with `:`
- Default DB name becomes `default` with migration if `kv` db exists:
  - Back up `kv` db, then rename db file to default, leaving the backup as-is
- Prohibit the name `kv` for DB, enforce in:
  - KV_DB env variable (in configs)
  - `--db` flag (in root command)
  - `@` syntax
  - In `db rename` command
- DB precedence order: `<key>@db`, `--db` flag, env variable, config's db
- DB name and path in DB commands' docs needs to be updated
