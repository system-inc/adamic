u118 audited at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; nproc=5.
Clean baseline passed in 60.763s; five top-level tests, no families or skips.
Verdicts: one sacred, two subsumed, one setup-check, one cannot-judge.
Four production mutants: three caught, one confirmed non-equivalent survivor; three construction breaks and two witness checks also caught.
Evidence: test-audit/stage1-cohere-lint-regex, review/test-audit/stage1-cohere-lint-regex/.

```json
[
  {
    "test": "TestDynamicPatternGap",
    "package": "stage1/cohere/lint/regex",
    "file": "stage1/cohere/lint/regex/regex_test.go:114",
    "seconds": 0.15,
    "oracle": "Node executes the dynamic fixture and must print true; native lowering must reject with a self-written diagnostic substring. The substring proves the named refusal, not successful native regex execution.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: regex_test.go:128: expected named native gap, got /workspace/adamic/stage1/cohere/lint/regex/testdata/dynamic_gap.a:3:27: stage 0 can't lower RegExp pattern unavailable yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestShapeFixtures"
    ],
    "mutants_in_matrix": 7,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 45.74,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestDynamicPatternGap",
      "TestFixedPatterns",
      "TestInventoryMatchesPinnedSource",
      "TestOptionDialectGap",
      "TestShapeFixtures"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/regex/ -run . > M1.log 2>&1; regex_test.go:128: expected named native gap, got /workspace/adamic/stage1/cohere/lint/regex/testdata/dynamic_gap.a:3:27: stage 0 can't lower RegExp pattern unavailable yet",
    "subsumption_mutants": 1
  },
  {
    "test": "TestFixedPatterns",
    "package": "stage1/cohere/lint/regex",
    "file": "stage1/cohere/lint/regex/regex_test.go:67",
    "seconds": 8.07,
    "oracle": "Go standard regexp, run under pinned cohere module, compares full match spans and UTF-16 text against source Node, emitted JavaScript, sanitized native. M2 shows missing ignore-case input coverage.",
    "oracle_kind": "external-run",
    "kills": [
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M3",
      "M4"
    ],
    "last_proven_fail": "M4: regex_test.go:87: backend 0 differs: nexus/shouting.go:124",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 7,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestDynamicPatternGap",
      "TestFixedPatterns",
      "TestInventoryMatchesPinnedSource",
      "TestOptionDialectGap",
      "TestShapeFixtures"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/regex/ -run . > M4.log 2>&1; regex_test.go:87: backend 0 differs: nexus/shouting.go:124",
    "witness_evidence": "W1: regex_test.go:109: translation mutant escaped backend 0",
    "witness_verdict": "witness"
  },
  {
    "test": "TestInventoryMatchesPinnedSource",
    "package": "stage1/cohere/lint/regex",
    "file": "stage1/cohere/lint/regex/regex_test.go:147",
    "seconds": 0.36,
    "oracle": "Own Go AST census is compared byte-for-byte against sites.json, a repository snapshot; no outside authority checked.",
    "oracle_kind": "self",
    "kills": [
      "S1",
      "S2",
      "S3"
    ],
    "unique_kills": [
      "S1",
      "S2",
      "S3"
    ],
    "last_proven_fail": "S3: regex_test.go:157: regexp census drifted from pinned cohere AST; regenerate and review the table",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 7,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestDynamicPatternGap",
      "TestFixedPatterns",
      "TestInventoryMatchesPinnedSource",
      "TestOptionDialectGap",
      "TestShapeFixtures"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/regex/ -run . > S3.log 2>&1; regex_test.go:157: regexp census drifted from pinned cohere AST; regenerate and review the table"
  },
  {
    "test": "TestOptionDialectGap",
    "package": "stage1/cohere/lint/regex",
    "file": "stage1/cohere/lint/regex/regex_test.go:133",
    "seconds": 0.13,
    "oracle": "Standalone Go regexp and Node RegExp observations compare against hand-written strings. Neither script imports options.a or calls optionPattern; no legitimate port mutant can reach this row without changing its oracles.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 7,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestDynamicPatternGap",
      "TestFixedPatterns",
      "TestInventoryMatchesPinnedSource",
      "TestOptionDialectGap",
      "TestShapeFixtures"
    ],
    "evidence": "go test -list . and baseline.log; regex_test.go:133-144 and standalone oracle scripts contain no call to optionPattern. No mutation failure claimed."
  },
  {
    "test": "TestShapeFixtures",
    "package": "stage1/cohere/lint/regex",
    "file": "stage1/cohere/lint/regex/shapes_test.go:14",
    "seconds": 45.74,
    "oracle": "Fresh Go regexp matches validate recorded spans; source Node and emitted JS compare full bytes. All 107 runtime-string native fixtures are named refusals. Self-written coverage count and diagnostic labels also checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: shapes_test.go:45: per-shape fixture comparison failed: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDynamicPatternGap"
    ],
    "mutants_in_matrix": 7,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.15,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestDynamicPatternGap",
      "TestFixedPatterns",
      "TestInventoryMatchesPinnedSource",
      "TestOptionDialectGap",
      "TestShapeFixtures"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/regex/ -run . > M1.log 2>&1; shapes_test.go:45: per-shape fixture comparison failed: exit status 1",
    "witness_evidence": "W2: shapes_test.go:54: mutant escaped or failed outside comparison; area emitted JavaScript fixture mismatch",
    "witness_verdict": "witness",
    "probe_limit": "P2 tests Lower only. The 214 static/runtime port entry programs were not individually probed; whole-row vacuity remains null.",
    "subsumption_mutants": 1
  }
]
```

