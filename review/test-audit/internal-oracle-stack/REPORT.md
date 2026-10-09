Unit u069: two rows present, neither moved or vanished.
Base: 6c60da091afddc9c2fe88b3a1067845b6dc79cb3; nproc 5; toolchain warm.
Whole package timed out at 90.062 s; clean bounded runs passed.
Four production mutants: all killed; both rows rejected both empty-answer probes.
Bounded verdicts: long arguments subsumed; small stacks sacred.

```json
[
  {
    "test": "TestLongArgumentsLeaveTheStackItsLimit",
    "package": "internal/oracle",
    "file": "internal/oracle/stack_test.go:17",
    "seconds": 0.107,
    "oracle": "Original TypeScript source executed by Node; full stdout, stderr and exit status compared to sanitized and release native products.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M03",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04: stack_test.go:75: program: stderr differs",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSmallStacksStillPanic"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": 0.103,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLongArgumentsLeaveTheStackItsLimit",
      "TestSmallStacksStillPanic",
      "TestNativeAgreesWithNode"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestLongArgumentsLeaveTheStackItsLimit|TestSmallStacksStillPanic|TestNativeAgreesWithNode)$/^(arguments|arguments_alone|environment|1024_KiB|512_KiB|256_KiB|internal)$/^oracle$/^testdata$/^stack_overflow[.]a$; ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u069/cache/M04; stack_test.go:75: program: stderr differs"
  },
  {
    "test": "TestSmallStacksStillPanic",
    "package": "internal/oracle",
    "file": "internal/oracle/stack_test.go:89",
    "seconds": 0.103,
    "oracle": "Node validates the exact start output, panic text and exit 70 at 1024 KiB. At 512/256 KiB the expected same contract is self-written; Node is deliberately not run. M04 proves stderr is checked even when exit 70 and stdout remain correct.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M04: stack_test.go:129: program at 512 KiB: stderr differs: exit 70, stdout \"start\\n\", stderr \"adamic: panic: RangeError: stack limit exceeded\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLongArgumentsLeaveTheStackItsLimit",
      "TestSmallStacksStillPanic",
      "TestNativeAgreesWithNode"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestLongArgumentsLeaveTheStackItsLimit|TestSmallStacksStillPanic|TestNativeAgreesWithNode)$/^(arguments|arguments_alone|environment|1024_KiB|512_KiB|256_KiB|internal)$/^oracle$/^testdata$/^stack_overflow[.]a$; ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u069/cache/M04; stack_test.go:129: program at 512 KiB: stderr differs: exit 70, stdout \"start\\n\", stderr \"adamic: panic: RangeError: stack limit exceeded\\n\""
  }
]
```

| ID | origin/main location | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/runtime/stack.c:56 | `uintptr_t arguments = size / 4 > ARGUMENTS_FLOOR ? size / 4 : ARGUMENTS_FLOOR;` to `uintptr_t arguments = 0;` | TestLongArgumentsLeaveTheStackItsLimit, TestSmallStacksStillPanic |
| M02 | internal/native/runtime/stack.c:57 | `size / 8 < MARGIN ? size / 8 : MARGIN` to `size / 8 > MARGIN ? size / 8 : MARGIN` | TestSmallStacksStillPanic |
| M03 | internal/native/emit_functions.go:55 | `e.line("ADAMIC_CHECK_STACK();")` to `` | TestLongArgumentsLeaveTheStackItsLimit, TestNativeAgreesWithNode, TestSmallStacksStillPanic |
| M04 | internal/native/runtime/stack.c:66 | `"RangeError: Maximum call stack size exceeded"` to `"RangeError: stack limit exceeded"` | TestLongArgumentsLeaveTheStackItsLimit, TestNativeAgreesWithNode, TestSmallStacksStillPanic |

Survivors: none in the bounded matrix. No equivalent candidates claimed.

