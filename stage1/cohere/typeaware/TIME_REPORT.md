Built shared typeaware products and cached TSGo runtime objects for step 04.
Branch codex/typeaware-time, based on origin/main 031a1259.
The full package timed out before at 1800.053s and passed after at 652.503s.
All 47 compiler mutants and 19 malformed-frame input cases remain caught; two cache mutants also fail.
The full gate and load 66.9 were not replayed; the extra race build was stopped on disk pressure.

`BuildTSGo` now uses the existing native runtime archive builder. Its key includes every runtime source and header, the public TSGo ABI header, feature and representation flags, target, compiler path and compiler version. The selected Go checker archive changes the final link, not the runtime objects. Whole-archive linkage retains the translation units and their exit hooks, as the previous direct link did.

The package builds the stage-zero compiler, normal and sanitized checker archives, and unchanged volume Go oracle once. It lowers and emits each immutable entry once, then shares the completed C string. Emission's mutable IR is never shared. Unchanged binaries are reusable; single-use mutant binaries remain in their child directories. Every agreement and mutant execution still runs. Sources and archives are immutable for the package's lifetime; modified inputs have separate paths and keys.

Independent mutants run as parallel subtests, bounded by Go's test parallelism. Each child owns its test handle, sources and logs. Its Go subprocesses use GOMAXPROCS=1, and mutant archive builds use -p=1. No unit deadline or timing assertion was added. The package timeout remains 30 minutes.

Main does not carry internal/native/testbuildcache or internal/buildcache. The small keyed helper in shared_test.go follows the existing lint shared-product pattern. It owns products for the whole package and returns builder errors to all waiters. The CSS printer's parallel subtests provide the scheduling pattern. The estree tests on this base do not contain a shared-product helper.

The before measurement used the original build sequence through the same load, lowering, TSGo emission and build APIs, with timing logs added. It reached TestVolumeAgreementAndMutants/widened-shape before the package timeout. Its phase census is partial. The after measurement completed every test, including base-shapes, which the independent cohere byte oracle caught at byte 25376.

| Observation | Before | After |
|---|---:|---:|
| Full package test wall | 1800.053s, timeout | 652.503s, pass |
| Command wall including test compilation | 1809.604s | 664.605s |
| Six-rule suite including mutants | 656.170s | 143.828s |
| Volume profile suite including mutants | 522.900s | 156.698s |
| Volume suite including mutants | incomplete | 198.415s |
| Checker archive builds, count / summed elapsed | 29 / 443.868s | 37 / 858.569s |
| Loads, count / summed elapsed | 43 / 5.193s | 19 / 2.767s |
| Lowering, count / summed elapsed | 43 / 64.561s | 19 / 38.562s |
| C emission, count / summed elapsed | 43 / 561.789s | 19 / 286.774s |
| Native builds, count / summed elapsed | 43 / 559.040s | 57 / 595.677s |
| Runs, count / summed elapsed | 98 / 6.765s | 121 / 7.903s |
| Compiler AST mutant tests, summed elapsed | 20.830s | 9.835s |

Summed elapsed times overlap when mutants run in parallel. They are not CPU times and do not add to package wall time. The before run stopped before all its archive builds, native builds and executions. The after run therefore has more of those invocations while completing with fewer loads and emissions. Before, each native build made one clang invocation and included writing its temporary files. After, individual clang children were also recorded: 168 runtime object compilations took 18.664s in summed elapsed time; 57 program compile-and-link invocations took 561.318s, ranging from 0.185s to 41.004s; 148 cgo compilations and compiler probes took 13.331s. The runtime objects cover three distinct flag sets, then are reused across checker variants. Individual commands and timings are in verification/time/before.json and after.json.

These are observed package outcomes, not a controlled speedup ratio. The before run included a 170.573s cold checker build and overlapped an extra race-detector compilation. That compilation filled the root filesystem's existing Go cache and was stopped. Only large rebuildable cache products older than this unit and the stopped build's scratch directory were removed; source files and logs were retained. The after run had warm Go dependencies and a fresh runtime cache key through the timing wrapper. Its one-minute load started at 1.441 and ended at 3.385. No result is claimed for the unfinished race-detector check or the reported gate load of 66.9.

Every compiler mutant retains its previous catcher:

