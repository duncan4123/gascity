Build DoltLite-linked binaries from the Gas City and beads-doltlite source trees.

The default target is `gc`.

Use `gc beads-doltlite build gc --install --no-restart` for normal Gas City iteration, native read fastpath fixes, or build-tag changes.

Use `gc beads-doltlite build bd --install --no-restart` only after the
beads-doltlite source or bd link inputs change.

Use `gc beads-doltlite build client --no-restart` only when refreshing the
DoltLite diagnostic client.

Use `gc beads-doltlite build all --install --no-restart` only for a coordinated
rebuild that includes the optional diagnostic client.

`all` builds `bd`, `doltlite-client`, then `gc`; it does not skip unchanged
targets. Fresh `gc init` installs only the required `bd` and `gc` targets.

The command first looks for an installed `libdoltlite` and then downloads the
pinned DoltLite release library into the pack runtime cache when needed. Pass
`--lib DIR` or set `DOLTLITE_LIB`/`GC_DOLTLITE_LIB` to use a development build.
Likewise, local source checkouts are preferred, but the command can fetch
default Gas City and beads-doltlite sources into the pack runtime cache on a
fresh machine.

With `--install`, the `gc` target updates every distinct home-owned entrypoint
the city may use: the running supervisor binary, the configured supervisor unit
binary, and the active controller `gc` path. Symlinks are resolved before
writing so aliases such as `$HOME/go/bin/gc` keep pointing at the same real
installed binary.

Examples:

```bash
gc beads-doltlite build gc --install --no-restart
gc beads-doltlite build bd --install --no-restart
```
