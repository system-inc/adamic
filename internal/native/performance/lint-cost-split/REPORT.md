Measured batch8 at -O2 without sanitizers and prototyped numeric-kind listener selection in scratch only.
Commits: profile base 5f070b5599635c459ce5870f6e74bf6cceac037b; main f8013f0baac41ddc340d76f83bddde38536a8f07; pre-report tip 477c653196f3149c76588e7f8c0529bd282c6b48; final report commit is this file's commit.
Commands and outputs: interleaved best-of-five native 2.247552 s, prototype 1.565212 s, Go 0.360563 s; 96,624 user CPU samples; 161 findings.
Mutants: missing StringLiteral listener and incorrect numeric kind each lose a Go finding; dropping a profile sample fails the partition assertion.
Not covered: type-aware bridge, other architectures, full repository gate, exact off-CPU attribution, and production Go combined-listener timing.

The kind table saves 0.682340 s (30.36%) versus today's batch8. It removes 97.131% of rule invocations while preserving the serialized findings. The current driver takes 2.248 s here, rather than reproducing the earlier 2.11 s exactly. The split below normalizes the measured CPU distribution to that earlier 2.11 s and labels the unsampled wall-time remainder; these are estimates, not a claim that a fresh run took 2.11 s.

| Driver | Best of five | Relative to Go | All five wall times (s) |
|---|---:|---:|---|
| Current batch8, release | 2.247552 s | 6.23x | 2.464535, 2.342423, 2.272620, 2.247552, 2.264519 |
| Kind table prototype, release | 1.565212 s | 4.34x | 1.700131, 1.719889, 1.565212, 1.708137, 1.784203 |
| All-listeners table control, release | 2.370773 s | 6.58x | 2.370773, 2.428417, 2.501116, 2.508338, 2.473013 |
| Pinned Go comparison | 0.360563 s | 1.00x | 0.396270, 0.386416, 0.505508, 0.453836, 0.360563 |

Native build: clang 20.1.8, LLVM commit 87f0227cb60147a26a1eeb4fb06e3b505e9c7261, x86_64-unknown-linux-gnu. All three timed native binaries use the exact command below, executed in their respective artifact directories. `-g` supplies DWARF only. No sanitizers, allocation counters, LTO or architecture tuning are enabled.

```sh
/workspace/adamic-tools/llvm/bin/clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -o scanner main.c adamic.c array.c array_from.c bitwise.c case.c class_features.c class_inheritance.c class_static.c closure.c count.c directory.c dtoa.c exceptions.c from_codes.c heap.c hypot.c ieee754.c input.c json_stringify.c library_array.c library_language.c library_math_number.c library_object.c map.c map_set.c math.c maybe.c normalize.c number.c object.c parse.c radix.c regexp.c region.c set.c sort.c sort_undefined.c spread.c stack.c string.c string_append.c string_from.c string_index.c string_share.c tsgo.c union.c utf8.c weak.c -lm
```

Go: go version go1.27.1 linux/amd64, normal optimizing gc compiler, executable build mode, GOAMD64=v1, CGO_ENABLED=1, no race instrumentation or gcflags. The unchanged batch8 oracle is compiled through its original virtual-main overlay. Its embedded build record is in [provenance.json](evidence/provenance.json). The scratch Adamic compiler is built with `go build -buildvcs=false -o /workspace/lint-cost-adamic ./cmd/adamic`; this flag addresses the scratch submodule symlink and does not change the oracle's optimization mode.

The compiler and runtime include integer fast paths a183e50 and the four runtime fixes from the preceding unit. Main's integrated Map/Set implementation now supersedes the earlier e7ea1a4 scratch hash implementation. The compiler is built from scratch tip a0e690548da3dc60a90dc09f1d326c6a0e5f21dc, including current main. Baseline C is generated while the parser nodes file is original, then the numeric-kind node prototype is restored for both table variants. This avoids accidentally measuring a baseline containing the prototype. Preliminary measurements generated with a compiler lacking the integer hooks were discarded; all figures and profiles here use the compiler recorded in provenance.json.

Machine: Intel Xeon Platinum 8573C, Linux 6.18.44, `nproc` 5, cgroup `cpu.max` 400000 100000 (four CPUs), 16 GiB. Toolchain setup took 111 s: go/clang/node/submodules 0 s, cache 111 s. Timing load before: `0.93 1.62 1.12 1/203 49102`; after: `0.96 1.56 1.11 1/202 49175`. Four drivers run sequentially in a rotated order for each of five rounds; every timed run exits zero, has empty stderr and prints `161`. No local build or test was running during timing or profiling. Raw wall/user/system results and ordering are in [timing.json](evidence/timing.json).

