Unit u148 at a7448d73cd17f16362b6cbc5c5c111080da64e43.
Two listed top-level rows, no moved or missing tests.
Four frozen production mutants; 0 survivors.
Two empty-answer probes and one weakened-comparison witness.
Evidence branch test-audit/stage1-cohere-values; source restored.

```json
[
  {
    "test": "TestThePortParsesAsGoCohereDoes",
    "package": "stage1/cohere/values",
    "file": "stage1/cohere/values/values_test.go",
    "seconds": 37.733,
    "oracle": "Go cohere and postcss-values-parser 2.0.1 run on Node; full serialized trees and refusal text, plus sanitizer/leak success",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M1",
      "M2",
      "M3"
    ],
    "last_proven_fail": "M3: values_test.go:68: native and Go cohere differ: line 948: \"    comment inline=false raws={after:\\\"\\\",before:\\\" \\\"} source={end:{column:13,line:1},start:{column:5,line:1}} sourceIndex=4 value=\\\" one */\\\"\", Go cohere \"    comment inline=false raws={after:\\\"\\\",before:\\\" \\\"} source={end:{column:13,line:1},start:{column:5,line:1}} sourceIndex=4 value=\\\" one \\\"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_VALUES_LIBRARY=/tmp/u148-library ADAMIC_BUILD_CACHE_DIR=/tmp/u148/cache/replay-M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run . > review/test-audit/stage1-cohere-values/M3.log 2>&1; values_test.go:68: native and Go cohere differ: line 948: \"    comment inline=false raws={after:\\\"\\\",before:\\\" \\\"} source={end:{column:13,line:1},start:{column:5,line:1}} sourceIndex=4 value=\\\" one */\\\"\", Go cohere \"    comment inline=false raws={after:\\\"\\\",before:\\\" \\\"} source={end:{column:13,line:1},start:{column:5,line:1}} sourceIndex=4 value=\\\" one \\\"\""
  },
  {
    "test": "TestEachGapStandsWhereGapsMdSaysItDoes",
    "package": "stage1/cohere/values",
    "file": "stage1/cohere/values/gaps_test.go",
    "seconds": 0.887,
    "oracle": "Node executes all five fixtures and validates handwritten stdout; native output matches that stdout; one exact NotYet.What label is self-written",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: gaps_test.go:80: natively: exit 0, stdout \"1\\n\", stderr \"\"; Node prints \"3\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_VALUES_LIBRARY=/tmp/u148-library ADAMIC_BUILD_CACHE_DIR=/tmp/u148/cache/replay-M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run . > review/test-audit/stage1-cohere-values/M4.log 2>&1; gaps_test.go:80: natively: exit 0, stdout \"1\\n\", stderr \"\"; Node prints \"3\\n\""
  }
]
```

| ID | origin/main location | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/values/tokenize.ts:201 | `let offset = -1; -> let offset = 0;` | TestThePortParsesAsGoCohereDoes |
| M2 | stage1/cohere/values/nodes.ts:173 | `sourceIndex, unit, false, -1); -> sourceIndex, '', false, -1);` | TestThePortParsesAsGoCohereDoes |
| M3 | stage1/cohere/values/parser.ts:51 | `pair === '/*' || pair === '*/' -> pair === '/*' || pair === '**'` | TestThePortParsesAsGoCohereDoes |
| M4 | internal/native/emit_strings.go:68 | `"adamic_string_index_of_from(%s, %s, %s)", value, arguments[0], arguments[1] -> "adamic_string_index_of_from(%s, %s, %s)", value, arguments[0], "0.0"` | TestEachGapStandsWhereGapsMdSaysItDoes |

Survivors: none.

Code and oracle were named before mutants; plan.json freezes the code-derived menu. functions-static.txt inventories declarations, port-reached.json records actual V8 reach, coverage-functions.log records measured compiler/native function coverage from the gap row. Static declarations are not claimed as proof of reach. Count-only driver functions are outside the tested entry mode.

Witness: W1 replaces firstDifference with an empty answer. It is a harness-only exception, excluded from production kills. The port row combines positive agreement and 18 built-in mutation witness subcases; W1.log gives exact outcomes. No separate top-level witness verdict is assigned to that mixed row.

Probes: P1 returns an empty parsed tree via new Parser(loose, []).parse(), the valid empty-input answer. P2 makes Lower return an empty ir.Program and removes the now-unused imports. Only P1 judges port vacuity and P2 judges gap vacuity. P2 effects on preparation for the port do not count as port probe kills. Neither probe supports sacred.

Brief problems and time costs:

- P1 returns an empty tree, yet all 18 built-in witness subcases pass because empty answers already disagree with Go cohere. The positive subcase fails, so the whole row is not vacuous. W1 separately demonstrates the weakened comparison is detected.
- Stage1 rebuild exception caps mutants at four, overriding the approximate three per row target; three port changes plus one compiler change were tested. Each TS diff has an actual sanitized native build; M4 passes go vet and its generated native products compile during the matrix.
- The unit checks two different products: a port and the compiler. Mutating only the port cannot assess the gap row. M4 targets native stringCall instead of changing fixtures or oracle code.
- The port top-level row combines a positive agreement check and mutation witnesses. A single verdict cannot describe both roles; its sacred verdict rests only on production answer differences, while W1 separately demonstrates the witness check.
- Valid typed empty values need representation rather than nil pointers. P1 is an empty tree and P2 is an empty IR program; neither is a production mutant.
- Audit execution error: /usr/bin/time does not exist. The first wrapped npm ci did not run. npm ci was corrected using Python timing and passed. The restored clean package baseline then passed; qualifying timings and production matrices were repeated after it. Initial logs are retained separately and do not decide verdicts.
- One read tool call lost its exec-server transport and was retried. No code or test result was obtained from that failed call.
- Full serialized outputs provide stronger evidence than exit codes; the gap refusal subcase additionally requires a specific error type and exact self-written label. It does not derive that label from Node.
- The gap fixture comments still describe closed features as refused; GAPS.md and the actual gap table mark their current status. Scope came from go test -list, not those comments.

Timing: {"nproc": 5, "setup": "warm env, cloud/setup.sh skipped", "npm": {"wall": 1.8971379050017276, "exit": 0}, "builds": {"M1-build": 2.3161529270000756, "M2-build": 2.0378082400020503, "M3-build": 2.0401783640008944, "P1-build": 1.8494287309986248}, "run_wall": 726.923239020005, "all_recorded_wall": 737.0424054200084}. Raw runs.json records every measured command; test binary seconds, rather than command wall time, decide row medians. Initial baseline binary seconds were 31.355.

Not covered: other packages and repo-wide uniqueness, real CSS corpora, other random seeds, count-only driver mode, key insertion order, non-UTF-8 inputs and lone surrogates, alternate operating systems, every compiler path. Four mutations are bounded evidence, not a proof that every possible break is caught. Standalone diffs apply to the pinned starting commit; replay validation logs are preserved. No product source changes remain.
