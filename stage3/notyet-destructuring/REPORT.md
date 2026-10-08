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