| ID | Origin file:line | Change | Failed rows |
| --- | --- | --- | --- |
| M2 | stage1/cohere/lint/regex/patterns.a:6 | /^no default(?![\s\S])/gui -> /^no default(?![\s\S])/gu |  |
| M3 | stage1/cohere/lint/regex/patterns.a:26 | [a-z]{2,3} -> [a-z]{2,2} | FixedPatterns |
| M4 | stage1/cohere/lint/regex/patterns.a:28 | [AEIOUY] -> [aeiouy] | FixedPatterns |
| M1 | internal/lower/regexp.go:39 | "RegExp with a nonconstant pattern" -> "RegExp pattern unavailable" | ShapeFixtures, DynamicPatternGap |
| S1 | stage1/cohere/lint/regex/testdata/inventory.go:102 | sel.Sel.Name != "Compile" -> sel.Sel.Name != "CompileBROKEN" | InventoryMatchesPinnedSource |
| S2 | stage1/cohere/lint/regex/testdata/inventory.go:68 | depth > 20 -> depth > 0 | InventoryMatchesPinnedSource |
| S3 | stage1/cohere/lint/regex/testdata/inventory.go:108 | Line: fs.Position(c.Pos()).Line, -> Line: fs.Position(c.Pos()).Line + 1, | InventoryMatchesPinnedSource |
| W1 | stage1/cohere/lint/regex/regex_test.go:86 | bytes.Equal(expected, actual) -> true | FixedPatterns |
| W2 | stage1/cohere/lint/regex/testdata/shapes/gate.go:153 | equal(f, "source Node", want, actual) -> equal(f, "source Node", actual, actual) | ShapeFixtures |
| P1 | stage1/cohere/lint/regex/patterns.a:2 | export function fixedPatterns(): RegExp[] { -> export function fixedPatterns(): RegExp[] { return []; | FixedPatterns |
| P2 | internal/lower/lower.go:21 | files := program.Files() -> if program != nil { return nil, nil } 	files := program.Files() | FixedPatterns |
| P3 | stage1/cohere/lint/regex/testdata/inventory.go:28 | Return at main entry with no output | InventoryMatchesPinnedSource (alone) |

M2 survived: fixedPatterns()[2].test("NO DEFAULT") is true before and false after. See M2-witness.json and both Node logs. No other production survivors. S/W/P are construction checks, witness checks and probes, not production survivors.

The brief leaves mixed tests unclear. FixedPatterns and ShapeFixtures perform real parity checks and then run built-in mutants. They fail when the implementation is wrong, so I judged their production coverage as ordinary rows and reported their witness subcases separately. W1 and W2 do not enter production kill sets. Both witness subcases failed when their comparisons were weakened. The forced-native-transition witness was not separately weakened.

Subsumption here rests on exactly one diagnostic mutant, M1. ShapeFixtures also runs 107 fixture checks that DynamicPatternGap does not run. Their mutual subsumption is a hint about this small matrix, not a deletion recommendation or a claim that their full purposes are interchangeable.

The port-only rule needs an exception for the dynamic-gap row: its actual assertion is about compiler lowering. M1 therefore changes lower.regexConstant, not Go regexp or Node. Its expected rejection phrase is self-written. Matching that phrase can accept another diagnostic containing the same text; the test does not prove native execution of a dynamic regex.

OptionDialectGap does not exercise the claimed port function. Its Go and Node programs are the observed dialect authorities, while its expected strings are hand-written. Mutating those programs would violate the oracle rule. I report cannot-judge rather than fabricate a port kill or call the row untrue from unrelated mutations.

The fixed-pattern oracle is Go standard regexp executed within the pinned cohere module, not a call to cohere's lint matcher. Its byte comparison covers spans and UTF-16 text, but the M2 survivor proves that the corpus misses at least one ignore-case distinction. The independent input NO DEFAULT changes pattern index 2 from true to false. This is unguarded behavior, not an equivalent candidate.

The shape gate reports all 107 runtime-string native cases as awaiting compiler support. No tests skipped, but those cases prove named refusal and transition enforcement, not successful native runtime-string parity. Source Node and constant-string emitted JS do run. The optional approved library runtime compiler was not supplied, and I did not invent an installable replacement.

Empty-answer probing is incomplete for the shape port entries: Lower was probed, but its 107 static and 107 runtime entry programs were not individually emptied. Whole-row vacuity therefore remains null. OptionDialectGap has no reached Adamic entry to probe. P1, P2 and P3 caught the working-answer versus empty-answer distinction for FixedPatterns, DynamicPatternGap and the census respectively. P2 panicked in the initial package run; all five rows were then run alone, and only those observed results enter the probe matrix.

The initial W1 edit also replaced the inventory comparison and failed compilation because it left unused variables. It was rejected before execution, restored, narrowed to the two fixed-pattern comparisons, and rerun. The final W1 diff compiles and contains only that corrected scope. This cost one failed compile check. replay.log preserves it; no verdict uses it.

I used separate source rebuilds rather than a selector insertion, keeping three port mutants within the four-mutant rebuild cap. Every production/construction/witness run covered the complete five-row package and fit its 90-second test-binary budget. P3 ran only its owning census row because probe kills do not establish uniqueness. Repo-wide uniqueness remains for central replay. All standalone diffs apply to the starting commit, and successful sanitized native builds validate the .a production diffs.

The frozen inventory names port functions and relevant regex lowering handlers. Supplemental coverage lists every observed in-process lowering function reached by FixedPatterns and DynamicPatternGap. It does not instrument the separate shapes-gate compiler subprocess, so an exhaustive compiler-helper reachability inventory for that subprocess is not claimed.

Warm tools required no setup; stage3/api npm ci took 0.892s. nproc=5. Baseline test binary: 60.763s. Three clean row samples and medians are in timings.json. Compile/vet checks in runs.json total 4.261s; recorded whole-package mutation/probe command wall times total 607.195s, including Go compilation and native builds. Panic reruns, P3, coverage and Node witnesses are additional and preserved in their logs. Native stopwatch rebuilds: M2 1.485s, M3 1.427s, M4 1.331s, plus about 0.12s load/lower each. No run cooked. Session began around 12:27 UTC and evidence preparation finished around 12:52 UTC, about 25 minutes. Production source is restored. No other packages' tests, repo-wide replay, uncalled port functions, optional external compiler, individual shape-entry probes, or exhaustive shape-subprocess helper coverage were audited.
