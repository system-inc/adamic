# The wasm32 errno table's mutant (#z4a9p9e)

Measured through `ahra loom job` (J-seexmnn) at 2353bb30644c5f825ad089d443952e4dbc81b9f0, on a scratch box with Go 1.27.1, Node v24.19.0, `WASI_SYSROOT` set and wasmtime present, `GOFLAGS=-buildvcs=false`, `ADAMIC_GATE_UNCACHED=1`. The edit was applied once (`sed`, counted at 1), the two tests ran, and the header was restored from a saved copy; `git diff --quiet` held afterwards. The same job first ran the oracle's `^(TestWASI|TestWasmtime)` families with `ADAMIC_ORACLE_WASI=1 ADAMIC_ORACLE_WASMTIME=1`, 5834 runs, all passing.

| Mutant | Exact edit | Red |
| --- | --- | --- |
| The map dropped | internal/native/runtime/node_errno.h: `#ifdef ADAMIC_TARGET_WASI` made `#ifdef ADAMIC_TARGET_WASI_DROPPED`, so a wasm32 build takes `-(double)error`, wasi-libc's own numbers | TestWASIErrnoIsNodesOnLinux: Node prints `EPERM -1`, `ENOENT -2` and the rest of Linux's numbers and `UNKNOWN -4094`; wasm32 prints `EPERM -63`, `ENOENT -44`, `EINTR -27`, `EACCES -2` and the rest of `WASI`'s numbering, and `UNKNOWN -75` (EXDEV's). TestWASIEmptySymlinkAgreesWithNode_001/control/wasm32-wasi: stdout differs, the runtime's second line `-44` where Node prints `-2` |
