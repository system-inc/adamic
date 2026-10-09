# Host Error identity port to area/library

Base: `e1c26db3c73e7b4b48739080fa674a2c7bc687a2`, the newest fetched
`origin/area/library`. Only the host tagging fix from `6054117e` is ported;
no combined-proof compiler history is imported. The symlink and stat-options
widening witnesses already match that fix byte for byte on this base.

Host errors retain their own string code and message and the existing three-slot
ownership layout. Immutable Error, TypeError and RangeError descriptors supply
nominal identity, with TypeError and RangeError inheriting Error. Filesystem,
process and generic runtime error constructors attach those descriptors.

The area compiler predates the built-in Error lowering from the combined proof.
Assumption: that compiler work belongs to the incoming embed, not this runtime
port. Accordingly, the source witness is run on Node and the same host operations
are exercised directly through the runtime on sanitized native Linux and WASI.
Both outputs must match Node byte for byte. This probe avoids the area's existing
catch-guard constant folding and requires actual nominal identity. It is not
registered as an ordinary lowered fixture on this older compiler.

The WASI drop-tag mutant clears every constructed object's class tag. It exits
cleanly with no stderr and produces six `not Error`/`other` pairs. Only Node's
stdout comparison catches it. The good probe passes native ASan, UBSan and
LeakSanitizer as well as WebAssembly. Existing host comparisons also pass.

Linux toolchain: Node 24.19.0, Go 1.27.1, clang 20.1.8, WASI SDK 27.
Setup took 58.468s: Go 0.045s, Node 0.044s, markdown 0.131s,
submodules 0.172s, clang 0.313s, WASI SDK 0.533s, build 58.199s,
cache warm 58.432s. `nproc` is 5, CPU quota 4, memory 17.6 GB.
Environment: `/workspace/adamic-tools/env.sh`. API `npm ci` succeeded.

Commands (all test output sent directly to logs):

```sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIHostErrorIdentity|TestWASIInputAgreesWithNode|TestWASIFileAgreesWithNode|TestInputAgreesWithNode|TestNodeFSFileAgreesWithNode|TestNodeFSDirectory.*)$' -count=1 -v -timeout 30m
go test ./internal/oracle -count=1 -v -timeout 30m -args -update-counts
go test ./internal/native -count=1 -timeout 30m
go vet ./internal/native ./internal/oracle
gofmt -l internal/oracle/wasi_host_error_test.go
git diff --check
```

Host comparisons passed in 90.520s, including the tag mutant. The complete
native package passed in 305.357s. Vet, formatting and diff checks passed.
The whole oracle passed in 617.195s, including counts (385.72s). Linux
counts regeneration produced no changes to the existing table; no proof-branch
rows were copied. No deadline-only failures occurred.

Limits: no full repository gate is claimed. The new source witness's built-in
Error lowering is tested through its runtime interface here, pending the embed;
this port does not import or change lowering.
