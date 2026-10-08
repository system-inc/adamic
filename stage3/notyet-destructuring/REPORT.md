# Destructuring lowering evidence

Base: area/compiler b410340dc8f889b5799c3bc519117c63def3aa24. Replay merged at 9a1f14c5. Checked non-null merged at c41c0e06 (merge ebfbfe43).

## Computed field names (6 roots)

Lowered literal string concatenations and safe integral const enum keys. Runtime enum keys retain evaluation before the value. Computed __proto__ remains an own field in both backends. Dynamic keys and numeric concatenation remain explicit NotYet.

Replays use the exact adapted source manifest from stage3-latent-full 9d534d3a31814f1a192a528e701f6c2ea7c910bc (81/81 matching hashes). Evidence: computed-replay.json. Parser reaches a generic function expression; scanner and sys report no next finding; diagnostics reaches a census statement panic on ComputedPropertyName; visitor reaches its next structural-method stop. These are measurement results on a checker-rejected compiler corpus, not executable compiler builds.

Validation: three fixtures run through TestNativeAgreesWithNode with ADAMIC_GATE_UNCACHED=1 (source Node, native ASan/UBSan/leaks, emitted JavaScript Node), passed after non-null merge in 2.689s. Lower boundary tests passed 0.196s; JavaScript package has no tests. TestCountsAreRecorded -update-counts passed 209.148s and added three rows.

Mutants: string-key, enum-key, binding-key, runtime-key-evaluation, own-proto each caught by stdout differs. singleton-call and numeric-concatenation each caught by the lower boundary test. Run internal/oracle/testdata/run-notyet-computed-mutants.py with the toolchain environment sourced. Build failures were rejected and are not counted as mutant kills. Logs: /tmp/destructuring-computed-mutants/.

Toolchain setup: node .409s; Go .393s; clang 2.087s; markdown 4.299s; submodules 62.639s; Go build 2339.790s; deferred test binaries 2340.484s; cache warm 2340.491s; done 2340.731s. nproc 5, quota 4 CPUs. Setup log /tmp/destructuring-setup.log.

## Arrays, binding slots, defaults and nested patterns

Lowered array declarations with holes, copied rest, live-length stepping and sticky exhaustion; nested object/tuple bindings; defaults evaluated once and only for undefined; optional boolean and boxed union slots. The tuple slot fixture includes exactly the watch-like family string | number | boolean | readonly string[] | file-like object | undefined, and checks reference identity after declaration and assignment. No runtime C helper changed.

Four new fixtures passed Node/native sanitizers/emitted JavaScript; the final binding plus library_string_raw regression selection passed 2.820s, and the extended reference-slot fixture passed 11.988s. Lower boundary/prototype checks passed 4.604s. Counts refresh passed 132.676s, adding four rows and updating library_string_raw's retains/releases by four (allocations/frees unchanged); its oracle regression also passed.

Nine scratch-overlay mutants were run independently: evaluate-once, holes, rest-copy, sticky-exhaustion, lazy-undefined-default, nested-field, object-union-box, tuple-union-box, tuple-field. Every one was caught by stdout differs, and none was credited for a Go/clang build failure. Runner: internal/oracle/testdata/run-notyet-binding-mutants.py; logs: /tmp/destructuring-binding-mutants/.

post-non-null-replay.json and post-bindings-replay.json record all 27 roots. Confirmed signature removals: array kind 3/5 (the other two already stop at unchecked casts); held-slot kind 2/4 (builder/watchUtilities already stop at Path); non-plain binding kind 3/4 (checker 50576 already stops at __String); tuple-union kind 2/2. Each confirmed site reaches its next named stop. Earlier blockers are not claimed as fixed. Nullable reference defaults whose slot cannot distinguish null from undefined remain explicit NotYet; object rest, tuple rest, general dynamic keys and erased structural iterator origins remain outside these additions.

## Remaining collection lessons and final scope

Lowered intrinsic array/string length bindings (UTF-16 string units); Map construction from a proven custom iterator of two-slot tuples, inserting before next and leaving next failures unclosed; optional array/string inputs only for library Map/Set constructors; fixed tuple-member assignment targets with receiver-before-element-read and live alias behavior. No runtime C helpers were added. Supporting object-literal checks keep normalized duplicate fields and native-inexpressible NUL field names explicit rather than silently corrupting shapes.

The four collection fixtures passed the Node/native sanitizer/emitted JavaScript oracle. Eight independent overlay mutants were caught by stdout differences: array-length, utf16-length, map-interleaving, map-next-close, absent-array, absent-string, tuple-target-index, tuple-target-before-read. The duplicate-field mutant and storage-field mutant were caught by the lower boundary assertions. Together with the earlier seven and nine mutants, this unit has 26 caught mutants: 22 behavioral stdout differences and four boundary failures. No build failure counts as a kill. Runners and full logs are under internal/oracle/testdata/run-notyet-*-mutants.py (plus the singular duplicate/storage runners) and /tmp/destructuring-{computed,binding,collection,duplicate,storage}-mutant*.

