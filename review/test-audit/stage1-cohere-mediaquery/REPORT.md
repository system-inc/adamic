# u132 mediaquery audit

Starting commit: d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb. Two top-level rows, neither a family. Both sacred within this package and this four-mutant menu. Repo uniqueness awaits central replay. All source restored. Baseline 6.880 seconds, final baseline 6.927 seconds. No skips or timeouts.

## Code under test and oracle, declared before mutation

Port: the five TypeScript files and their native lowering. Oracle: unchanged Go cohere, Node, JavaScript backend and actual postcss-media-query-parser@0.2.3. Gap row: Adamic Lower and native build/run of the closed gaps. Oracle: independent Node execution plus recorded GAPS.md output. Never mutate the Go cohere oracle or gap fixtures.

`port-functions-reached.json` enumerates all 21 observed port functions. `lower-functions-reached.txt` enumerates all 396 lowering functions observed by the clean package coverage run. This compiler inventory is a unit-wide union. Trace instrumentation in scratch copies is supplemental reach evidence, not a mutant: original and traced stdout match exactly. Fixed menu and sites were saved before matrix runs in plan.json. Four production diffs use argument swap or constant changes; W1 is an explicitly permitted witness weakening; P1/P2 are probes only.

## Findings and evidence

See results.json for the complete requested schema, matrix.json for every observed failed subcase, and individual JSON test logs for failing lines. All production runs cover the whole package. M1/M2/M3 kill only the port row; M4 kills only the gap row. No production survivors. Four standalone production diffs compile with the port native build or Go vet. All seven diffs apply cleanly to the starting source. There are no non-menu production mutants.

The port row mixes positive agreement checks with ten built-in witness subcases. W1 makes all ten witness subcases fail while positive checks pass. Its sacred verdict rests only on M1/M2/M3. P1 drops the module executable body: positive checks fail while witness subcases pass, as expected for tests requiring disagreement. P2 makes Lower return nil,nil: the whole run panics, so both rows were rerun individually; both panic. Only P2 judges the gap row's entry and only P1 judges the port entry. The nil panic is a precondition failure, not proof of semantic IR validation.

The baseline exercised 4137 cases: 3984 parsed, 153 refused, 2479 nodes of undecided type. Full output comparison is stronger than exit status or count alone. The library adapter's unclosed-url guard substitutes its own refusal text, so those inputs do not establish agreement with an actual terminating library invocation.

## Brief ambiguities and work costs

- The brief describes stage1 packages as port tests, but the gap row's production subject is the compiler and native pipeline. Mutating gap fixtures would change the expected program, not the implementation. M4 therefore changes lowering.
- One top-level row contains both positive tests and mutant witnesses. A single witness verdict would discard its positive coverage. The report uses sacred from production kills and separate witness evidence, without counting W1 as production uniqueness.
- Empty output deliberately satisfies witness subcases that require disagreement. These are listed as vacuous_subcases while the positive entry probe prevents a whole-row vacuous finding.
- Lower's empty value is nil,nil, and the harness dereferences it before producing a semantic comparison. This required two individual panic reruns; later rows in the initial run remain unknown.
- The fixed menu asks for about three mutants per row, while the stage1 rebuild allowance limits ordinary mutants to four. Used four plus separate probes and witness weakening.
- A switch would alter the port source and require a selector protocol. Used the allowed per-mutant rebuild route instead. Every ID gets an isolated native cache; fresh temporary products prevent stale native answers.
- Test-level PASS duration excludes parallel child completion. Costs use the test binary's package ok line from three individual runs, not the parent PASS line.
- Actual origin/main differs from the brief's historical 8de93800f4. All sites and replay diffs refer to d054e357, and scope comes from list.log.
- Library dependencies were not warm. Installed the optional upstream parser and enabled its check for every run; no skipped row was treated as evidence.
- Whole-package execution fit well below 90 seconds, so no narrowed matrix or unknown package uniqueness was needed. Repo-wide uniqueness remains outside scope.

## Timing and limits

nproc=5. Warm setup skipped: 0 seconds. API npm ci 0.454 seconds; upstream parser install 0.545 seconds. Clean baseline command wall 8.455 seconds, binary 6.880 seconds. Alone medians: gap 0.713 seconds, port 6.673 seconds; all six individual runs passed. Final baseline binary 6.927 seconds.

Production matrix command walls: M1 15.709, M2 8.730, M3 8.997, M4 17.012 seconds. M4 includes Go vet and recompilation. W1 12.417, P1 4.901, P2 10.414 plus individual reruns 1.790 and 2.114 seconds. runs.json preserves exact commands and walls. Compiler timing records preserve every actual clang invocation unchanged; compiler-summary.json sums parallel job durations, which are not elapsed wall time. Initial M1 rebuild also compiles runtime cache misses. Each native product is rebuilt; all observed compile jobs succeed. Timing instrumentation changes no compiler arguments or outputs.

Covered all top-level rows and all default generated inputs plus enabled upstream-library checks. Did not establish repo-wide uniqueness, exhaust all possible inputs or mutations, or claim a survivor without a behavior witness. Evidence-only branch, no production change and no PR.

## Standalone replay table

