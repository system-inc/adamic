u080 audited at cf79ecec3723604428ab91ebcb283400d05a1548; all 389 requested names exist in their stated file.  
Two rows: one Setup check and one 388-member family; production matrix bounded to 33 fixed shards plus Setup.  
Three of four production mutants caught; M02 survives this slice with a demonstrated native output change.  
Setup median 4.562s; bounded family median 12.797s; full-family median unavailable after three 90s timeouts; nproc=5.  
Source restored; all standalone diffs compile; evidence under review/test-audit/stage1-cohere-css-top_level_shards/.

```json
[
  {
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/top_level_shards_test.go",
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 4,
    "bounded": true,
    "matrix_rows": [
      "TestThePortParsesAsGoCohereDoes_Setup",
      "TestThePortParsesAsGoCohereDoes family"
    ],
    "test": "TestThePortParsesAsGoCohereDoes_Setup",
    "seconds": 4.562,
    "oracle": "Self-written shard enumeration and union invariants: 24,594 live cases x four variants = 98,376 unique IDs. Port products are built, not semantically compared by this row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S01: css_test.go:154: union count 73782, live enumeration 98376",
    "verdict": "setup-check",
    "probe_kills": [
      "P02"
    ],
    "vacuous": false,
    "construction_kills": [
      "S01"
    ],
    "evidence": "ADAMIC_CSS_FIXTURES=/tmp/u080-css-fixtures ADAMIC_CSS_LIBRARY=/tmp/u080-css-library ADAMIC_AUDIT_PROBE= timeout 120 go test -overlay=/tmp/u080-S01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '^TestThePortParsesAsGoCohereDoes_Setup$'; css_test.go:154: union count 73782, live enumeration 98376"
  },
  {
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/top_level_shards_test.go",
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 4,
    "bounded": true,
    "matrix_rows": [
      "TestThePortParsesAsGoCohereDoes_Setup",
      "TestThePortParsesAsGoCohereDoes family"
    ],
    "test": "TestThePortParsesAsGoCohereDoes family",
    "seconds": null,
    "bounded_seconds": 12.797,
    "full_timing_lower_bound_seconds": 90,
    "oracle": "Exact native, Node port-source and JS-backend output compared with live Go cohere answers. Pinned PostCSS 8.5.16 / postcss-scss 4.0.9 checked with self-written gap allowances. Built-in mutant witnesses accept any difference, including empty output (P01: 198 logged acceptances); disabling comparison W01 makes all three designated witnesses fail.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M01",
      "M03",
      "M04"
    ],
    "unique_kills_scope": "Only within the bounded two-row matrix; package uniqueness unknown.",
    "last_proven_fail": "M04: css_test.go:199: native: shard-372: line 18, byte 36: \"{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\"],\\\"kind\\\":\\\"root\\\",\\\"nodes\\\":[{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\",\\\"selector\\\"],\\\"kind\\\":\\\"rule\\\",\\\"nodes\\\":[],\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\", Go cohere \"{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\"],\\\"nodes\\\":[{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\",\\\"selector\\\"],\\\"nodes\\\":[],\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\",\\\"between\\\":\\\" \\\"},\\\"selector\\\"\"",
    "verdict": "slow-worthy",
    "probe_kills": [
      "P01"
    ],
    "vacuous": false,
    "vacuous_subcases": [
      "P01: built-in mutant groups 0, 1 and 2 accept empty output as a caught mutant in all 33 selected shards. These groups are not Go subtest functions."
    ],
    "members": "All 388 wrappers _000 through _387, listed in scope.json",
    "matrix_members_file": "scope.json: bounded_members",
    "evidence": "pathlib.Path('/tmp/u080-mutant').write_text('M04'); ADAMIC_CSS_FIXTURES=/tmp/u080-css-fixtures ADAMIC_CSS_LIBRARY=/tmp/u080-css-library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '^(TestThePortParsesAsGoCohereDoes_000|TestThePortParsesAsGoCohereDoes_012|TestThePortParsesAsGoCohereDoes_024|TestThePortParsesAsGoCohereDoes_036|TestThePortParsesAsGoCohereDoes_048|TestThePortParsesAsGoCohereDoes_060|TestThePortParsesAsGoCohereDoes_072|TestThePortParsesAsGoCohereDoes_084|TestThePortParsesAsGoCohereDoes_096|TestThePortParsesAsGoCohereDoes_108|TestThePortParsesAsGoCohereDoes_120|TestThePortParsesAsGoCohereDoes_132|TestThePortParsesAsGoCohereDoes_144|TestThePortParsesAsGoCohereDoes_156|TestThePortParsesAsGoCohereDoes_168|TestThePortParsesAsGoCohereDoes_180|TestThePortParsesAsGoCohereDoes_192|TestThePortParsesAsGoCohereDoes_204|TestThePortParsesAsGoCohereDoes_216|TestThePortParsesAsGoCohereDoes_228|TestThePortParsesAsGoCohereDoes_240|TestThePortParsesAsGoCohereDoes_252|TestThePortParsesAsGoCohereDoes_264|TestThePortParsesAsGoCohereDoes_276|TestThePortParsesAsGoCohereDoes_288|TestThePortParsesAsGoCohereDoes_300|TestThePortParsesAsGoCohereDoes_312|TestThePortParsesAsGoCohereDoes_324|TestThePortParsesAsGoCohereDoes_336|TestThePortParsesAsGoCohereDoes_348|TestThePortParsesAsGoCohereDoes_360|TestThePortParsesAsGoCohereDoes_372|TestThePortParsesAsGoCohereDoes_384|TestThePortParsesAsGoCohereDoes_Setup)$'; css_test.go:199: native: shard-372: line 18, byte 36: \"{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\"],\\\"kind\\\":\\\"root\\\",\\\"nodes\\\":[{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\",\\\"selector\\\"],\\\"kind\\\":\\\"rule\\\",\\\"nodes\\\":[],\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\", Go cohere \"{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\"],\\\"nodes\\\":[{\\\"keys\\\":[\\\"raws\\\",\\\"nodes\\\",\\\"source\\\",\\\"selector\\\"],\\\"nodes\\\":[],\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\",\\\"between\\\":\\\" \\\"},\\\"selector\\\"\""
  }
]
```

| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | stage1/cohere/css/parser.ts:82 | `increment: number = 1` to `increment: number = 0` | TestThePortParsesAsGoCohereDoes family (all 33 selected members) |
| M02 | stage1/cohere/css/tokenize.ts:23 | `code === 13 || code === 12;` to `code === 13 || code === 11;` | none, bounded survivor |
| M03 | stage1/cohere/css/input.ts:71 | `new Position(offset, low + 1,` to `new Position(offset, low + 2,` | TestThePortParsesAsGoCohereDoes family (all 33 selected members) |
| M04 | stage1/cohere/css/nodes.ts:79 | `fields.set('type', quote(node.type));` to `fields.set('kind', quote(node.type));` | TestThePortParsesAsGoCohereDoes family (all 33 selected members) |
| S01 | stage1/cohere/css/parser_shards_test.go:77 | Loop bound `len(mutants)` to `len(mutants)-1`; allowed construction break | Setup |
| W01 | stage1/cohere/css/parser_shards_test.go:197 | Conditional early empty comparison; allowed witness weakening | Family witness owners _227, _052, _124; supplemental, excluded from production kills |
| P01 | stage1/cohere/css/main.ts:43 | Return at command entry without output | Family; probe only |
| P02 | stage1/cohere/css/parser_shards_test.go:72 | Return nil at shard-construction entry with ADAMIC_AUDIT_PROBE=P02 | Setup; probe only |

M02 survivor witness: native input `\fa{}` prints selector `"a"` and raw before `"\f"` before mutation, then selector `"\fa"` and raw before `""` after mutation. Both runs exit zero. M02-witness.json records the exact native command, cached switched build identity and complete outputs. This is changed behavior unguarded by the selected 33 members; coverage elsewhere is unknown. No equivalent candidates.

The following brief and environment issues affected the audit:

- The brief names an older commit, 8de93800f4. Fetching origin/main selected cf79ecec3723604428ab91ebcb283400d05a1548. All requested names still exist in top_level_shards_test.go, and none moved or vanished. Every location and standalone diff uses the fetched commit.

- The advertised two rows become 389 top-level functions. The 388 numbered wrappers call one checker with different ordinals, so they are one family. Setup performs construction and product preparation, so it stays separate. A union/stability test elsewhere asserts additional behavior and is outside this unit.

- TestMain in the generated file is not a listed Test row. My first source-name regex included it and rejected scope validation. I corrected the regex before any mutation, then checked all 389 requested names against go test -list.

- The full package timed out at 90 seconds while preparing a nonparallel composition check. Enabling the complete CSS corpus and running the requested slice also timed out. All three full-family-alone timing attempts timed out at 90 seconds without a completed individual assertion failure. Consequently there is no honest full-family median. The 33 evenly spaced members were chosen before mutant outcomes, passed cleanly, and constitute a bounded audit, not evidence of full-package uniqueness.

