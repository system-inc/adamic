Added catchable stack, padding and concatenation RangeErrors, and corrected the frozen-object refusal for fresh Error initialization.
Implementation commit: eb9fcd2; the following audit commit records the completed evidence on codex/error-classes-counts.
The uncached affected gate, Node/native/JavaScript comparisons, count regeneration, scoped Cohere and vet pass; all 268 prior fixture rows are exact.
Four compiled native mutants, seven source-overlay mutants and the previous seven precision mutants are caught.
Not covered: the complete repository gate, benchmarks, every string-producing runtime primitive, or the unavailable original with() attachment.

Continuing from `a0b57c2`, on the same branch. No checkout from main, no PR, no edits to narrowed.go or the four prohibited compiler files.

| Probe | Classification | Choice and evidence |
| --- | --- | --- |
| `d96d304_try_stack.a` | **fixed-now** | Catchable nominal RangeError. Recursive functions reachable from protected try bodies or catches with finally, and every callee they can reach, get an ordinary stack-margin throw instead of the fatal native entry check. Node, sanitized native, release native and generated JavaScript print `caught the overflow\nafter\n`, exit 0. |
| `d96d304_try_repeat.a` | **fixed-before** | `be5a509` introduced repeat's count and result-length guard. `154e646` retained that guard unless a constant product fits. Two units times 268435456 is 536870912, above the 536870888-unit limit. All backends print `caught the long repeat\nafter\n`, exit 0. This unit additionally fixes libraryFailure's fallback: a finite constant count alone is not sufficient proof. |
| `d96d304_try_pad.a` | **fixed-now** | Catchable nominal RangeError before allocating padding. A nonempty fill and truncated target greater than the limit fail. The fixture preserves the exact padStart probe and additionally checks padEnd with Infinity and the empty-fill exception. All backends print `caught the long pad\nafter\ncaught the long padEnd\naa\n`, exit 0. |
| `d96d304_try_finally_concat.a` | **fixed-now** | Catchable nominal RangeError before joining operands. Both operands evaluate once, left to right; their UTF-16 lengths are summed before allocation. All backends print `finally ran\n`, then `adamic: panic: RangeError: Invalid string length\n` on stderr, exit 70. The throw unwinds through finally before escaping. |
| `57f2d04_with_frozen.a` | **fixed-now**, reconstructed probe | The integration attachment was not present in this workspace. The permanent representative probe freezes an unrelated object, tests successful with(), then catches its out-of-range RangeError and verifies the original array. The old Lower refusal was observed. Built-in Error initialization writes only into a fresh allocation or fresh subclass receiver; marking that internal initializer LibraryGuarded excludes its writes from the frozen-object refusal. User writes remain subject to the existing boundary. All backends print `3|2\ncaught RangeError\n1|1|2\n`, exit 0. |

The repeat attribution is from inspecting `git show be5a509:internal/lower/error_library.go` and the unchanged guard on this branch, then running the exact probe. This unit did not rebuild every historical commit. No claim that the original integration with() attachment was run is made.

The stack path uses the existing pending-exception cleanup ABI, not longjmp or a signal handler. Its failing entry constructs the RangeError literal directly, without entering another checked constructor on an exhausted stack. Parameter ownership is established before the IR throw and released by the normal function exit path. Guarding recursive entries alone would be insufficient: an acyclic helper or a hidden temporal-dead-zone error constructor could cross the margin first. The call graph therefore includes closure and interface targets, array callbacks and checked-global error construction, and it is rebuilt after introducing string helpers. All descendants use the catchable entry check. The existing runtime margin and 8 MiB counted stack remain unchanged.

Node's own JavaScript stack failure is a built-in RangeError, rather than an object emitted by Adamic's constructor. Catch now attaches its correct nominal identity while preserving the actual object. Specific built-in identities precede the Error fallback, and objects already carrying Adamic identity are unchanged. The live-heap recursion probe requires `instanceof RangeError` in both backends, not merely an unbound catch.

Two supplemental fixtures independently check stack cleanup and a stack error thrown from a catch that still owes its own finally. The first keeps heap strings live across recursive calls and must finish with no LeakSanitizer report. The second prints `finally ran` before the uncaught overflow, exit 70. Their depth-dependent counts are deliberately not recorded, for the same reason the pre-existing stack_overflow fixture is excluded from counts.

