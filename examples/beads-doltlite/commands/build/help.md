Build DoltLite-linked binaries from the Gas City and beads-doltlite source trees.

The default target is `gc`.

Use `gc beads-doltlite build gc --install --no-restart` for normal Gas City iteration, native read fastpath fixes, or build-tag changes.

Use `gc beads-doltlite build bd --install --no-restart` only after the
beads-doltlite source or bd link inputs change.

Use `gc beads-doltlite build client --no-restart` only when refreshing the
DoltLite diagnostic client.

Use `gc beads-doltlite build all --install --no-restart` for bootstrap or a coordinated rebuild

`all` builds `bd`, `doltlite-client`, then `gc`; it does not skip unchanged targets and it does not build libdoltlite itself.

The command expects an existing libdoltlite build. Pass `--lib DIR` or set
`DOLTLITE_LIB`/`GC_DOLTLITE_LIB` when the default `doltlite-work/build` or
`doltlite/build` paths are not correct.

Examples:

```bash
gc beads-doltlite build gc --install --no-restart
gc beads-doltlite build all --install --no-restart
```