The input is the entire 77-file TypeScript compiler manifest pinned by the preceding unit: TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 (6.0.3), cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Batch8 driver origin is 4189abd, with the preceding unit's scratch integration of batch4-typescript 63782c5 and integer-fast-paths a183e50. The input was not reduced. [compiler.txt](evidence/compiler.txt) records every path. Full serialized output is 11,441,458 bytes, SHA-256 b5ed3ec40d59abaf2c491872e5f143a2a16db08cfa971dcc2ff188a22f6c13cf, identical across Go, original native, prototype native, all-listeners native and source prototype under Node 24.19.0. [parity.json](evidence/parity.json) records exit codes, empty stderr and hashes.

## Exclusive sampling split

```sh
perf record -e task-clock:u -F 1999 --call-graph dwarf,16384 -o perf.data -- bash -c 'for round in $(seq 1 20); do /workspace/lint-cost-baseline/scanner --manifest /workspace/lint-cost-baseline/compiler.txt --count; done'
perf script -i perf.data --show-mmap-events --inline -F comm,pid,time,event,ip,sym,dso
```

Every retained sample receives one bucket. Sampling uses user-mode task-clock, not kernel time or sleep time. MMAP events undo ASLR; LLVM addr2line expands inlined scopes at each stack address, including caller return addresses. This matters because clang inlines rule bodies into visit. Merely assigning all visit frames to the tree walk would hide rule execution. Source line guard boundaries, per-process totals, and symbols are in [split.json](evidence/split.json); the complete sampled stacks, DWARF resolution and one-row-per-sample audit are compressed alongside it.

Bucket precedence is physical leaf work first: retain/release and object-child destruction; allocation/free and allocation helpers; virtual/callee/closure lookup helpers; then string helpers (equality, slicing, UTF-16 indexing/decoding, string conversion). For libc or generic helpers, walk from the nearest expanded caller until one of those physical costs or a semantic owner is found. Context node/kind/parent/child access, visit and child iteration belong to tree walk. Rule enabled checks and entry prefixes up to the first rule-specific action belong to rule selection; bodies and their local helpers belong to their named rule. Scanner/parser methods belong to scanner/parser except Parser.node called by Context or a rule, which is node access. Reads, arguments, writes, the loader and remaining main/collector frames belong to startup/I/O. The shell/seq harness gets an explicit startup row. Anything unresolved remains in the remainder. These rules intentionally charge allocation/string/refcount instructions to their physical categories even when the caller is the parser or a rule.

Virtual dispatch here measures lookup helper CPU. Instructions for an indirect call itself stay in the caller's tree/selection/parser bucket; sampling cannot infer their counterfactual dispatch cost. String equality and UTF-16 costs share one string row, with individual symbols available in split.json. The rule logic rows exclude their physical runtime work and their entry selection. A zero means no sample, not proof that a rule costs no time.

| Bucket | Samples | User CPU share | Estimated part of 2.11 s |
|---|---:|---:|---:|
| retains and releases | 19,758 | 20.448% | 0.42019 s |
| virtual dispatch | 407 | 0.421% | 0.00866 s |
| allocation and freeing | 7,541 | 7.804% | 0.16037 s |
| string work | 16,366 | 16.938% | 0.34805 s |
| tree walk and node access | 11,174 | 11.564% | 0.23764 s |
| rule entry and selection | 6,333 | 6.554% | 0.13468 s |
| rule logic: no-octal-escape | 11 | 0.011% | 0.00023 s |
| rule logic: no-unexpected-multiline | 1,145 | 1.185% | 0.02435 s |
| rule logic: no-unused-private-class-members | 6 | 0.006% | 0.00013 s |
| rule logic: no-useless-constructor | 92 | 0.095% | 0.00196 s |
| rule logic: prefer-template | 58 | 0.060% | 0.00123 s |
| rule logic: react/forward-ref-uses-ref | 31 | 0.032% | 0.00066 s |
| rule logic: react/jsx-no-comment-textnodes | 0 | 0.000% | 0.00000 s |
| rule logic: react/no-find-dom-node | 53 | 0.055% | 0.00113 s |
| rule logic: react/no-is-mounted | 63 | 0.065% | 0.00134 s |
| rule logic: react/no-redundant-should-component-update | 138 | 0.143% | 0.00293 s |
| scanner and parser | 30,632 | 31.702% | 0.65145 s |
| startup and I/O | 1,027 | 1.063% | 0.02184 s |
| startup and I/O: harness | 19 | 0.020% | 0.00040 s |
| remainder: unclassified | 1,770 | 1.832% | 0.03764 s |
| Unsampled wall time: kernel, scheduling, launch and waits | — | — | 0.05512 s |
| Total | 96,624 | 100.000% | 2.11000 s |