Scope and limitations:
- The brief cites 8de93800f4; fetched origin/main was 6c60da091afddc9c2fe88b3a1067845b6dc79cb3. Both requested names exist there in stack_test.go at lines 17 and 89. No grouping applies: bodies assert distinct contracts.
- Whole-package baseline exceeded its 90-second binary budget without an observed assertion failure. Narrowed via stack fixture references and native emission/runtime call paths to both unit rows plus only TestNativeAgreesWithNode/internal/oracle/testdata/stack_overflow.a. Other members of that corpus row and all other package rows remain unknown. Unique kills mean unique only in this bounded set. No repo-wide testing performed.
- matrix_rows contains the top-level corpus row because counts are over rows. Only its named fixture ran; its full corpus family was not measured or judged. The long row is subsumed by the small row on three caught production mutants, a small-set hint, not deletion advice.
- CODE UNDER TEST: native.C and native functionBody stack-check emission; native runtime find_stack_limit, ADAMIC_CHECK_STACK, adamic_stack_overflow and adamic_panic. ORACLE: original source run by Node, plus the self-written small-stack contract below 1024 KiB. Coverage records 314 reached Go production functions, listed in reached-functions.txt; runtime-reach.txt records the C chain. Lowering and unused JavaScript generation prepare the fixture, so were not mutated or empty-probed.
- Ordinary three-run timing uses the suite's warm gate cache. Those binary seconds measure the row as normally invoked, not uncached native execution. Mutation and probe runs explicitly disable the gate cache and give each mutant its own build cache.
- The arguments-alone subcase clears its environment, so an environment mutation switch would silently stop selecting C mutations. The scratch switch instead reads /tmp/u069/mutant. Its clean uncached baseline passed. Switch scaffolding is absent from every standalone diff and supports no verdict by itself.
- Four mutants were fixed before outcome inspection. The brief's approximate three-per-row goal would suggest six, but its explicit maximum-four rule for native rebuilds limits this unit to four. They span reservation, margin, emission and panic text.
- Native.C empty probe fails during native build, which demonstrates rejection of no generated answer, not a production semantic kill. Runtime constructor empty probe leaves a zero limit and fails native execution. Both are separate probe_kills. No positive subcase passed either probe.
- Native runtime diffs compile with the builder's release and sanitizer clang flags. Go emitter/probe diffs pass go vet ./internal/native/. Exact commands and durations are in mutant-plan.json and probe-plan.json. Linux native behavior was exercised; Darwin/WASI alternate stack-limit branches were not separately covered by this slice. WASI SDK opt-in was enabled for the baseline. No skip event was observed before that baseline timed out; tests not reached are unknown.
- No outside authority is claimed for the below-1MiB panic expectation. The 1MiB Node execution checks the same expected value. Assertions compare complete stdout/stderr/exit, not just panic exit status: M04 still exits 70 with start output, and both rows fail on its changed stderr.
- The final restored-source bounded run passed in 1.362 binary seconds; runtime/emitter production sources are restored. Replaying standalone diffs is necessary to settle package and repo uniqueness.

Costs:
Toolchain setup skipped (warm env); nproc=5. npm ci installed 3 packages in 0.438 s. Whole package binary: 90.062 s timeout. Clean two-row uncached binary: 1.313 s (3.399 s command wall). Clean three-row uncached binary: 1.293 s. Scratch Go compilation: 8.118 s. Solo timings: long 0.133/0.107/0.103 s, median 0.107; small 0.100/0.110/0.103 s, median 0.103. Runtime validation and Go vet totals are recorded per diff.
Native rebuilding is included in each mutation command, with distinct caches. Separate native compiler wall time was not isolated, so the following report is an inclusive rebuild-and-run measurement, not a compiler-only claim.
M01: command wall 4.003 s; binary including native build 1.780 s.
M02: command wall 3.251 s; binary including native build 1.236 s.
M03: command wall 3.746 s; binary including native build 1.576 s.
M04: command wall 3.157 s; binary including native build 1.238 s.
Probes command wall: P01/TestLongArgumentsLeaveTheStackItsLimit 2.174 s, P01/TestSmallStacksStillPanic 2.336 s, P02/TestLongArgumentsLeaveTheStackItsLimit 2.736 s, P02/TestSmallStacksStillPanic 2.760 s.