- A family spanning hundreds of leaf tests cannot reliably fit the brief's 90-second family-row budget. The requested whole-family timing still consumed three failed 90-second attempts. I retained the failure logs and reported null seconds with a 90-second lower bound, plus a separately labelled subset median. slow-worthy denotes proven bounded worth and this full-family cost lower bound.

- Warm tools did not include the pinned npm comparison packages or fixture repository. I installed postcss 8.5.16 and postcss-scss 4.0.9, ran npm ci in both stage3/api and that scratch directory, and fetched the pinned prettier fixture checkout cb4b33fba24a8428d00e54be85fc886288a374ea. No selected rows skipped after enabling both environments. Initial whole-package skip coverage is incomplete because that baseline cooked.

- Setup does not run the port. Returning empty output from the port would not test its construction oracle. It therefore gets its own nil-shard entry probe P02 and variant-dropping construction break S01; production-mutant build success is recorded as pass without treating that as semantic validation.

- The family mixes ordinary exact agreement checks and built-in planted-mutant witnesses. The witnesses require only a difference and accept completely empty output: P01 logged 198 native/Node witness acceptances. W01 proves the three designated witness checks fail when comparison is disabled, but that does not strengthen their accepted evidence. The ordinary agreement checks reject P01, so the family is not vacuous.

- Only four production mutants were used, rather than the approximate six-row target, to meet the port build guidance and budget. They were declared from four different port functions before observing kills. S01/W01 and P01/P02 never contribute to production uniqueness or subsumption.

- Source-file switching was used solely to amortize native rebuilding. Every selector reads /tmp/u080-mutant at execution time. Four freshly rebuilt switched products serve all matrix runs. Source-only standalone diffs have no switch and each was independently applied and built through the port's native product test. Cache logs prove misses for fresh switched code; no compiler/runtime source was mutated.

- The code-function list is conservative: all named functions/methods in the six copied port slices, plus the exact six-file entry value-import graph. To substantiate reach, I also collected Node V8 coverage over the 2,178 clean agreement inputs assigned to the 33 selected shards: six source modules and 70 reached V8 function records. This is source-side coverage, not native C coverage, and anonymous/transformed-code records cannot establish exact native function reachability.

- The survivor M02 changes form-feed tokenization. Its standalone native witness proves changed output, but no selected shard kills it. This must not be generalized to the unmeasured 355 family members or other package rows. The central replay must settle those results.

- The workspace is read-only under its default sandbox. Explicitly authorized writes, runs and push required execution escalations. Automatic reviews accepted every request; no action was blocked. Fixture fetch wall time was not recorded separately, so it is unknown rather than estimated.

Toolchain setup: skipped, env.sh worked; nproc=5. npm reported 362ms for stage3/api ci, 807ms for library install, and 225ms for library ci; these are npm lines, not independent whole-command wall measurements. Fixture clone wall time is unmeasured.

Clean timing lines: Setup [4.618, 4.503, 4.562]s (median 4.562); bounded family [12.905, 12.708, 12.797]s (median 12.797). Full family package timeout lines [90.063, 90.073, 90.066]s are failed run elapsed values, not completed test durations. Whole package and full slice baselines each reached 90 seconds.

Fresh switched products, lowering plus native build: parser 1.44+2.12s, builtin mutant-0 1.47+2.10s, mutant-1 1.42+2.11s, mutant-2 1.41+2.17s, totaling 14.24s. Production matrix binary elapsed: {'M01': 12.663, 'M02': 12.558, 'M03': 13.104, 'M04': 12.722, 'P01': 11.003}. Standalone verification total wall 23.837s for M01/M02/M03/M04/P01; per-product misses and hits are in build-times.json. S01, P02 and W01 each passed Go vet with the recorded overlays.

Not covered: full family completion, full-package red/green outcome after timeout, kills outside the fixed slice, repository uniqueness, Darwin-only leaks, exhaustive dynamic native reachability, or external-authority checks. The bounded baseline, all four production matrix runs, both own-entry probes, setup construction guard and embedded witness guard completed. No production source changes remain.

Replay: from the starting commit, apply any M01.diff through M04.diff individually and run the scope; each is switch-free. For the amortized recorded matrix, apply switch.diff, create /tmp/u080-mutant with the desired ID (empty selects baseline), and use the exact command in that ID-run.json with the enabled fixture/library environments. Probes and harness mutations are labelled separately. scope.json lists all family members and the 33 measured ones.
