All three rows are defended by distinct aimed production mutants.
Each whole-package replay failed only its target among 274 executed top-level rows.
Two top-level rows and one subcase skipped; their mutant responses remain unknown.

Starting origin/main: d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4.
Audit base: e2492670b06a4dce1deafe837158ce0366cf2bc4. The fetched audit report, code/oracle notes, menu, row results and scope are retained here. Current scope has 276 top-level tests versus 239 at audit time, with 37 added and none vanished. All current rows were included in each matrix.

CODE UNDER TEST and ORACLE were declared before mutation in CODE-AND-ORACLE.md. Production changes concern String internal-slot admission, tuple array-method dispatch, and detached-method repair generation. Library-string and tuple expectations are self-written diagnostics. Method refusal expectations are self-written; three current positive neighbors also run Node and compare generated JavaScript and native execution. One function-field positive neighbor only excludes Refused and can pass with NotYet. No tests, checker, Node, fixtures or harness were changed.

Coverage commands ran each target and its named subsumer separately using timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower -coverprofile=<Name>.cover ./internal/lower/ -run '^Name$'. All five coverage runs passed. Full profiles and exclusive blocks are retained. Exclusive blocks against the named subsumer number 89 for library-string, 93 for tuple, and 1125 for method-value. These counts do not establish exclusivity against all package rows. The leads used were library_string.go:128 (String internal-slot guard), object.go:786 (direct tuple method dispatch), and refusals.go:200 (parameter forwarding in a suggested repair). Semantic differences are valueOf versus trim, tuple.join versus tuple-to-array conversion, and a required-argument method versus a zero-argument method. The enlarged method coverage also includes its runtime agreement neighbors, so not every exclusive block concerns repair generation.

D1 changes the valueOf recognition constant to valueOff. The invalid String.prototype.valueOf.call(42) call is admitted. Its target fails on the missing internal-slot refusal. An independent unmodified Node witness throws TypeError, recorded in node-internal-slot-witness.log; this does not change the test's self oracle classification.
D2 drops the complete tuple method guard. Its target observes an inherited-library-member Refused instead of the promised representation-specific NotYet. It defends diagnostic classification and the unsupported-operation boundary, not a runtime execution result.
D3 drops the complete parameter-name collection loop. Its target observes the invalid zero-argument repair () => shelter.admit() instead of (pet) => shelter.admit(pet). Zero-argument repairs exercised elsewhere remain correct.

Each standalone diff applies cleanly to the starting base and passed go vet ./internal/lower/ while installed. D2 drops a whole if statement; D3 drops a whole loop, leaving no unused variables. run-defense.py restores each original source in finally. Final git diff --exit-code -- internal passed, all three git apply --check commands passed, and the restored target test run passed. No mutant survived; each failed exactly one top-level row. passed-rows.json lists all 273 passing rows for each replay. Full failing lines and commands are in rows.json and results.json. No panic or timeout occurred, and no narrowed matrix was needed.

Limitations and costs in the brief:
- The 15 GB disk floor cannot be reached on the 8.8 GB /tmp filesystem. Initial free space was 8.6 GB in /tmp and 13 GB in /workspace. The only identified earlier-unit scratch directory, /tmp/defend-prototype, was removed. Neither repository nor tools were deleted. The final disk check is retained; no disk-related baseline failure occurred.
- The package gained 37 tests, and the method row gained actual Node/native agreement checks since the audit. test-changes-since-audit.diff records the assertion changes. Old oracle descriptions cannot simply be reused.
- TestOriginalCycleLedger requires a pinned pristine TypeScript corpus and generated diagnostic evidence. TestOptionalWideningCensus requires a caller-selected project configuration and output. Neither was enabled with arbitrary substitute inputs. The MixedUnionContractGraph interface/array subcase also remains skipped. Unique catches here are against all executed rows, not proof concerning skipped rows or other packages.
- Coverage identifies blocks exclusive against named subsumers, not necessarily the whole package. The subsequent whole-package matrices settle the executed-row uniqueness claim.
- None of these rows promises a timing threshold. All were defended, so there is no undefended name/assertion mismatch to report. Diagnostic tests guard classification and repair text; they do not execute every suggested repair. The one weak function-field neighbor noted above remains an owner finding.

Warm toolchain setup was skipped, nproc=5. npm ci ran before the clean baseline (npm.log). Baseline binary time was 35.076 seconds. Coverage wall times, including command/build overhead, were 8.478, 2.278, 2.100, 2.238 and 2.407 seconds. Mutant build and native construction costs are included in replay wall times rather than separately instrumented. No other package test suite was run. No test was deleted, rewritten or weakened; no main push or pull request was made.

D1: vet wall 0.539s; full replay wall 45.310s; test binary 37.408s; only failing row TestLibraryStringRefusals.

D2: vet wall 0.512s; full replay wall 45.829s; test binary 37.598s; only failing row TestATupleSeenAsAnArrayIsNotYet.

D3: vet wall 0.483s; full replay wall 46.051s; test binary 37.951s; only failing row TestAMethodReadAsAValueIsRefused.

```json
[
  {
    "test": "TestLibraryStringRefusals",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestWhatZeroOneRefusesIsRefusedWithAFix",
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/library_string.go:128",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/library_string.go:128",
        "change": "Stop recognizing valueOf in the String internal-slot guard, admitting valueOf.call(42) through numeric string conversion.",
        "rows_failed": [
          "TestLibraryStringRefusals"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-library-object/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D1.log 2>&1; library_string_test.go:27: want refusal containing \"requires a String internal slot\", got <nil>"
  },
  {
    "test": "TestATupleSeenAsAnArrayIsNotYet",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
    "defense": "defended",
    "unique_mutant": "D2 internal/lower/object.go:786",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/object.go:786",
        "change": "Drop the whole tuple-method dispatch guard, leaving tuple.join to the ordinary fallback instead of its representation-specific refusal.",
        "rows_failed": [
          "TestATupleSeenAsAnArrayIsNotYet"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-library-object/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D2.log 2>&1; lower_test.go:604: got /tmp/adamic-gate/TestATupleSeenAsAnArrayIsNotYetan_array_method_called_on_a_tupl1388780747/001/main.a:2:13: Adamic 0.1 refuses inherited library member join read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method), want a not-yet ending \"main.a:2:13: stage 0 can't lower join on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) yet\""
  },
  {
    "test": "TestAMethodReadAsAValueIsRefused",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestLibraryStringRefusals",
    "defense": "defended",
    "unique_mutant": "D3 internal/lower/refusals.go:200",
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/lower/refusals.go:200",
        "change": "Drop the whole parameter-name collection loop; suggested arrows for detached methods no longer forward required parameters.",
        "rows_failed": [
          "TestAMethodReadAsAValueIsRefused"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-library-object/cache/D3 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D3.log 2>&1; lower_test.go:644: got /tmp/adamic-gate/TestAMethodReadAsAValueIsRefusedheld_in_a_variable4123879763/001/main.a:11:15: Adamic 0.1 refuses a method read as a value (admit would lose its object, and this with it); call it in an arrow that keeps the object: () => shelter.admit() (unbound-method), want a refusal containing \"main.a:11:15: Adamic 0.1 refuses a method read as a value (admit would lose its object, and this with it); call it in an arrow that keeps the object: (pet) => shelter.admit(pet) (unbound-method)\""
  }
]
```
