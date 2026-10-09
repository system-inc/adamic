Unit u158 starts at 23a19b8b48662dfb6a19baa00b8efbb52ab5275e.
All eight named functions exist, grouped into five rows; none moved, vanished or skipped in narrowed enabled runs.
Whole package timed out; all five narrowed clean baselines and three isolated timing samples passed.
Verdicts: three witnesses, one bounded sacred row, one bounded subsumed row resting on two kills.
M2 survives with Go/clean Node count 7 versus mutant Node/native count 14; both production rows reject empty answers.

```json
[
  {
    "test": "TestWholeMutants family",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_mutants_split_test.go",
    "members": [
      "TestWholeMutants_000",
      "TestWholeMutants_001",
      "TestWholeMutants_002",
      "TestWholeMutantsUnion"
    ],
    "seconds": 4.488,
    "timing_samples": [
      44.035,
      2.393,
      4.4879999999999995
    ],
    "timing_logs": [
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/baseline-1.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-1-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-1-3.log"
    ],
    "nproc": 5,
    "oracle": "Unmodified typescript-go whole-tree bytes, compared with Node and sanitized native control output. Both planted outputs must differ, but any mismatch counts; the mismatch is not constrained to the intended field. Union assignment is a self-written coverage check.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: whole_mutants_test.go:65: Node mutant survived: keyof becomes readonly",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_production_mutants": 4,
    "witness_kills": [
      "W1"
    ],
    "construction_kills": [
      "S1"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWholeMutants_000",
      "TestWholeMutants_001",
      "TestWholeMutants_002",
      "TestWholeMutantsUnion"
    ],
    "evidence": "Apply W1.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeMutants_000|TestWholeMutants_001|TestWholeMutants_002|TestWholeMutantsUnion)$' > W1-witness.log 2>&1; whole_mutants_test.go:65: Node mutant survived: keyof becomes readonly",
    "subsumption_basis_mutants": null,
    "uniqueness_scope": "Bounded production rows only; package and repo uniqueness unknown."
  },
  {
    "test": "TestWholeMutantsRejectsSurvivor",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_mutants_split_test.go",
    "members": [
      "TestWholeMutantsRejectsSurvivor"
    ],
    "seconds": 3.173,
    "timing_samples": [
      2.374,
      4.185,
      3.173
    ],
    "timing_logs": [
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/baseline-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-2-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-2-3.log"
    ],
    "nproc": 5,
    "oracle": "Runs real sanitized checks, then emulates equal native bytes for one mutant. Requires exactly owner TestWholeMutants_002 to fail with the named survived message. W3 disables that leaf guard; parent rejects successful children.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3: whole_mutants_split_test.go:302: planted surviving mutant escaped",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_production_mutants": 4,
    "witness_kills": [
      "W3"
    ],
    "construction_kills": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWholeMutantsRejectsSurvivor"
    ],
    "evidence": "Apply W3.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeMutantsRejectsSurvivor)$' > W3-witness.log 2>&1; whole_mutants_split_test.go:302: planted surviving mutant escaped",
    "subsumption_basis_mutants": null,
    "uniqueness_scope": "Bounded production rows only; package and repo uniqueness unknown."
  },
  {
    "test": "TestWholeCountCheckCatchesMutant",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_mutants_test.go",
    "members": [
      "TestWholeCountCheckCatchesMutant"
    ],
    "seconds": 21.056,
    "timing_samples": [
      24.014,
      21.056,
      20.857
    ],
    "timing_logs": [
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/baseline-3.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-3-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-3-3.log"
    ],
    "nproc": 5,
    "oracle": "Runs typescript-go tree and --count output. Mutated Node/native trees must stay identical and counts must differ. It checks inequality, not a precise mutant count; W2 broadens equality to regard every count as matching.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: whole_mutants_test.go:90: Node counter mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_production_mutants": 4,
    "witness_kills": [
      "W2"
    ],
    "construction_kills": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWholeCountCheckCatchesMutant"
    ],
    "evidence": "Apply W2.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCountCheckCatchesMutant)$' > W2-witness.log 2>&1; whole_mutants_test.go:90: Node counter mutant survived",
    "subsumption_basis_mutants": null,
    "uniqueness_scope": "Bounded production rows only; package and repo uniqueness unknown."
  },
  {
    "test": "TestWholeCompilerAgrees",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_test.go",
    "members": [
      "TestWholeCompilerAgrees"
    ],
    "seconds": 42.573,
    "timing_samples": [
      48.953,
      42.573,
      39.779
    ],
    "timing_logs": [
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/baseline-compiler.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-5-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-5-3.log"
    ],
    "nproc": 5,
    "oracle": "Full serialized whole-tree bytes from unmodified typescript-go over 77 pinned compiler files, compared with Node and sanitized native; execute requires exit success and empty stderr. It does not request --count, so M2 survives. Its two kills also occur in YieldLookaheadAgrees.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: whole_test.go:34: Node: case 0 line 4: port \"1 ExportDeclaration 0 111 0 0 -1 0 0\\t\\t\\t\\t0\", Go \"1 ExportDeclaration 0 109 0 0 -1 0 0\\t\\t\\t\\t0\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestYieldLookaheadAgrees"
    ],
    "mutants_in_matrix": 4,
    "unit_production_mutants": 4,
    "witness_kills": [],
    "construction_kills": [],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 20.681,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWholeCompilerAgrees",
      "TestYieldLookaheadAgrees"
    ],
    "evidence": "Apply M4.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCompilerAgrees|TestYieldLookaheadAgrees)$' > M4-matrix.log 2>&1; whole_test.go:34: Node: case 0 line 4: port \"1 ExportDeclaration 0 111 0 0 -1 0 0\\t\\t\\t\\t0\", Go \"1 ExportDeclaration 0 109 0 0 -1 0 0\\t\\t\\t\\t0\"\nApply P1.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCompilerAgrees|TestYieldLookaheadAgrees)$' > P1-matrix.log 2>&1; whole_test.go:34: Node: case 0 line 2: port \"case 1\", Go \"file\"",
    "subsumption_basis_mutants": 2,
    "uniqueness_scope": "Bounded production rows only; package and repo uniqueness unknown."
  },
  {
    "test": "TestYieldLookaheadAgrees",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/yield_test.go",
    "members": [
      "TestYieldLookaheadAgrees"
    ],
    "seconds": 20.681,
    "timing_samples": [
      23.936,
      20.681,
      20.654
    ],
    "timing_logs": [
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/baseline-4.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-4-2.log",
      "review/test-audit/stage1-typescript-parser-whole_mutants_split/timing-4-3.log"
    ],
    "nproc": 5,
    "oracle": "Full serialized expression-tree bytes from unmodified typescript-go over 17 yield cases, compared with Node and sanitized native; execute requires exit success and empty stderr. M3 is caught by a Node parser panic before tree comparison. It does not request --count.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M4: yield_test.go:45: Node: case 9 line 43: port \"1 Identifier 0 6 0 0 -1 0 0\\t\\tyield\\t\", Go \"1 Identifier 0 5 0 0 -1 0 0\\t\\tyield\\t\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "unit_production_mutants": 4,
    "witness_kills": [],
    "construction_kills": [],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWholeCompilerAgrees",
      "TestYieldLookaheadAgrees"
    ],
    "evidence": "Apply M4.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCompilerAgrees|TestYieldLookaheadAgrees)$' > M4-matrix.log 2>&1; yield_test.go:45: Node: case 9 line 43: port \"1 Identifier 0 6 0 0 -1 0 0\\t\\tyield\\t\", Go \"1 Identifier 0 5 0 0 -1 0 0\\t\\tyield\\t\"\nApply P1.diff; ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCompilerAgrees|TestYieldLookaheadAgrees)$' > P1-matrix.log 2>&1; yield_test.go:45: Node: case 0 line 2: port \"case 1\", Go \"expression\"",
    "subsumption_basis_mutants": null,
    "uniqueness_scope": "Bounded production rows only; package and repo uniqueness unknown."
  }
]
```