The scaled seconds multiply user-CPU sample fractions by the user/wall fraction of the best unprofiled current run (0.973879) and then by 2.11. The remainder includes system CPU and time without a user-mode sample; its internal split cannot be inferred from this perf event. To estimate seconds at today's 2.247552 s instead, multiply the last column by 1.065191. Rows sum to 2.11 before decimal rounding. The unclassified remainder is retained rather than redistributed: 1,741 of its 1,770 samples have an unresolved libc leaf and no classifiable caller; the remaining 29 include truncated array-search/libc/libm stacks. The symbol list is recorded in split.json.

At around 100,000 samples, a 1% share has roughly 1,000 observations (nominal independent-sample 95% error about 0.06 percentage points). Samples within one run are correlated; 20 separate executions and per-PID counts are provided, so this is sampling resolution rather than a claim of that confidence interval for wall-time causality. Perf 6.12.107 reports 32 lost samples (0.033% of recorded plus lost samples) and five lost chunks. Profile load before was `0.96 1.56 1.11`; after was `1.02 1.47 1.11`. Raw loss and load records are preserved in profile.stderr, perf-self.txt.gz and the profile-load files. The attempted 8 MiB perf buffer exceeded the container's perf_event_mlock limit; the supported default buffer was used.

## What the unconditional rule pattern costs

Original registry.ts invokes all ten rules on every node. Each enabled rule selects a node kind before doing useful logic, generally through Context.kind/Context.node and string comparisons; noUnexpectedMultiline caches its kind once and compares it against four kinds. The untimed generated-C counter counts visits and kind-eligible pairs, without modifying the timed binaries:

| Rule | Baseline calls | Kind-eligible calls | Calls avoided |
|---|---:|---:|---:|
| noOctalEscape | 887,803 | 6,228 | 881,575 |
| noUnexpectedMultiline | 887,803 | 90,554 | 797,249 |
| noUnusedPrivateClassMembers | 887,803 | 9 | 887,794 |
| noUselessConstructor | 887,803 | 149 | 887,654 |
| preferTemplate | 887,803 | 6,620 | 881,183 |
| forwardRefUsesRef | 887,803 | 50,379 | 837,424 |
| jsxNoCommentTextnodes | 887,803 | 0 | 887,803 |
| noFindDomNode | 887,803 | 50,379 | 837,424 |
| noIsMounted | 887,803 | 50,379 | 837,424 |
| noRedundantShouldComponentUpdate | 887,803 | 9 | 887,794 |
| Total | 8,878,030 | 254,706 | 8,623,324 (97.131%) |

The original registry, Context, Go comparator and production Go linter are preserved as source fixtures in evidence. Production Go cohere `internal/lint/linter/linter.go` combines listeners in `map[ast.Kind][]func(*ast.Node)` and invokes only `listeners[node.Kind]` before ForEachChild. The pinned batch8 Go comparator instead keeps ten `rule.Listeners` maps and probes all ten integer-key maps per node, invoking only matching listeners. Thus the quoted Go time is the established comparison driver, not a new measurement of production's single combined lookup. It still avoids calls to nonmatching rule bodies and string-kind checks.

DWARF PCs in a rule entry prefix also receive a cross-cutting guard annotation. This counts 25,500 samples (26.391% of user CPU), approximately 0.577657 s at today's best time or 0.542304 s on the 2.11 s scale. It spans ownership, node fetches, string comparisons and selection instructions, so it must not be added to the exclusive table.

| Guard-associated bucket (cross-cutting annotation) | Samples |
|---|---:|
| retains and releases | 7,203 |
| tree walk and node access | 6,437 |
| rule entry and selection | 6,333 |
| string work | 3,255 |
| allocation and freeing | 2,272 |

The annotation includes guards of matching rules, can miss prologue/epilogue work without a matching source line, and does not label an individual invocation as hit or miss. Multiplying it by the 97.13% miss count would not prove an exact miss-only time because different kinds take different paths. The causal evidence is the paired table policy control: it has the same numeric-kind field/cache, table visit and numeric switch as the prototype, but every kind invokes all ten rules. Filtering listeners saves 0.805560 s in that control (33.98%). Its extra machinery makes it slower than the original visitor, so that differential is not charged wholesale to the original 2.247552 s. The deployable prototype's net saving against the original is 0.682340 s; normalized to the earlier 2.11 s, 0.640580 s. This is the measured answer to how much changing the dispatch pattern recovers, not an assertion that every remaining miss-related instruction has been isolated.