String guards are added to protected bodies and functions they reach. A catch is protected when it has a finally to run; an unprotected finally is guarded only by an enclosing protected body. This avoids charging an unrelated final string append with a new throwing wrapper. UTF-8 constant byte lengths conservatively bound UTF-16 units. Number and boolean conversion lengths, closed-program stores, direct-call arguments and return values prove small joins; unknown producers, indirect parameters and cycles keep the full runtime limit. Spread, Object.assign, runtime object construction and unknown object producers prevent a false field bound. Uncalled constructors cannot store into live objects and are excluded from the closed call graph facts. Future generated stack-error fields and hidden ready-check constructor arguments are included before proving bounds. Class dispatch is refreshed after introducing error classes.

Constants bypass padding guards only when truncation fits the limit, or a literal fill is empty. NaN and negative targets preserve the native/Node identity-return behavior. The existing repeat guard handles negative and infinite counts and oversized results; libraryFailure now separately refuses an unguarded repeat whose result bound does not fit, with `repeat length or count` as its reason. The precision assertions hold these decisions independently of the counts table.

The four original limit fixtures live under `internal/oracle/testdata/catchability-limits/` and are explicitly registered in the ordinary three-backend oracle. They are outside flow's executable root glob: its JavaScript tracing calls change where a stack overflow happens and can leave its trace frames unbalanced, while printing huge strings would also change the workload. Small existing recursive exception fixtures remain in the full flow gate. Both supplemental stack probes have explicit source Node, generated JavaScript and sanitized native comparisons; the successful heap-string probe additionally runs LeakSanitizer. No sanitizer or ordinary oracle comparison is disabled.

| Mutant | Check that catches it | Observation |
| --- | --- | --- |
| Stack entry falls back to the fatal prologue | Exact stack fixture, TestRuntimeRangeMutants | Native builds; Node exit 0, mutant exit 70. |
| Repeat's result guard removed | Exact repeat fixture, same test | Native builds; Node exit 0, mutant exit 70. |
| Padding guard removed | Pad fixture, same test | Native builds; Node exit 0, mutant exit 70. |
| Concatenation guard removed | Finally/concat fixture, same test | Native builds; both exit 70, but mutant loses `finally ran`. |
| Fresh Error initializer marker removed | with()/freeze fixture, source overlay | Go builds; Lower reintroduces the precise potentially-frozen-object refusal. |
| Concat exemption ignores the bound | Finally/concat fixture, source overlay | Go and native build; stdout comparison loses finally. |
| Unprotected finally always guarded | TestProtectedStringGuardPrecision | Go builds; the bounded finished function incorrectly becomes MayThrow. |
| Unknown object field bound ignored | TestStringLengthBoundsIncludeUnknownObjects | Go builds; the unknown object's bound incorrectly becomes the small observed literal. |
| Any finite repeat constant treated as safe | TestRepeatRefusalChecksResultLength | Go builds; the oversized unguarded repeat loses its named refusal. |
| Catch with finally omitted from protected roots | TestRuntimeStackCatchFinally | Go and native build; native loses the finally output. |
| JS native-error nominal identity omitted | TestRuntimeStackCleanup | Go and native build; generated JavaScript reports the wrong error class. |

Reproduce the seven source overlays with `python3 cloud/error-runtime-mutants.py`. The runner accepts only each intended assertion, never a Go or C build failure. The four compiled IR mutants run in the ordinary oracle gate. The previous `cloud/error-classes-counts-mutants.py` was also rerun: all seven mutants were caught, including every-call MayThrow, every-function MayThrow, ignored counter writes, ignored unknown field stores, invalid presence after mutation, unproved readiness and all-calls-pure use-after-free. The affected gate retains the existing exception under-approximation mutants.

All **268 previous fixture rows** at `a0b57c2` remain byte-for-byte exact in all six columns. Regeneration adds five rows, for **273 fixture rows**. Earlier narrative totals counted the table header: the preceding catchability report's three references to 267 prior fixtures are corrected to 266, and that unit added two rows to reach 268. The original 63-row audit and its 34 restored / 29 legitimate-error classification are unchanged.