Current main efe9f404 was merged as ec63a775. The checked-non-null dependency c41c0e06 was merged earlier at the user's explicit request, before the later standing rule against unlanded dependency merges. Future unlanded dependencies will be named rather than merged. The completed unit is pushed once after this final validation.

### Per-kind results

| Implementation commit | Kind | Roots | Confirmed assigned stops removed | Remaining evidence |
| --- | --- | ---: | ---: | --- |
| a4d733d6 | computed field name | 6 | 6 | Scanner/sys have no next findings; parser generic function; diagnostics census ComputedPropertyName panic; visitor structural method |
| 2fc6f0ee | non-tuple array destructuring | 5 | 3 | checker 46259/46271 already blocked at unchecked casts |
| 2fc6f0ee | field representation differs | 4 | 2 | builder 601/watchUtilities 578 already blocked at Path |
| 2fc6f0ee | non-plain destructured name | 4 | 3 | checker 50576 already blocked at __String |
| final collection commit | destructuring a value | 2 | 1 | checker 11491 clears; tsbuildPublic 2295 did not reproduce its target |
| 2fc6f0ee | watch-like tuple union slot | 2 | 2 | Next stops recorded in final-replay.json |
| final collection commit | Map from non-array pairs | 2 | 1 | commandLineParser 141 reaches for...of over object; builder 2406 already blocked at Path/map overload |
| final collection commit | destructuring a string | 1 | 1 | scanner 1333 has no findings |
| final collection commit | iterating a value | 1 | 0 | programDiagnostics 219 never reproduced the target, earlier structural method/optional call/cache reads remain |

Thus 19/27 assigned signatures are confirmed removed. Eight roots are earlier-blocked or unreproduced, not claimed fixed. All nine reduced compiler lessons are implemented; none is refused for an Adamic-design ruling. Existing unchecked-cast and nominal-type refusals remain outside the unit. General dynamic field names, object/tuple rest, ambiguous nullable reference defaults and erased iterable origins remain explicit unsupported boundaries.

Replay snapshots distinguish observation from inference: final-replay.json is the final 27-site measurement; no absence of a finding proves the checker-rejected adapted compiler executes correctly. Fixtures provide executable cross-backend evidence for the reduced rules.

### Final commands

All commands source /workspace/adamic-tools/env.sh first; output is written directly to logs, never piped. No whole-package test run or full gate was invoked.

- `go test ./internal/lower ./internal/javascript -run '^(TestComputed.*|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet)$' -count=1`: lower passed 2.007s; JavaScript has no test files. Log /tmp/destructuring-final-lower.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(notyet_|tuple_values.a|tuples_kept.a|library_string_raw.a|library_for_in.a|class_features_private.a)' -count=1 -timeout 10m`: passed 33.767s, 11 new fixtures and five existing regressions, source Node/native sanitizers/generated JavaScript. Log /tmp/destructuring-final-oracle-corrected.log.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: passed 190.792s; log /tmp/destructuring-final-counts-storage.log.
- `python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/destructuring-overlay`, then `go build -buildvcs=false -overlay=/tmp/destructuring-overlay/overlay.json -o /tmp/destructuring-replay ./stage3/census/latent/replay/worker`, then replay each assigned where/reason with `-project /tmp/destructuring-adapted/src/tsc/tsc.ts -where <site> -kind NotYet -reason <reason>`: all 27 measurements recorded in final-replay.json. Log /tmp/destructuring-final-replay-storage.log.
- `python3 internal/oracle/testdata/run-notyet-computed-mutants.py`, `run-notyet-binding-mutants.py`, `run-notyet-collection-mutants.py`, `run-notyet-duplicate-mutant.py`, `run-notyet-storage-mutant.py`: 7 + 9 + 8 + 1 + 1 caught, with exact kill names above. Each invocation's stdout/stderr was redirected to its /tmp/destructuring-*-mutant*.log.

Outside the assigned functions, changes are limited to the shared destructure holding hook, tupleLiteral/objectLiteral/methodName, the computed-field and destructuring helper files, and the JavaScript object-literal own-__proto__ emission. No IR, native backend or runtime source change was necessary. Each implementation commit names its supporting source files.

The final counts refresh caught an overbroad field-name guard: library_for_in.a and class_features_private.a deliberately cover public quoted # names. The guard was corrected to reject only embedded NUL, and the computed-field fixture now checks reflection of a public computed # name. Both existing regressions were added to the final oracle selection. This was a local validation failure corrected before the completed-unit push.

The added public # fixture was reduced further after TypeScript inferred an index signature for a concatenated-key object: the final fixture checks only construction and reflection of its computed own key. The index-signature binding remains an explicit unsupported boundary. Initial local reductions failed type-checking / field-proof lowering; no failure was claimed as a passing validation.

A standalone final-source counted build also passed: `go run ./cmd/adamic build internal/oracle/testdata/notyet_computed_fields.a -o /tmp/destructuring-final-computed-count --count`, then execute with an 8 MiB stack. Its row is exactly the refreshed table: allocations 34, frees 34, retains 15, releases 41, peak 18, regions 0. Logs /tmp/destructuring-final-computed-count-{build,run,stdout}.log.