| ID | Origin site | Change | Observed failed rows |
|---|---|---|---|
| M1 | stage1/cohere/mediaquery/nodes.ts:38 | return new MediaNode(type, after, before, value, sourceIndex, undefined); -> return new MediaNode(type, before, after, value, sourceIndex, undefined); | TestThePortParsesAsGoCohereDoes |
| M2 | stage1/cohere/mediaquery/whitespace.ts:13 | case 0x09: -> case 0x08: | TestThePortParsesAsGoCohereDoes |
| M3 | stage1/cohere/mediaquery/parsers.ts:350 | !rest.startsWith('url') -> !rest.startsWith('URL') | TestThePortParsesAsGoCohereDoes |
| M4 | internal/lower/generic.go:14 | const maximumGenericDepth = 32 -> const maximumGenericDepth = 0 | TestEachGapStandsWhereGapsMdSaysItDoes |
| W1 | stage1/cohere/mediaquery/mediaquery_test.go:323 | Weaken firstDifference to empty string (witness only) | TestThePortParsesAsGoCohereDoes |
| P1 | stage1/cohere/mediaquery/main.ts:121 | Drop executable module body (empty-output probe) | TestThePortParsesAsGoCohereDoes |
| P2 | internal/lower/lower.go:20 | Lower returns nil,nil and removes unused imports (probe) | TestEachGapStandsWhereGapsMdSaysItDoes, TestThePortParsesAsGoCohereDoes |

Production survivors: none.

## Row JSON

```json
[
  {
    "test": "TestEachGapStandsWhereGapsMdSaysItDoes",
    "package": "stage1/cohere/mediaquery",
    "file": "stage1/cohere/mediaquery/gaps_test.go",
    "seconds": 0.713,
    "oracle": "Node independently runs each original gap and confirms handwritten GAPS.md stdout; sanitized native stdout must match.",
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
    "last_proven_fail": "M4: gaps_test.go:83: stage 0 refuses this gap now (/workspace/adamic/stage1/cohere/mediaquery/gaps/1_generic_return.ts:5:13: Adamic 0.1 refuses a generic function instantiated without end (polymorphic recursion); call it with the same type arguments it was called with, or write a function per type): record that in GAPS.md, where it says the program lowers",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestEachGapStandsWhereGapsMdSaysItDoes",
      "TestThePortParsesAsGoCohereDoes"
    ],
    "evidence": "ADAMIC_MEDIA_QUERY_LIBRARY=/tmp/u132/library ADAMIC_BUILD_CACHE_DIR=/tmp/u132/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/mediaquery/ -run . > M4.log 2>&1; gaps_test.go:83: stage 0 refuses this gap now (/workspace/adamic/stage1/cohere/mediaquery/gaps/1_generic_return.ts:5:13: Adamic 0.1 refuses a generic function instantiated without end (polymorphic recursion); call it with the same type arguments it was called with, or write a function per type): record that in GAPS.md, where it says the program lowers"
  },
  {
    "test": "TestThePortParsesAsGoCohereDoes",
    "package": "stage1/cohere/mediaquery",
    "file": "stage1/cohere/mediaquery/mediaquery_test.go",
    "seconds": 6.673,
    "oracle": "Go cohere overlay, Node, JavaScript backend and installed postcss-media-query-parser@0.2.3 compare full output. Nonzero census is self. Library adapter substitutes a handwritten refusal for unclosed url inputs to avoid an upstream hang.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
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
    "last_proven_fail": "M3: mediaquery_test.go:69: Node and Go cohere differ: line 496: \"media-query-list \\\"url(foo.css) screen\\\" @0 before=\\\"\\\" after=\\\"\\\" nodes=1\", Go cohere \"media-query-list \\\"url(foo.css) screen\\\" @0 before=\\\"\\\" after=\\\"\\\" nodes=2\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestEachGapStandsWhereGapsMdSaysItDoes",
      "TestThePortParsesAsGoCohereDoes"
    ],
    "evidence": "ADAMIC_MEDIA_QUERY_LIBRARY=/tmp/u132/library ADAMIC_BUILD_CACHE_DIR=/tmp/u132/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/mediaquery/ -run . > M3.log 2>&1; mediaquery_test.go:69: Node and Go cohere differ: line 496: \"media-query-list \\\"url(foo.css) screen\\\" @0 before=\\\"\\\" after=\\\"\\\" nodes=1\", Go cohere \"media-query-list \\\"url(foo.css) screen\\\" @0 before=\\\"\\\" after=\\\"\\\" nodes=2\"",
    "witness_kills": [
      "W1"
    ],
    "witness_proven": true,
    "vacuous_subcases": [
      "TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped",
      "TestThePortParsesAsGoCohereDoes/catches_U+FEFF_not_whitespace",
      "TestThePortParsesAsGoCohereDoes/catches_trim_as_trimStart",
      "TestThePortParsesAsGoCohereDoes/catches_an_escaped_quote_closing_a_string",
      "TestThePortParsesAsGoCohereDoes/catches_U+2000_not_whitespace",
      "TestThePortParsesAsGoCohereDoes/catches_U+2029_not_whitespace",
      "TestThePortParsesAsGoCohereDoes/catches_a_feature_value's_sourceIndex_not_counting_the_colon",
      "TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped",
      "TestThePortParsesAsGoCohereDoes/catches_the_word_after_an_expression_typed_as_a_media_type",
      "TestThePortParsesAsGoCohereDoes/catches_any_closing_quote_ending_a_string"
    ]
  }
]
```
