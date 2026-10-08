Built: static and dynamic RegExp replace/replaceAll function replacements on native, JavaScript and WASI.
Commits: 96bb40f6b8bf1861ec7d07f1960d47f19615c96c; 132d99194fe1560c4c527171f017b611bd976897.
Commands/results: Node 24.19.0 fixture gate 11.848s; native regression 84.682s; test262 zero disagreements.
Mutants: three Node output witnesses, two argument guards, unused-runtime symbol check, and three callback-effect witnesses all caught.
Uncovered: Symbol.replace tests feature-skipped; custom exec/ToPrimitive refused; host fixture WASI filesystem operation refused; full host integration gate not run.

## Behavior and implementation

The callback receives match, captures in source order, UTF-16 offset, original input, then named groups when present. Missing captures are undefined. Immediate arrow callbacks support typed rest packs and primitive returns converted to strings; replacement strings are literal, including dollar sequences. Global matching finishes before callbacks execute, including empty Unicode advancement and global lastIndex reset. Sticky success/failure and callback lastIndex mutations are Node-compared. Callback exceptions propagate after releasing collected matches and argument holds.

New slice files are internal/lower/library_regexp_replace.go, internal/native/library_regexp_replace.go, internal/javascript/library_regexp_replace.go and internal/native/runtime/regexp_replace.c. Small shared hooks provide intrinsic rest/union typing, exception edges, and conservative global ownership around hidden calls. There is no matcher change and no native.go hook in this unit. The new C file has no mutable static or global state; docs/runtime-statics.md lists it.

The native emitter writes an exact ADAMIC_REGEXP_REPLACE_CALLBACK directive. RuntimeLibraryForSource opts in only for that directive; test262 chooses the matching cached runtime variant per program. This avoids linking callback code into native whole-archive builds as well as unused WASI builds.

## Linux and WebAssembly gates

Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0, WASI SDK 27; nproc 5, CPU quota 4. Setup timing: Go/Node 0.025s, submodules 0.072s, clang 0.192s, build 26.092s, total 26.325s. GOPROXY=https://proxy.golang.org|direct. WASI_SYSROOT explicitly exported after setup.

Commands (test output recorded directly to files):

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestWASIAgreesWithNode)/internal/oracle/testdata/regexp_replace/|^TestRegExpReplacement.*Mutants$' -count=1 -v
go test ./internal/native ./internal/flow ./internal/ir -run 'TestRegExp|TestRuntimeStaticsAreListed|Test.*Throw|Test.*CallTargets|TestClosureTargets' -count=1 -timeout 15m
go test ./internal/load ./internal/lower ./internal/javascript -count=1
go test ./cmd/adamic-test262 -count=1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -update-counts -count=1
```

All eight new fixtures pass both backends and WASI, including dynamic patterns, named groups/rest packs, callback throws and hidden global effects. Annotation guard fixtures use the inserted JavaScript guard as truth. Linux runs sanitizers, leak checks, release and slab variants. WASI runs the SDK runtime without sanitizers; no sanitized WASI runtime was available. Native regression also reruns the inherited compiler identity, rejection, divergence and WASI gates. Flow package compiled but the selected filter had no tests there. Full load/lower tests passed (2.099s/50.813s); runner tests passed (31.045s); counts passed (100.017s). Vet and diff checks passed. Runtime statics scan ran before each push.

Counts include the new fixture rows and two inherited changes: dynamic flags_errors had stale prior counts; node_fs_directory_system observes the added testdata directory and consequently allocates more directory entries.

## test262 before and after

Corpus 7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd. cmd/adamic-test262 with -adapt -jobs 4 -json and the two requested directory filters. Baseline ef3bbf53, implementation branch. Attached JSON preserves refusal reasons. No newly passing stock tests; callback coverage is in the Node-compared fixtures.

Before:

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/String/prototype/replace | 13 | 0 | 28 | 0 | 14 |
| built-ins/RegExp/prototype/Symbol.replace | 0 | 0 | 0 | 0 | 70 |

After:

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/String/prototype/replace | 13 | 0 | 28 | 0 | 14 |
| built-ins/RegExp/prototype/Symbol.replace | 0 | 0 | 0 | 0 | 70 |

All Symbol.replace cases remain skipped by the existing runner feature gate; these counts do not certify that directory's algorithms. The directory measurement preceded the final opt-in directive/effect hooks; those directories have no passing native callback cases affected by the hooks. The final dedicated fixtures and native regression ran after those changes.

## Mutants

TestRegExpReplacementNodeMutants changes offset by one, swaps two captures, and drops the named-groups argument. Each mutated C program compiles cleanly, finishes with exit zero and empty sanitizer stderr, then fails solely against Node stdout.

TestRegExpReplacementTypeGuardMutants removes actual argument-kind and required named-field checks. Both finish cleanly but disagree with the JavaScript guard exit status.

TestRegExpReplacementOptInMutant always enables callback runtime code. TestRegExpRuntimeCompilerNotLinked catches the callback symbol in hello.

internal/native/testdata/run-regexp-replacement-effects-mutants.py independently disables global lending protection, global move protection and exception edges. Node comparison and sanitizers catch heap-use-after-free, a null moved-global access and a null access after a lost exception edge respectively. No mutation was certified by a build error or warning.

## Pay when used

Common-baseline unused WASI files are byte-identical, not just equal size. Brotli quality 11.

| Program | Before raw | After raw | Before Brotli | After Brotli |
|---|---:|---:|---:|---:|
| hello | 288644 | 288644 | 84071 | 84071 |
| request.a | 325240 | 325240 | 97552 | 97552 |

Native hello: 451024 raw / 161919 Brotli; no callback, compiler or compiler-owner symbols. Native .text and .rodata sizes/addresses are unchanged on the baseline comparison (0x2bdee and 0x2937b). Native compressed metadata can vary with cache paths. Dynamic callback fixture cost: native 1896824 raw / 293689 Brotli, WASI 856611 raw / 172075 Brotli. Measurements were made at the opt-in implementation before the final effect hooks, which only apply to callback programs.

## Host scratch merge and platform relay

Scratch merge 87ba1951 combines this implementation with host-proof-combined 08b5b2c; the scratch branch is local, not pushed. Host closure ABI adds argument_count. Scratch callback dispatch supplies packed_count, and two parallel callback call sites supply their counts. Conflict resolutions preserve host census/rest handling and current atomic slot caches and field initialization tracking. The cohere pin is identical on both branches (7945d102a6c18dd36adf9114a758ce646e8b2359).

Relay to #adamic_runtime_platforms: codex/regex-replace-callback, implementation 132d9919; new regexp_replace.c has no mutable globals, runs on WASI, opts in only for callback programs. Host-proof-combined has a closure argument-count ABI overlap: callback dispatch needs its third packed_count argument. That adaptation is proven only on the scratch merge. Full combined host/concurrency gate remains integration work.

The final host proof passes native and JavaScript in 28.139s, with Node stdout `false\nfalse\ntrue\nfalse\naBc_09.TS\n` and exit zero. The separate WASI subtest skips with an explicit fs.mkdtempSync target refusal; no replacement approximation is used. See host-24.log and the retained scratch test source.
