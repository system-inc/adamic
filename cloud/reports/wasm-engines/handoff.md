Runtime ownership: the fix and input fixture are on `codex/wasi-empty-path`; the engine branch retains exact walk.a assertions for the pending bug.

# Empty directory and status paths on WASI

Observed: Node source returns `cannot read directory : no such directory` and
`cannot read status of : no such file`. The same source built through
`adamic build --target wasm32-wasi` returns `listed a directory` and `directory`
on both wasmtime 38.0.3 and V8 / Node 24.19.0, with exit 0 on all three runs.

The minimal source has moved to `internal/oracle/testdata/empty-path.a` on `codex/wasi-empty-path`. The
registered fixture `internal/oracle/testdata/walk.a` also catches both errors.
This is a real Adamic WASI runtime portability bug, shared by both engines.
It is not a bug that only V8 hides. The existing WASI oracle's ordinary fixture
list does not include the separate input fixture list; this unit includes it.

Inference: WASI libc treats empty directory/status paths as the current preopen.
The WASI path normalization used for reads/writes needs equivalent coverage for
`readDirectory` and `fileStatus`. The responsible runtime files are outside this
unit's ownership and were not changed.

Reproduce from the repository root after sourcing the setup environment:

```sh
go build -o /tmp/adamic ./cmd/adamic
/tmp/adamic build --target wasm32-wasi internal/oracle/testdata/empty-path.a -o /tmp/empty-path.wasm
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/empty-path.a
node --disable-warning=ExperimentalWarning oracle/wasi.mjs /tmp/empty-path.wasm
wasmtime run --dir /::/ /tmp/empty-path.wasm
```

Run these commands on `codex/wasi-empty-path`: the source is now an input oracle fixture with its required count-table row. The engine branch keeps the pending runtime disagreement asserted by name until that fix is integrated.
