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

The `bd` target is pinned to beads-doltlite commit
`61fadf8c5d87f929bb47a3b23350f53b5960c1eb` by default. Pass `--bd-ref` or set
`GC_BEADS_DOLTLITE_BD_REF` to override the source revision check.

With `--install`, the `gc` target updates every distinct home-owned entrypoint
the city may use: the running supervisor binary, the configured supervisor unit
binary, and the active controller `gc` path. Symlinks are resolved before
writing so aliases such as `$HOME/go/bin/gc` keep pointing at the same real
installed binary.

Examples:

```bash
gc beads-doltlite build gc --install --no-restart
gc beads-doltlite build all --install --no-restart
```
