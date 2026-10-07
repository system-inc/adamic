Built two WASI-only empty-path hooks and registered empty-path.a in the input oracle.  
Fix commit: `aa6801ca590c3fa601b4b56e2bd4c1090632485e`; base: `origin/area/runtime` at `1525884`.  
V8 empty-path: before stdout mismatch, after exact agreement; input oracle, counts and native packages pass.  
Mutants: independently disabled listing/status hooks compiled and failed V8 stdout comparison; 96 native objects are identical.  
Not covered: complete repository gate, every WASI fixture on this branch, ARM, reactor or deployment.

The hooks are in `internal/native/runtime/directory.c`: conditions at lines **73** and **145**, returning ENOENT at **74** and **146**, guarded at **71** and **143**. They run after output flushing and before libc path conversion, matching input.c's normalization. Native preprocessing excludes them.

Node and corrected V8 both exit 0 with empty stderr and these exact stdout bytes:

```text
cannot read directory : no such directory
cannot read status of : no such file
```

Before the fix V8 instead printed `listed a directory\ndirectory\n`, also exit 0 with empty stderr. This is a shared WASI runtime bug, not an engine difference. The source is `internal/oracle/testdata/empty-path.a`; its input registration and counts row are the only additional oracle edits. Counts are allocations 4, frees 4, retains 4, releases 8, peak 4, regions 0.

Commands used (all output saved to the corresponding `.log.gz` evidence):

```sh
source /workspace/adamic-tools/env.sh
export GOFLAGS=-buildvcs=false
go build -o /tmp/empty-adamic-before ./cmd/adamic
/tmp/empty-adamic-before build --target wasm32-wasi internal/oracle/testdata/empty-path.a -o /tmp/empty-before.wasm
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/empty-path.a
node --disable-warning=ExperimentalWarning oracle/wasi.mjs /tmp/empty-before.wasm
# After applying the two hooks, rebuild the compiler and module and repeat.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 15m
go test ./internal/native/... -count=1 -timeout 15m
go vet ./internal/native ./internal/oracle
```

V8 coverage here: **1 fixture, 1 agreement after the fix, 0 failures, 0 skips**. Before: that same fixture disagreed only in stdout. The input oracle ran all **7 input fixtures** successfully (3.124 s). The counts update passed (52.314 s), with only the new row changing. Native package tests passed (126.609 s). Vet, gofmt and whitespace checks passed.

`native-object-proof.py` compiled all 48 runtime translation units with native clang 20.1.8, both ordinary and `ADAMIC_COUNT`, before and after. It reported `objects 96` and `differing []`. `native-objects.tsv` records both SHA256 hashes for every object. The hooks' two independent mutants were compiled in isolated scratch C objects and linked to the unchanged emitted fixture; neither altered the checked-in runtime. The listing mutant printed `listed a directory` while the status error stayed correct; the status mutant printed `directory` while the listing error stayed correct. Both preserved exit 0 and empty stderr and were caught only by stdout comparison.

Tools reused the installed W5 environment: Go 1.27.1, Node 24.19.0, WASI SDK 27, native clang 20.1.8, nproc 5. Original setup timing: combined SDK/engine setup 200 s, SDK retry 79 s. The new worktree shared the initialized cohere submodule at the identical commit; VCS stamping was disabled for the initial shared-path build. No main or area branch was pushed or merged into this runtime-only change.