| ID | Category | origin/main file:line | Change | Failed grouped rows |
|---|---|---|---|---|
| M1 | production | stage1/typescript/parser/nodes.ts:47 | printed depth to depth + 1 | TestYieldLookaheadAgrees, TestWholeCompilerAgrees |
| M2 | production | stage1/typescript/parser/nodes.ts:58 | countTree initial count 1 to 2 | survivor |
| M3 | production | stage1/typescript/parser/parser.ts:222 | yield same-line flag comparison === 0 to !== 0 | TestYieldLookaheadAgrees |
| M4 | production | stage1/typescript/parser/parser.ts:37 | node end fullStart to start | TestYieldLookaheadAgrees, TestWholeCompilerAgrees |
| W1 | witness | stage1/typescript/parser/parser_test.go:142 | byte equality widened to len(got) >= 0 | TestWholeMutants family |
| W2 | witness | stage1/typescript/parser/whole_mutants_test.go:89 | count equality widened to len(side.count.output) >= 0 | TestWholeCountCheckCatchesMutant |
| W3 | witness | stage1/typescript/parser/whole_mutants_test.go:61 | survival guard excludes planted observation | TestWholeMutantsRejectsSurvivor |
| S1 | construction | stage1/typescript/parser/whole_mutants_split_test.go:255 | leaf index 2 to 1, duplicating assignment | TestWholeMutants family |
| P1 | probe | stage1/typescript/parser/main.ts:5 | per-file port entry run returns 0 before work | TestYieldLookaheadAgrees, TestWholeCompilerAgrees |

