## Known Bugs
- [ ] DB backup/restore default path does not use DB selected by `--db` flag
  - The default value for the flag is defined in `init()`, while `--db` is processed at root's `PersistentPreRun`
  - Using KV_DB works though
     
## Feature Ideas
- [ ] List from all DBs (using `--all` flag for `kv ls`)