The prototype remains 4.34x Go, a 1.204649 s gap. This unit profiles the original driver, not the prototype, so it does not assign that residual gap an exact new split. Original scanner/parser CPU alone accounts for an estimated 0.694 s at today's time, with physical ownership, allocation and string work charged separately. The remaining work is substantial even after unnecessary rule invocations are removed.

## Scratch prototype and validation

[kind-listeners.patch](kind-listeners.patch) contains the complete prototype for the port owner. It assigns an immutable numeric kind once in ParseNode construction, interns names in a process-local Map, and precomputes each kind's ordered numeric rule IDs. Visit fetches the node once, reads its numeric kind, switches only over the listeners for that kind, then iterates children. Rule bodies and their existing entry guards are retained. The numeric-ID switch avoids adding closure calls. Both ClassDeclaration/ClassExpression and Constructor/quoted-constructor-as-MethodDeclaration registrations are preserved. The intern table can grow for unknown kinds, which receive no listeners. Numeric IDs are internal and need not equal TypeScript/Go enum values.

The all-listeners control shares this machinery but registers all ten IDs for every known kind and falls back to all ten for unknown kinds. It deliberately restores the original calls. The patch passes `git apply --check` against the clean driver base. No stage1 port source is committed on this branch; the patch is an artifact under internal/native/performance/lint-cost-split.

Eighteen positive source controls cover all ten rules and every registered kind, including class expressions, quoted constructors, all three template kinds and four unexpected-multiline kinds. Every source is authored as .a. A .a.tsx symlink selects Go's TSX parsing for the JSX case, and the Go-provided JSX span manifest exercises the port's documented adapter contract. All controls produce byte-identical Go/current-native/prototype-native/Node output (10,102 bytes, SHA-256 5bb015dbdd5d3a03b693dfe3d5ddc3c33ed44196c27fc55bc2a80b2e7ff58642). The compiler corpus itself has no JSX eligible calls; this positive control prevents that zero from concealing a missing registration.

Mutants run on disposable generated C: remove prefer-template's StringLiteral registration, caught by template_string (Go 1, mutant 0); force numeric kind zero in both node constructors, caught by private_class (Go 1, mutant 0). Both mutants exit normally with empty stderr, demonstrating semantic loss rather than a compiler failure. A classifier mutant drops the final input sample; the partition total assertion fails. See mutants.json and partition-mutant.log.gz.

Re-green commands against current main, with direct output logs:

```sh
go test ./internal/native -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRuntimeLastIndexOfMatchesNode|TestNativeAgreesWithNode/internal/oracle/testdata/(devirtualize|call_targets|borrow_element_virtual)|TestLibraryMapSetIterator' -count=1 -timeout 30m
```

These passed in 156.297 s, 26.196 s and 4.384 s respectively; direct logs are in evidence. The whole repository gate, sanitizers for the prototype, other architectures and the type-aware bridge were not measured. This unit changes no runtime implementation or emitter.

## Reproduction and evidence

The tools preserve the concrete /workspace paths used in this run. Use the preceding lint-runtime-profile artifact generator to create the baseline batch8 snapshot and Go oracle, set up the scratch driver at its recorded tip with the integer fast paths, then apply kind-listeners.patch there. Snapshot the prototype registry with its parser imports pointing at scratch; the all-listeners snapshot is the same numeric switch with every ID registered for each kind and a full-ID fallback. `regenerate.py` builds original baseline with original nodes, restores prototype nodes for the other two builds, copies the current scratch runtime and invokes `build.py`. `measure.py` rotates four drivers and asserts the finding count. `profile.sh` captures the 20-run task-clock profile; dump MMAP events as above before `split.py`. `parity.py`, `controls.py`, `count.py` and `run-mutants.py` provide the semantic and invocation-count checks. Original node contents and the patch can be recovered from the driver base, so the restore file is not an additional port change.

Build JSON includes the full argv, compiler version, and every C/header SHA-256. Generated main.c files, sampled stacks, DWARF resolutions, exclusive sample assignments, raw timing runs and direct check logs are committed compressed. The multi-gigabyte raw perf.data stays in scratch; its text sample export is committed so bucket assignments can be audited without that binary recording. Root starts from origin/codex/lint-runtime-profile; it incorporates current main only on its own branch. Only codex/lint-cost-split is pushed. No main, area or scratch branch is pushed.
