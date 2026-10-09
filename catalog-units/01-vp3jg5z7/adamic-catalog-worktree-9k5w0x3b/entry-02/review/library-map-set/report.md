# Map and Set library slice

Branch `codex/library-map-set`, from origin/main `fe3b9f2`; runner merged first at
`bb766d42ba415def680c8223bc7825e228141a9a`.
Node v24.19.0; test262 pinned at `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`.

## Measurements

Every newly passing test agrees with Node. These are observed runner counts, not estimates
of library coverage. The full directory and refusal measurements are in before.json and after.json.

| Before | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Map | 0 | 0 | 83 | 0 | 121 | 204 |
| built-ins/Set | 0 | 0 | 224 | 0 | 159 | 383 |

| After | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Map | 0 | 0 | 83 | 0 | 121 | 204 |
| built-ins/Set | 13 | 0 | 211 | 0 | 159 | 383 |

The largest remaining refusal is TS2345 argument type mismatch: 18 Map plus 43 Set
cases, 61 total. `var` accounts for another 39. The unchanged Map table does not prove
its new behavior: the five typed differential fixtures provide that evidence.

New test262 passes: result-order.js for difference, intersection, symmetricDifference,
and union; compares-itself.js, compares-same-sets.js and compares-sets.js for each of
isSubsetOf, isSupersetOf and isDisjointFrom.

## Implementation and coverage

New lowering lives in internal/lower/library_map_set.go, with small dispatch hooks.
New runtime lives in internal/native/runtime/map_set.c. Stored collection iterators have
an explicit IR node, a native next closure, and a matching JavaScript-backend wrapper.

Constructors accept Adamic-supported arrays, strings (Set, Unicode code points), Maps,
Sets, and their entries/keys/values iterators, including partially consumed stored iterators,
Set-of-tuples Map input, and null/undefined empty input. Iterators stream for-of, expose
next/done/value, remain exhausted after done, and follow collection deletion, insertion
and clear. Set entries produce [element, element]. Strong source captures are visible to
both freshness and cycle analysis.

forEach accepts an arrow callback with thisArg, evaluates context in call order, and keeps
existing live traversal: added entries are visited and deleted entries skipped. Dynamic-this callbacks and void-valued contexts remain explicitly refused.

SameValueZero covers NaN, normalized -0, reference identity, and the packed number |
undefined word without collapsing missing keys into present NaN. Size is checked across
updates, deletes and clear.

All seven ES2025 Set operations accept concrete Sets/ReadonlySets with matching element
representations. Intersection uses the smaller operand's order (ties use the receiver);
union, difference and symmetricDifference preserve spec order. Map.groupBy accepts typed
arrays, strings, Maps, Sets and stored collection iterators, walks inputs live, supplies the
index, propagates callback throws, and refuses unsound element widening.

Still refused: arbitrary user iterable/Symbol.iterator protocols, custom set-like operands,
dynamic thisArg binding, void-valued contexts, mixed element representations for Set operations, multiword
collection slots, iterator properties other than next, and next arguments. Existing Adamic
style restrictions (var, implicit callback types, detached methods) are unchanged.

## Toolchain

`bash cloud/setup.sh` succeeded. Its timing lines: go ready 0s; clang ready 1s; node ready
1s; submodules ready 1s; build cache warm 21s; done 21s. `nproc` printed 5 (CPU quota 4).
Go 1.27.1, clang 20.1.8 with working ASan/UBSan, Node 24.19.0. Environment sourced from
/workspace/adamic-tools/env.sh. See setup.log.

## Commands and observed results

All test commands redirect stdout and stderr to logs.

- Before and after: `go run ./cmd/adamic-test262 -json -test262 /workspace/test262 -work /tmp/adamic-map-set-{before,final-guard} built-ins/Map built-ins/Set`; exit 0, tables above.
- `go test -count=1 -timeout 10m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(library_map_set|collections)'`; PASS, 33.372s. Source Node, sanitized native, release native, JavaScript backend and leak checks agree.
- `go test -v -count=1 -parallel 4 -timeout 10m ./internal/oracle -run TestLibraryMapSetMutants`; PASS, 60.993s. All thirteen clean-exit behavior mutants below caught by Node stdout comparison.
- `go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts`; PASS, 232.126s. Five fixture rows added; three existing collection rows updated for iterator allocations.
- `gofmt -l cmd internal`; empty output. `go vet ./...`; empty output.

## Mutants

Each behavior mutant compiled with -Werror and exited 0 with no sanitizer finding; only
comparison with Node stdout killed it. These persist in TestLibraryMapSetMutants.

