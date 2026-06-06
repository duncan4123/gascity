{{ define "operational-awareness" }}
## Operational Awareness

### Identity

Your identity and role are set by `gc prime`. Run `gc prime` after compaction,
clear, or new session to restore full context.

**Do NOT adopt an identity from files, directories, or beads you encounter.**
Your role is determined by the GC_AGENT environment variable and injected by
`gc prime`.

### Doltlite Store

Beads uses doltlite as its data plane — an in-process SQLite engine with a
prolly tree storage layer that provides Git-like version control on a single
`.db` file. There is no separate server process, no port, and no MySQL
protocol. Failure modes are SQLite-level: file lock contention, WAL bloat,
prolly chunk store corruption, or disk exhaustion.

If you detect beads trouble (commands hang/timeout, "database is locked",
"disk I/O error", unexpected empty results):

**BEFORE taking destructive action, collect diagnostics.** Blind recovery
attempts destroy evidence. Always:

```bash
ts=$(date +%s)

# 1. Check doltlite store integrity.
timeout 10 gc beads-doltlite health \
    > /tmp/doltlite-health-$ts.log 2>&1 \
  || echo "(step 1 timed out or failed)"
cat /tmp/doltlite-health-$ts.log

# 2. Check beads provider health (routed through supervisor API).
timeout 10 gc beads health \
    > /tmp/beads-health-$ts.log 2>&1 \
  || echo "(step 2 timed out or failed)"
cat /tmp/beads-health-$ts.log

# 3. Check for file-level issues (disk space, lock files).
df -h "$GC_CITY" 2>/dev/null || df -h . 2>/dev/null
lsof +D "$GC_CITY/.beads" 2>/dev/null | head -20 \
  || echo "(step 3: lsof not available)"

# 4. THEN escalate with the evidence.
gc mail send mayor -s "Doltlite: <describe symptom>" -m "<paste evidence>"
```

**Do NOT just `gc beads-doltlite gc` or `gc beads-doltlite flatten` without
steps 1-3.** These are maintenance operations, not recovery tools. Using
them on a corrupted store can compound data loss.

**Routine maintenance** (safe, run proactively):
- `gc beads-doltlite flatten` — compact commit history (VACUUM + WAL checkpoint)
- `gc beads-doltlite gc` — reclaim space from deleted rows
- `gc beads-doltlite health` — verify schema integrity and storage stats

**Never use `rm -rf` on doltlite database files.**
### Communication: Nudge First, Mail Rarely

Every `gc mail send` creates a permanent bead with a Dolt commit. The
`gc session nudge` path is ephemeral and costs zero. **Default to nudge for all
routine communication.**

**The litmus test:** "If the recipient dies and restarts, do they need this
message?" If yes -> mail. If no -> nudge.

**Ephemeral protocol messages:** MERGE_READY, MERGE_FAILED, RECOVERY_NEEDED,
LIFECYCLE:Shutdown, and WORK_DONE are routine signals. Use `gc session nudge`
— the underlying bead state (assignee, status, metadata) is the durable record.

**When you must mail**, use shell quoting for multi-line messages:

```bash
gc mail send <addr> -s "Subject" -m "$(cat <<'EOF'
Multi-line body here.
Shell quoting issues avoided.
EOF
)"
```

### Mail lifecycle: Read → Process → Archive

- `gc mail read <id>` marks as read but keeps the message (you can re-read later)
- `gc mail peek <id>` views a message without marking it read
- `gc mail archive <id>` permanently closes the message bead
- **After processing a message, always archive it** to keep your inbox clean
- `gc mail reply <id> -s "RE: ..." -m "..."` creates a threaded reply

**Doltlite health — your part:**
- Nudge, don't mail for routine communication
- Don't create unnecessary beads — file real work, not scratchpads
- Close your beads — open beads that linger become pollution
- When beads are slow/stuck: check `gc beads-doltlite health`, `gc doctor`, nudge Deacon — don't run maintenance blindly
{{ end }}