M2 survivor witness: for `const x = 1;`, Go and clean Node count 7; mutated Node and the rebuilt sanitized native binary count 14, with successful exits. Both bounded production rows pass because they compare tree output without --count. This is changed count behavior unguarded by those two rows, not a package-wide claim. See survivor-witness.json and the count logs. No other production mutant survives this bounded matrix.

The brief, uncertainty, and costs need these qualifications:

- The brief says five rows but lists eight Test functions. The three numbered WholeMutants leaves plus their Union form one family under the explicit shard-plus-union rule, yielding the five rows stated by the brief. The union checks assignment rather than byte disagreement, so S1 separately breaks its construction. The family's primary verdict is witness, with that construction evidence recorded independently.
- The reference commit 8de93800f4 is older than fetched origin/main. The actual starting SHA is 23a19b8b. All eight names remain in their listed files. list.log is the authoritative census; no name was silently substituted with another test from the file.
- WholeMutants leaves and the count row are built-in mutant witnesses, not ordinary production agreement rows. RejectsSurvivor is a witness of the leaf survival guard. Production mutants were not applied to these witnesses, so neither broken source-replacement preconditions nor altered controls can earn production kills. W1, W2 and W3 each weaken only a guarded comparison; the oracle source and production inputs remain untouched. The permitted witness edit is in test-harness comparison code, including the inline count comparison. No failing assertion was inserted.
- The family's union is a self-written oracle; the other parts actually execute the unmodified Go parser. This mixed oracle must be named rather than calling every family member external-run. The count witness requires a difference and unchanged tree bytes but does not require a particular wrong count. The whole-mutant family accepts any differing tree byte rather than proving the intended field changed. The survivor witness is stronger than an exit-code-only check: it requires one named failed owner and the exact survived diagnostic.
- The package baseline exceeded 90 seconds in unrelated sequential JSX rejection work, before the requested parallel rows ran. There were no failing Test events before the timeout. A test-binary timeout is not a red assertion baseline. Narrowed family and row baselines all passed; each selected run used the same 90-second binary budget. The production matrix contains only WholeCompilerAgrees and YieldLookaheadAgrees from the assigned unit. Static helper callers outside it are saved in callers.json, with their results explicitly unknown.
- The first package run used the default corpus environment because no checkout existed. The compiler-corpus opt-in was then installed and its own clean baseline and every subsequent timing/matrix run enabled ADAMIC_TYPESCRIPT_SOURCE. No requested row remained skipped. The existing cohere TypeScript checkout is a different commit, so it was not substituted for the pinned compiler corpus. A broad filtered clone exceeded the step budget; it was stopped and replaced by a depth-1 fetch of exactly 050880ce59e30b356b686bd3144efe24f875ebc8.
- All four port mutants were frozen from the function inventory before any mutant run. They change three functions across nodes.ts and parser.ts using the allowed menu. M2 changes a count constant to 2, independently of the built-in count witness's 0 mutation. A conservative declaration inventory is saved for recursive parser/scanner descent; it is static reachability evidence, not an instrumented execution coverage profile.
- Existing agreement helpers build native products without a shared result cache. A runtime selector would not make those helpers compile once. Four standalone port mutations were therefore built and validated separately through the port's own sanitized build product. The current recipes hash changed port bytes, and direct buildPort calls rebuild, so compiler-cache isolation was unnecessary: no compiler source was mutated. Additional per-row native builds remain in the measured costs and logs.
- M1 and M4 are tree-byte mismatches. M3 is a logical parser panic on the Node path, caught by execute's success-and-empty-stderr requirement. The Go test binary did not panic or abort; both selected rows have observed pass/fail events. A port panic caught as a test failure is not a cooked test binary. For WholeCompilerAgrees, a Node mismatch prevents its later native comparison from running. YieldLookaheadAgrees constructs both side observations first but stops at the first failing comparison. Native compilation and execution must not be confused with proof that the native comparison fired.
- The production oracle checks serialized bytes, not all possible parser properties. Neither assigned production row requests --count. M2 demonstrates that limitation with changed output from both backends and an independently executed Go count. The separate count witness has its own weakened-check proof. Central replay can determine which other parser-package rows guard the count; no deletion or package-wide unguarded claim is supported here.
- The per-file run function in main.ts is the port's processing entry: it reads, parses and prints each input. P1 returns its zero value before that work. Top-level manifest case banners remain, but there is no parse result. Both production rows fail P1 and are non-vacuous. Production probes do not apply to the three witness rows, whose vacuous value is null rather than guessed. Their W mutations are recorded separately, never as probe or production kills.
- Yield's M3 kill is unique only within the bounded production matrix. It supports bounded sacred, pending central package/repo replay. WholeCompilerAgrees is subsumed by YieldLookaheadAgrees on exactly M1 and M4, with a faster 20.681-second subsumer median. This two-kill hint does not establish that the 77-file corpus is dispensable. M2 survives both, and none of the four mutations measures every semantic field or input in that corpus.
- Family timing runs select all four members together as one row; each other row runs alone. Three separate count=1 invocations are used, and medians come from the test binary's package Elapsed event rather than shell wall time. The family's first run includes a cold four-product build, while repeats reuse its hash-addressed setup. Timing samples are preserved so that warm/cold variation is reviewable. Tool installation was warm; Node dependencies were still installed before baseline. The first npm ci was not separately timed, and its timed repeat took about 0.715 seconds.
- All nine standalone diffs apply to the starting source. Four production diffs and P1 compile through the real sanitized port builder. W1-W3 and S1 pass go vet. Source restoration is checked with an empty package diff. Raw logs are force-added because the repository ignores *.log; they are evidence, not claims inferred from source inspection.