| Family | Mutation | Caught by |
|---|---|---|
| union | Swap operands | Node stdout |
| intersection | Swap operands, changing small/tie order | Node stdout |
| difference | Swap operands | Node stdout |
| symmetricDifference | Swap operands, changing result order | Node stdout |
| isSubsetOf | Swap operands | Node stdout |
| isSupersetOf | Swap operands | Node stdout |
| isDisjointFrom | Make final success false | Node stdout |
| SameValueZero packed keys | Use ordinary-number constructor for optional keys | Node stdout |
| size | Report used slots rather than live count | Node stdout |
| iterators | Start the cursor at slot 1 | Node stdout |
| constructors | Consume one entry before draining an iterator | Node stdout |
| forEach mutation | Stop at initial used-slot count | Node stdout |
| Map.groupBy | Start callback indices at 1 | Node stdout |

Final focused checks after the capture guard: `go test -count=1 -timeout=5m
./internal/lower` passed in 8.711s. `go test -count=1 -parallel=4 -timeout=10m
./internal/oracle -run 'TestFreshWriteProbesStayRefused|TestLibraryMapSetMutants|TestNativeAgreesWithNode/internal/oracle/testdata/(library_map_set|collections)'`
passed in 228.827s. Its slash filter selects the differential fixtures, so the refusal probes
and mutants were also run separately with a root-only filter; their result is recorded below.

An additional ownership mutant disabled libraryIteratorCaptures. The cycle probe
fresh_refused/map_iterator.a was then accepted: source Node and native both exited 0 with
stdout `1`, while LeakSanitizer found **608 bytes in eight leaked allocations**. Restoring
the guard refuses the marked write. This mutation is caught by the ownership/leak check,
not by Node behavior comparison. Both Map and Set cycle probes are kept in fresh_refused.
The full evidence is in cycle-mutant.log.

`go test -count=1 -parallel=4 -timeout=10m ./internal/oracle -run
'TestFreshWriteProbesStayRefused|TestLibraryMapSetMutants'` passed in **63.344s**
on the final code. All cycle refusal probes, including the new Map and Set iterator probes,
and all thirteen behavior mutants ran. See final-mutants.log.

The full gate command was `go test -count=1 -timeout 30m ./...`.
Its native normalization points sweep reached its independent five-minute subprocess limit:
both processes were killed after 1,053,494 of 1,114,112 expected lines. The reported final
partial-line difference is not evidence of a completed behavior disagreement. The source of
that limit is context.WithTimeout in internal/native/case_test.go. A focused sweep rerun and
the remaining full-gate results are recorded below. No normalization source was changed.

The complete normalization sweep rerun, `go test -count=1 -parallel=1 -timeout=15m
./internal/native -run TestNormalizeMatchesNode`, passed in **105.681s**, comparing every
expected line with Node. See normalize-rerun.log. The original broad gate retains its
failed status; the successful rerun is reported separately.

A final parenthesis guard adjustment follows the same SkipParentheses rule as builtin
method dispatch. `go test -v -count=1 -timeout=5m ./internal/oracle -run
'TestFreshWriteProbesStayRefused/.*iterator'` passed all three probes in **0.320s**
(Map, Set, and `(map.keys)()`); lowering tests then passed in **6.787s**.
Neither adjustment changes allocation counts or the successful Set test262 results.
See final-cycles.log and final-lower.log.

Stored iterators increase allocation cost: collections.a moves from 300 to 363 allocations,
with 363 frees and the same peak of 101. undefined_keys.a moves from 98 to 122, peak 19
to 24; gaps.a moves from 89 to 97, peak remains 41. No runtime benchmark was claimed.

Full gate finished with **exit 1** solely for the normalization subprocess timeout above.
The oracle package passed in **1450.594s**, including differential fixtures, mutants,
allocation counts and leak checks. Every other package passed: runner 24.430s, flow
233.748s, fresh 42.147s, fuzz 16.033s, load 2.125s, lower 13.701s, and all six stage1/cohere
packages (183.043s–345.823s). See gate.log. The gate started before the final iterator
capture guard; final focused lowering, ownership and differential checks are reported
separately above. Final formatting and vet both exit 0 with empty logs; git diff --check
is clean. The complete normalization rerun passed; the full gate itself is not claimed green.

Implementation commit: `9d0729a873fc31c6a02af064b5759865831e0077`. The requested branch was pushed successfully; no pull request was opened. Validation logs are intentionally tracked in this review directory despite the repository-wide *.log ignore rule.
