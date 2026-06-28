---
name: doltlite
description: Use when working with DoltLite-backed Gas City or Beads storage, the beads-doltlite pack, doltlite-client diagnostics, libdoltlite-linked builds, DoltLite SQL operations such as dolt_gc/hash functions/remotes, or debugging DoltLite locks, maintenance, flatten, gc, and native read fast path behavior.
---

# DoltLite

Use this skill for Gas City work involving the `beads-doltlite` backend or the
local `doltlite` checkout.

## Source Of Truth

Read the local DoltLite README before changing DoltLite semantics:

- `<city-root>/doltlite/README.md`
- `../../../../../doltlite/README.md`
- `../../../../tools/doltlite-client/README.md`

The README documents DoltLite's SQLite-compatible API and Dolt SQL functions.
Prefer that contract over assumptions from Dolt-server commands.

## Gas City Rules

- DoltLite cities use `[beads] backend = "doltlite"` and bead scopes under
  `.beads/doltlite/*.db`.
- Do not require a Dolt SQL server, runtime port, or
  `.gc/runtime/packs/dolt/dolt-state.json`.
- For normal Gas City iteration or native read fastpath fixes, build only `gc`
  with `gc beads-doltlite build gc --install --no-restart`.
- Fresh DoltLite init builds the required binaries with
  `gc beads-doltlite build bd --install --no-restart` and
  `gc beads-doltlite build gc --install --no-restart`. It should not require
  `doltlite-client`, sqlitebrowser, or a local DoltLite source build.
- Use `gc beads-doltlite build all --install --no-restart` only for coordinated
  rebuilds that include the optional diagnostic client. The build command may
  download a pinned DoltLite release library into pack runtime state; pass
  `--lib` only for an explicit development build.
- Use `doltlite-client` for direct test reads and writes. It supports `info`,
  `query`, `exec`, `show`, `set-metadata`, and `close`.
- Use `gc beads-doltlite sqlitebrowser build/open` for optional GUI inspection;
  stock SQLite Browser builds cannot open DoltLite-format databases. Plain
  `open --city <city>` generates a DB Browser project with HQ, rig attachments,
  and a formula-progress SQL tab.
- Use DoltLite SQL for native maintenance checks, including
  `SELECT dolt_gc();` for GC.
- Do not assume configurable SQLite checkpoint modes. DoltLite rejects
  `PRAGMA wal_checkpoint(TRUNCATE)` on DoltLite-format databases; use the
  default checkpoint form when probing.
- Treat `bd flatten` and `bd gc` as Beads CLI behavior, not the canonical
  DoltLite client oracle.
- Do not run heavyweight flatten or GC synchronously inside city startup.
  Keep non-critical maintenance bounded and non-fatal.
- Before debugging lock issues, check for active `bd`, `doltlite-client`,
  `gc session list`, and `gc-beads-bd` processes. Kill only probes you started
  or processes the user has approved.

## Pack Troubleshooting

When a Gas City pack formula appears stuck under the DoltLite backend, treat
each live query and write as a test case. Do not stop at checking whether a
bead exists.

For every controller, worker, formula, or helper operation involved:

1. Capture the exact operation the system runs: the `bd` command, shell
   pipeline, jq filter, Go store call, or metadata write.
2. Run the equivalent `bd` command from the rig root against the live
   DoltLite-backed store.
3. Run an equivalent direct read with `doltlite-client -db
   <rig>/.beads/doltlite/<prefix>.db query '<SQL>'`.
4. Compare the result of the direct DoltLite query with the `bd` result and the
   next controller lifecycle decision, such as pool demand, session
   materialization, hook claim, continuation assignment, drain acknowledgement,
   or finalization.

For ready-work bugs, verify the fully qualified route and blocking state with
both surfaces. The controller scale/query path should agree with a `bd ready`
probe such as:

```bash
bd ready \
  --metadata-field 'gc.routed_to=<rig>/<agent>' \
  --unassigned \
  --exclude-type=epic \
  --json \
  --sort oldest \
  --limit=20
```

Then check the same predicate directly in DoltLite, including `status`,
`assignee`, `issue_type`, `is_blocked`, and metadata such as `gc.routed_to`,
`gc.run_target`, `gc.kind`, `gc.session_affinity`, `gc.root_bead_id`,
`gc.root_store_ref`, and `gc.continuation_group`.

If `bd` and `doltlite-client` agree that work is ready but no session starts,
the bug is in route qualification, pool demand, desired-state construction, or
session materialization. If they diverge, debug the Beads/DoltLite fast path or
metadata/index translation before changing formula logic.
