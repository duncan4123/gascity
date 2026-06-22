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
- Build linked `gc`, `bd`, and `doltlite-client` with
  `gc beads-doltlite build all --install`.
- Use `doltlite-client` for direct test reads and writes. It supports `info`,
  `query`, `exec`, `show`, `set-metadata`, and `close`.
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
