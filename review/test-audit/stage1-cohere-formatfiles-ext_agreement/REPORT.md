Unit u090: TestExtAgreesWithGo exists at base 3a1c8b5792fb66503290d2d494ce96302bc4386d.
Clean whole-package baseline passed; all 50 top-level tests ran in every matrix mode.
Verdict sacred: M01 and M03 each failed only this row in the package.
Three production mutants caught; empty-answer probe caught; built-in witness weakening caught.
Evidence under review/test-audit/stage1-cohere-formatfiles-ext_agreement; production source restored.

[
  {
    "test": "TestExtAgreesWithGo",
    "package": "stage1/cohere/formatfiles",
    "file": "stage1/cohere/formatfiles/ext_agreement_test.go",
    "seconds": 0.172,
    "oracle": "Go filepath.Ext executed on all 17 table inputs; Node executes the port and JSON bytes are compared. The built-in index-zero negative control additionally requires a self-written empty JSON string.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M01",
      "M03"
    ],
    "last_proven_fail": "M03: ext_agreement_test.go:74: ext(\"x.y/z\"): Node \".y/z\", filepath.Ext \"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [
      "no_dot",
      "root",
      "empty",
      "dot_before_separator"
    ],
    "bounded": false,
    "matrix_rows": [
      "TestExtAgreesWithGo",
      "TestFormatfilesShardPlantedDisagreement",
      "TestProduct_FormatfilesGoOracle",
      "TestProduct_FormatfilesNative family",
      "TestProduct_FormatfilesNativeUnsanitized",
      "TestThePortParsesAsGoCohereDoes family"
    ],
    "evidence": "printf M03 > /workspace/u090-selector; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/formatfiles/ -run . > M03.log 2>&1; ext_agreement_test.go:74: ext(\"x.y/z\"): Node \".y/z\", filepath.Ext \"\""
  }
]

Code under test, named before mutation: ext in stage1/cohere/formatfiles/golang.ts, executed by Node for this row. Its only invoked port function is ext at origin line 45; it calls Node String.slice. Other exports and imported clean are not called by this row. The copied driver, Go helpers, JSON encoding and filepath.Ext are harness/oracle, not production mutation targets.
Oracle, named before mutation: Go filepath.Ext, called for each of 17 paths and encoded with Go json.Marshal. The port is run directly through oracle/node.mjs. This row does not compile or execute native code, even though other package rows do. Its built-in negative control also requires the self-written JSON empty string result.
The fixed menu is plan.json, generated before executing any mutant. Only ext is reached, so mutations span its loop bound, dot condition and separator constant rather than unrelated exports.

| id | origin file:line | change | failed rows |
|---|---|---|---|
| M01 | stage1/cohere/formatfiles/golang.ts:46 | index >= 0 -> index > 0 | TestExtAgreesWithGo |
| M02 | stage1/cohere/formatfiles/golang.ts:47 | path[index] === '.' -> path[index] !== '.' | TestExtAgreesWithGo, TestThePortParsesAsGoCohereDoes family |
| M03 | stage1/cohere/formatfiles/golang.ts:46 | path[index] !== '/' -> path[index] !== '\\' | TestExtAgreesWithGo |
| P01 probe | stage1/cohere/formatfiles/golang.ts:45 | ext returns empty string at entry | TestExtAgreesWithGo, TestThePortParsesAsGoCohereDoes family |
| W01 witness weakening | stage1/cohere/formatfiles/ext_agreement_test.go:73 | shared byte-equality calls always return true | TestExtAgreesWithGo: ext_agreement_test.go:92: index-zero mutant agrees with filepath.Ext on .i |

Survivors: none among the three planted production mutants.