| Group | Mutants | Catcher |
|---|---|---|
| Profile | binding-slot, scope-containment, first-binding | Independent cohere finding bytes |
| Profile index | index-kind, root-kind, index-end | Compiler AST identity oracle |
| Shadow | missing-binding | Required panic 70 |
| Six rules | declaration-source, wrong-node, last-declaration, nullable-default, union-members, type-name, raw-shape-constraint, assignability-direction, resolved-signature | Independent cohere finding bytes |
| Six rules lifetime | retained-handle | Exit 0 contradicts required stale-query panic 70 |
| Six rules bounds | facts-length-asan | ASan heap-buffer-overflow |
| Six rules frames | facts-empty, facts-integer, facts-length, facts-version | Required panic 70 and pinned frame message |
| Unary rule | wrong-node | Independent cohere finding bytes |
| Unary lifetime and selector | stale registry, ignored-kind | Exit 0 contradicts required panic 70 |
| Unary frames | empty-frame, missing-header, bad-length | Required panic 70 and pinned frame message |
| Requests | wrong-kind, unknown-question | Exit 0 contradicts required panic 70 |
| Volume | assignable-types, widened-shape, enum-types, type-symbol, scope-locals, call-returns, property-shape, contextual-shape, symbol-origin, type-origin, property-info, call-count, call-parameters, apparent-shape, base-shapes | Independent cohere finding bytes |
| Volume lifetime | released-registry | Exit 0 contradicts required panic 70 |
| Configuration | strict-this | Required refusal and independent cohere message bytes |

The two added cache mutants were run through Go overlays. Removing entry identity from the C key made binding-slot, scope-containment and first-binding use the unchanged program. All three compiled and ran, then failed with "mutant survived byte oracle". Replacing a shared key with a fresh key per request failed TestSharedProductPublication with "built shared product 16 times". Neither was killed by a compiler warning. The publication test also checks completed-file visibility, key isolation and propagation of failed builds to waiters.

Commands and results, with output written directly to logs:

```text
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/typeaware -count=1 -timeout 30m -v
before: FAIL at 1800.053s, package timeout
after: PASS, 652.503s
go test ./stage1/cohere/typeaware -run '^TestInspectRequestRefusals$' -count=1 -timeout 30m -v
PASS, 47.443s
go test ./stage1/cohere/typeaware -run '^Test(SixPinnedFlags|PinnedTypeFlags|SharedProductPublication)$' -count=2
PASS, 0.029s
go test -overlay /tmp/typeaware-key-mutant.json ./stage1/cohere/typeaware -run '^TestVolumeProfileAgreementAndMutants$/(binding-slot|scope-containment|first-binding)$' -count=1 -timeout 30m -v
FAIL as required, all three source mutants caught the incorrect cache key
go test -overlay /tmp/typeaware-rebuild-mutant.json ./stage1/cohere/typeaware -run '^TestSharedProductPublication$' -count=1 -v
FAIL as required, 16 builds instead of one
go vet ./internal/native ./stage1/cohere/typeaware
PASS, no output
go build -o /tmp/typeaware-final-adamic ./cmd/adamic
PASS, no output
go test ./internal/native ./cmd/adamic -run '^Test(TSGoBuildSeesTheProgramsFeatures|WASIRejectsTSGoArchive)$' -count=1
PASS, native 0.598s; cmd/adamic 0.012s
git diff --name-only origin/main -- '*.a'
empty
```

No Adamic fixture was added or changed, so no a-check input or counts row changed. Optional external compiler/repository corpora were unset, as in the before run. Their existing agreement checks and benchmark loops remain unchanged. The full gate was not run.

Setup used GOPROXY=https://proxy.golang.org|direct. Its timing lines were: Node ready 0.049s; Go ready 0.073s; markdown dependencies ready 0.164s; submodules ready 0.169s; clang ready 0.530s; Go build ready 20.756s; test binaries deferred 21.144s; build cache warm 21.147s; done 21.200s. nproc=5, cgroup cpu.max=400000 100000. Tools were Go 1.27.1, clang 20.1.8 and Node 24.19.0.

Logs are /tmp/typeaware-setup.log, before.log and after.log with the typeaware- prefix, /tmp/typeaware-runtime-smoke.log, /tmp/typeaware-final-compile.log, /tmp/typeaware-key-mutant.log, /tmp/typeaware-rebuild-mutant.log, /tmp/typeaware-vet.log, /tmp/typeaware-build.log and /tmp/typeaware-final-native.log. The stopped race build's log is /tmp/typeaware-cache-race.log. verification/time/measure.py summarizes the verbose phase logs. To record clang children, copy verification/time/clang.py to a scratch bin directory as clang, make it executable, set ADAMIC_REAL_CLANG to the real compiler's absolute path, set ADAMIC_CLANG_MEASURE to an existing output directory, and put the scratch bin first on PATH. The collector preserves compiler output and exit status.
