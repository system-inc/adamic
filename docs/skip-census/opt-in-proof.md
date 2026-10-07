# An enabled opt-in cannot skip

Based on `10e709cef5b6190c42fe8249f83d2cde20f5a402`, on
`devtools/skip-census`. No main push or history rewrite.

The scanner derives opt-in annotations from AST guards, sequential continuation,
callbacks and direct helpers. TestCensus rejects every post-opt-in skip unless
it is required-input. Six skip sites became Fatalf with their messages retained:
four TestWASI prerequisites, the release helper's Node exit exclusion and
TestWASIEmission's non-lowering fixture exclusion. The table now has 70 sites.
Parser performance also fails if opted in without the pinned compiler source.

## Commands and observations

All test output was redirected to the linked logs. Tool environment was sourced
from `/workspace/adamic-tools/env.sh`. Tests use `-count=1`.

| Check | Command and environment | Observed result |
| --- | --- | --- |
| Census | `go test -v -count=1 ./internal/skipcensus/...` | Exit 0; TestCensus and generic opt-in tests pass; historical log still yields exactly 17 required-input skips among 33. |
| Native clang mutant | `ADAMIC_TEST_WASI=1 ADAMIC_GATE_UNCACHED=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot PATH=/workspace/adamic-tools/llvm/bin:$PATH go test -v -count=1 -timeout 30m ./internal/native -run '^TestWASI$'` | Exit 1; TestWASI fails naming missing `libclang_rt.builtins.a`. |
| WASI SDK clang | Same command with `/workspace/adamic-tools/wasi-sdk/bin` first on PATH | Exit 0; 36 case subtests and requests pass. |
| Missing parser corpus | `ADAMIC_PARSER_BENCH=1 go test -v -count=1 ./stage1/typescript/parser -run '^TestPerformance$'`, with ADAMIC_TYPESCRIPT_SOURCE unset | Exit 1; names pinned v6.0.3 checkout. |
| Oracle opt-out | `go test -v -count=1 ./internal/oracle -run '^(TestWASIEmission|TestNativeReleaseFlagsAgreeWithNode|TestNativeExistingReleaseAgreesWithNode)$'`, lane switches unset | Exit 0; off guards still skip. |
| Vet | `go vet ./internal/skipcensus/... ./internal/native ./internal/oracle ./stage1/typescript/parser` | Exit 0. |

Logs: [native failure](opt-in-proof/native-clang.log),
[SDK success](opt-in-proof/wasi-clang.log), [census](opt-in-proof/tests.log),
[parser failure](opt-in-proof/parser-missing.log),
[oracle opt-out](opt-in-proof/oracle-off.log), [vet](opt-in-proof/vet.log).

## Mutants

In a scratch copy, changing the classification invariant to `if false` makes
`TestOptInClassification` fail for all three unrelated switch names because a
not-applicable enabled prerequisite survives. Dropping the sequential enabled
state also fails all three because the state is lost through a nested helper.
The tests additionally reject measurement and opt-in-lane classifications.
See [rule mutant](opt-in-proof/rule-mutant.log) and
[flow mutant](opt-in-proof/flow-mutant.log).

A Go overlay restored the old WASI probe Skipf. With the identical real SDK
sysroot and native clang, Go exited 0 and reported SKIP while naming the same
missing builtins. This demonstrates the previous green-gate failure directly;
the fixed source returns 1 for this box configuration. See
[restored skip](opt-in-proof/restored-skip.log).

## Setup and limits

`bash cloud/setup.sh --wasi-sdk` exited 1 when downloading
`github.com/klauspost/compress@v1.20.0`: the Go module proxy's redirect to
storage.googleapis.com returned Forbidden. Existing tools allowed targeted tests.
The official wasi-sdk 27.0 x86_64 Linux archive was downloaded from its GitHub
release and extracted to `/workspace/adamic-tools/wasi-sdk`, using setup's pin.
WASI_SYSROOT was exported explicitly because setup stopped before writing it.
See [setup log](opt-in-proof/setup.log), with the signed redirect query omitted.

Setup printed Go ready 0.106s, submodules ready 0.144s, clang ready 0.332s;
Node installation step 4.922s and ready 5.059s; Markdown step 1.078s and ready
6.197s; stage3 API step 1.334s and ready 7.586s. These are setup diagnostics,
not before/after performance claims. Build-flags context: base commit above,
`nproc=5`, cgroup `cpu.max=400000 100000`, Go 1.27.1 linux/amd64,
native clang 20.1.8, SDK clang 20.1.8-wasi-sdk, Node v24.19.0;
uncached WASI runs. Load before was not captured; after was 0.95 0.38 0.15.
No cache or performance optimization was added.

The full gate, the release lane with opt-in on, and the full WASI emission oracle
were not run. Their former fixture-scope exclusions now fail after opt-in; this
unit does not make unsupported fixtures compile or non-finishing fixtures finish.
The AST flow analysis handles conventional Getenv off guards and direct helpers;
it is not arbitrary Go control-flow or indirect-call analysis.