Native rebuild validation:
M01: native sanitized main.ts rebuild 5.170s, exit 0. Command and apply check in compile-status.json; sources copied from the base, with only that standalone change, no selector.
M02: native sanitized main.ts rebuild 4.641s, exit 0. Command and apply check in compile-status.json; sources copied from the base, with only that standalone change, no selector.
M03: native sanitized main.ts rebuild 4.204s, exit 0. Command and apply check in compile-status.json; sources copied from the base, with only that standalone change, no selector.
P01: native sanitized main.ts rebuild 3.161s, exit 0. Command and apply check in compile-status.json; sources copied from the base, with only that standalone change, no selector.
The inactive switched source built 14 native products once; every later matrix mode logged 14 cache hits and zero native cache misses. Each reported product build duration is in selector-native-builds.json and listed here:
formatfiles_shards_test.go:182: build formatfiles-native-a_trailing_dot_no_extension bd58cba882e0 miss 3.11
formatfiles_shards_test.go:182: build formatfiles-native a3c8abaf98cf miss 2.74
formatfiles_shards_test.go:182: build formatfiles-native-an_extension_lowercased_as_JavaScript_does 653637d8740c miss 3.88
formatfiles_shards_test.go:182: build formatfiles-native-the_declined_extensions_sorted_by_UTF-16_unit c33b27cab7b5 miss 3.96
formatfiles_shards_test.go:182: build formatfiles-native-Adamic_files_declined_instead_of_held_back d137fede13fa miss 4.24
formatfiles_shards_test.go:182: build formatfiles-native-a_bare_.a_name_held_back_as_Adamic 6302e16e312d miss 4.48
formatfiles_shards_test.go:182: build formatfiles-native-a_.git_link_to_nothing_taken_for_a_repository ed0acbc46e86 miss 4.63
formatfiles_shards_test.go:182: build formatfiles-native-uppercase_.A_held_back_as_Adamic 449f37ea8190 miss 4.94
formatfiles_shards_test.go:182: build formatfiles-native-a_link_round_a_loop_taken_for_nothing e7ea3ebd71ab miss 5.00
formatfiles_shards_test.go:182: build formatfiles-native-a_.prettierignore_link_to_nothing_let_through 78e3a1131083 miss 4.94
formatfiles_shards_test.go:182: build formatfiles-native-DEL_passed_by_the_quoting's_fast_path 7913aadb9c79 miss 4.80
formatfiles_shards_test.go:182: build formatfiles-native-a_link_to_nothing_taken_for_nothing 1ed0d72cb1e0 miss 4.68
formatfiles_prepare_test.go:117: build formatfiles-native-unsanitized 50c6a9c5801b miss 2.80
formatfiles_shards_test.go:182: build formatfiles-native-a_quote_passed_by_the_quoting's_fast_path b9fafaff618d miss 4.30

Brief ambiguities, costs and limitations:
1. The brief describes stage1 ports as requiring native rebuilds, but the assigned row runs the TypeScript source on Node only. Native validation was still performed for every standalone mutant and probe using the actual sanitized port main.ts build. Matrix runs include the entire package, so native products execute in other rows.
2. This row is a hybrid: 17 ordinary production agreement cases plus a built-in planted-mutant witness. Production kills are evidenced by ordinary ext(...) comparison failures, not by the built-in witness or replacement-precondition failures. W01 separately makes its byte comparator always declare agreement and the built-in witness then fails. The whole row remains sacred rather than witness-only.
3. M01 is the same loop-bound change as the existing built-in index-zero negative control. It was included in the fixed code-derived loop/condition/constant menu; the planted matrix source changes the port, not the test or its expected values. M03 is a second package-unique kill independent of that built-in mutant.
4. The selector preserves original textual replacement markers so the packages own mutants can still be installed. Standalone diffs contain only the deliberate port change. Standalone M01 removes the index >= 0 marker used by the target rows own negative control, so central replay also reaches that replacement-precondition failure after ordinary positive comparisons already fail. The direct semantic ext failures are the evidence here. Native compilation of its standalone port passes.
5. The first selector attempt treated readTextFile as a string, but it returns an Ok/Error result. Native preparation rejected that instrumentation with TS2367. This pre-mutant scaffolding failure was excluded and retained in rejected-selector-result-shape.log. Unwrapping the result produced a green full inactive-selector run before mutants.
6. No test moved or vanished from the requested file. The package has 50 observed top-level tests. The 32 formatter shards and their union form one family; sanitized per-recipe native product wrappers are grouped separately from the unsanitized and Go-oracle preparation rows. Raw failed_tests and tests_run are retained in matrix.json, so grouping is reviewable. All rows ran in all modes; no panic, timeout, or skipped row obscured uniqueness.
7. P01 is a probe, not a kill used for sacred. Four agreement subcases passed its empty result: no_dot, root, empty, dot_before_separator. Their correct Go answers are already empty. The full row failed on 13 nonempty answers, so vacuous is false. These passing subcases are reported without implying that their legitimate empty expected answers are defective.
8. Warm env.sh worked, so setup was skipped. npm ci in stage3/api added three packages in 377ms. The Node runner uses built-in type stripping and local modules; no further node_modules directory was needed.
9. The selector is a fixed file read inside the port source and is constant for the lifetime of each subprocess. Its contents do not need to be in the native cache key because the compiled products read it at runtime. Standalone sources have no selector or harness change. Node transformation occurs per process; cached native products are reused across selector modes.
10. Repository-wide uniqueness, other packages and extension inputs outside the fixed 17-case table were not audited. The one-time Go CLI build was not separately timed; measured native rebuild durations include lowering and clang invocation. Product cache reports time native construction as emitted by the existing harness. No elapsed duration was invented.

Timings:
{
  "setup_seconds": 0,
  "nproc": 5,
  "npm_reported_seconds": 0.377,
  "clean_package_seconds": 38.298,
  "inactive_selector_package_seconds": 30.321,
  "timing_wall_seconds": 6.95215253599963,
  "matrix_wall_seconds": 105.63814550999814,
  "matrix_binary_seconds": 97.16,
  "standalone_native_rebuild_seconds": {
    "M01": 5.170462270001735,
    "M02": 4.641349157001969,
    "M03": 4.203751954002655,
    "P01": 3.1606238249987655
  },
  "selector_native_product_build_reported_seconds": 58.5,
  "witness_wall_seconds": 6.5236884570003895
}