Building and running costs:
```json
{
  "bootstrap_seconds": 0,
  "npm_ci_seconds": 0.7150784140012547,
  "nproc": 5,
  "whole_package_baseline_test_seconds": 90.037,
  "port_rebuilds": [
    {
      "id": "M1",
      "test_binary_seconds": 20.516,
      "wall_seconds": 25.04315835200032,
      "log": "review/test-audit/stage1-typescript-parser-whole_mutants_split/M1-build.log"
    },
    {
      "id": "M2",
      "test_binary_seconds": 21.015,
      "wall_seconds": 24.692451225000696,
      "log": "review/test-audit/stage1-typescript-parser-whole_mutants_split/M2-build.log"
    },
    {
      "id": "M3",
      "test_binary_seconds": 21.475,
      "wall_seconds": 24.937956148000012,
      "log": "review/test-audit/stage1-typescript-parser-whole_mutants_split/M3-build.log"
    },
    {
      "id": "M4",
      "test_binary_seconds": 26.692,
      "wall_seconds": 32.898534965999715,
      "log": "review/test-audit/stage1-typescript-parser-whole_mutants_split/M4-build.log"
    },
    {
      "id": "P1",
      "test_binary_seconds": 13.01,
      "wall_seconds": 15.634114423999563,
      "log": "review/test-audit/stage1-typescript-parser-whole_mutants_split/P1-build.log"
    }
  ],
  "timing_repeat_command_seconds": 211.89586897399568,
  "matrix_and_build_command_seconds": 350.20098966600017,
  "session": "about 22 minutes from 13:30 UTC; native compile phases and additional uncached per-row builds are included in the logs",
  "corpus_install": "full-history clone exceeded 90 seconds and was stopped after observed 116-second process elapsed; narrowed depth-1 fetch succeeded. Exact install wall was not separately timed."
}
```

No mutant, probe or witness run cooked; there were no unknown rows from a Go panic. Setup bootstrap was skipped and nproc was 5. Production semantic coverage beyond four mutations, parser callers outside this slice, and package/repo-wide uniqueness were not covered. The complete original package baseline was not repeated per mutant after its timeout. No main push or pull request.