| Added fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| try_stack | 1 | 1 | 3 | 2 | 1 | 0 |
| try_repeat | 1 | 1 | 6 | 5 | 1 | 0 |
| try_pad | 3 | 3 | 11 | 11 | 2 | 0 |
| try_finally_concat | 4 | 3 | 5 | 8 | 4 | 0 |
| with_frozen | 10 | 10 | 11 | 21 | 5 | 0 |

Setup was rerun with `bash cloud/setup.sh > /tmp/adamic-runtime-setup.log 2>&1`, then the toolchain environment was sourced from `/workspace/adamic-tools/env.sh`. `nproc` printed 5. Go 1.27.1, clang 20.1.8 and Node 24.19.0:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (2s)
setup: submodules ready (2s)
setup: build cache warm (304s)
setup: done in 304s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Tests use `TMPDIR=/tmp/adamic-gate`, mode 1777; output is written directly to logs. Intermediate runs caught the fresh-Error refusal, missing handling of a function's own try, the new primitive missing from the freshness analysis and the unsuitability of dynamically tracing the overflow fixture. These were fixed; the complete affected gate was rerun. No failure is waived.

| Command | Result | Log |
| --- | --- | --- |
| `gofmt -l cmd internal` | exit 0, no output | `/tmp/adamic-runtime-gofmt-final.log` |
| `go vet ./...` | exit 0, no output | `/tmp/adamic-runtime-vet-final.log` |
| `git diff --check` | exit 0, no output | `/tmp/adamic-runtime-diffcheck-final.log` |
| Focused oracle command below | exit 0, all five registered fixtures, four native mutants and both supplemental stack tests pass | `/tmp/adamic-runtime-unit-last.log` |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | exit 0, 45.142s; 268 prior rows exact, five added | `/tmp/adamic-runtime-counts-final.log` |
| `python3 cloud/error-runtime-mutants.py` | seven compilable mutants caught | `/tmp/adamic-runtime-source-mutants-final2.log`, individual logs under `/tmp/adamic-runtime-source-mutants/` |
| `python3 cloud/error-classes-counts-mutants.py` | seven precision mutants caught | `/tmp/adamic-runtime-counts-mutants-final.log` |
| Scoped Cohere command below | exit 0, 276 rules, seven checked, seven Adamic-ready | `/tmp/adamic-runtime-cohere-verified.log` |
| Full affected gate below | exit 0; all affected packages pass | `/tmp/adamic-runtime-gate-complete.log` |

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(catchability-limits|57f2d04_)|TestRuntimeRangeMutants|TestRuntimeStackCleanup|TestRuntimeStackCatchFinally' -count=1 -v > /tmp/adamic-runtime-unit-last.log 2>&1
/tmp/adamic-catchability-cohere --directory /tmp/adamic-runtime-cohere-fixtures --no-fix d96d304_try_stack.ts d96d304_stack_cleanup.ts d96d304_stack_catch_finally.ts d96d304_try_repeat.ts d96d304_try_pad.ts d96d304_try_finally_concat.ts 57f2d04_with_frozen.ts > /tmp/adamic-runtime-cohere-verified.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql > /tmp/adamic-runtime-gate-complete.log 2>&1
```

The pinned Cohere binary, original compiler options and original prelude were reused from the preceding unit. Its seven `.ts` mirrors match the final `.a` fixtures byte for byte; the prelude was loaded but not selected for linting. Final output: `276 rules, 7 checked, 100% Adamic-ready (7 of 7), only 7 of 8 files`. This is a scoped check, not repository-wide lint.

Not covered: `go test ./...`, repository-wide Cohere, benchmarks, universal equality of Node/native overflow depth, allocation failure, C-runtime internal recursion or every other string-producing primitive's size failure. The named pre-existing refusal boundaries remain in docs/memory.md. Other string-producing primitives are not made catchable by this unit; the four requested runtime points are. The original with() attachment was unavailable, so the exact reported behavior is held by a reconstructed representative fixture. No PR is opened.

Final affected gate output, exit 0:

```text
? ir: no test files
ok lower 50.385s
? javascript: no test files
ok flow 228.069s
ok fresh 106.148s
ok native 349.528s
ok oracle 337.392s
ok graphql 150.134s
```

Implementation: `eb9fcd27935477df802562d61003c99181f1e5d4` (`Catch string length and stack failures through cleanup`). The following report commit records this evidence and corrects the previous report's header-inclusive row totals. Only documentation changed after the final passing gate. Both commits are pushed on `codex/error-classes-counts`.
