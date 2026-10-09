All three requested rows are defended by distinct aimed production mutants.
D1, D4 and D2 each fail only their target among 274 executed top-level rows.
All evidence is retained here; two top-level skips and one skipped subcase remain unknown.

Starting origin/main: 7b9d4272c28f59530ab13daa5c49067e47933b06. Audit base: 8171b3173bdbfce1f7982d3c4f731279307ece37. Audit REPORT.md, rows.json, plan.json, reached-functions.txt and scope list were read and retained with audit- prefixes. The current go test -list . scope contains 276 top-level tests, versus 239 in the audit, with 37 additions and no vanished names. All current tests were included in every whole-package replay. No family grouping can turn these three independent unique top-level failures into shared catches.

CODE UNDER TEST: Adamic's Load/load/diagnostics/formatDiagnostic adapter for the direct Load constructor row; cycleFinder.fields for the direct inherited-field query; Lower/prepareSuper/constructorSuperFlow/useOfThis for pre-super refusal. These were declared before mutation in CODE-AND-ORACLE.md. The upstream TypeScript checker, tests, fixtures, Node and agreement harness were untouched.
ORACLE: the constructor row's handwritten CheckError and abstract/super text assertions; the cycle row's handwritten symbol-membership assertion, including Base as the declaration owner of parent; the pre-super row's handwritten Refused, reason and repair assertions. All target rows use self oracles. Their named subsumers were read in full. The current class inheritance test file has gained Node agreement checks in other rows since the audit; the three targets retain these specific oracles.

Coverage and semantic differences:
Each target and its named subsumer ran separately with timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower,./internal/load -coverprofile=<Name>.cover ./internal/lower/ -run '^Name$'. All six runs passed. The full commands and command wall times appear in coverage-runs.json, the profiles in <Name>.cover, and function inventories in <Name>.cover.functions.txt.
There are zero positively covered blocks exclusive to each target against its named subsumer. This is recorded in exclusive-coverage.json. The defenses therefore rest on semantic differences on shared lines:
- Constructor rules: TS2511 abstract-class construction versus private-identifier scope diagnostics. D1 returns an empty formatted string only for TS2511. The checker still rejects the program and CheckError is still returned, but Adamic loses the diagnostic's explanation. The target's abstract-construction subcase fails, and TestClassFeaturesPrivateChecker passes.
- Inherited fields: a direct query for all fields of Child, including Base.parent with its original declaration owner, versus an end-to-end generic cycle refusal. D4 returns at a property whose declaring symbol differs from the queried type's symbol. The direct membership assertion fails; the end-to-end generic row and every other executed row pass. This proves the helper's field-list contract has a separate guard. It does not claim every possible cycle program remains safe under this mutant.
- Pre-super this: explicit call syntax in the repair, call super(...), versus the conditional row's weaker call super substring. D2 removes the parentheses and argument placeholder from the production repair constant. Both programs remain refused; only the stronger repair assertion fails. This unique catch establishes repair-text coverage, not a unique catch of an admission bug.

Five standalone diffs were written before observing their outcomes. D1 and D2 are the sole attempts needed for their rows. D3, D4 and D5 are the three inherited-field attempts. D3 returns after the first data field and has three top-level catchers; the named generic subsumer passes. D4 uniquely catches the direct query. D5 omits the last selected field and has nine catchers, including the generic subsumer. These shared catches are evidence retained alongside the unique defense, not grounds to remove a test.

Every diff applies independently against the starting base, with no selector. Each installed mutant passed go vet on its mutated package: internal/load for D1, internal/lower for D2 through D5. Return-early changes and the repair constant change use the specified menu; D5 is an off-by-one bound. All five full-package replays completed within the 90-second binary budget. No panic, compilation failure, survivor, narrowed matrix or unknown executed top-level result occurred. Each replay used ADAMIC_GATE_UNCACHED=1 and a distinct /tmp/defend-class-inheritance/cache/<ID> build cache.

Full commands and failing lines are in rows.json and results.json. matrix.json contains every current top-level row's pass/fail/skip outcome for each mutant. passed-rows.json lists all passing rows. For each unique mutant, 273 other top-level rows passed and two skipped.

Restoration: runner restores original production text in finally after every run. Final git diff --exit-code -- internal succeeded. git apply --check succeeded for all five diffs on restored sources. The three restored targets passed in 0.101 binary seconds. No test was deleted, rewritten or weakened.

Brief ambiguities, costs and limits:
- The requested 15 GB floor is larger than /tmp's 8.8 GB capacity. Initial df reported 8.6 GB free in /tmp and 13 GB in /workspace. The identified earlier-unit scratch directory /tmp/defend-library-object was removed, then df was repeated. No repository or tools directory was deleted. Final df is retained in final-disk.log. No disk-related baseline failure occurred.
- The audit's old scope is smaller and other inheritance rows have stronger execution assertions now. Current scope governs this matrix; relying on the historical matrix alone would omit 37 tests.
- Zero exclusive coverage blocks did not settle the defense question. Diagnostic codes, complete direct-query results and stronger repair substrings distinguish these rows on shared code paths.
- Uniqueness is among the executed current package rows. TestOriginalCycleLedger requires the pinned pristine TypeScript corpus and generated diagnostics. TestOptionalWideningCensus requires a specified project configuration/output. The MixedUnionContractGraph interface/array subcase also skips. Their mutant responses and responses in other packages remain unknown.
- Two unique catches concern diagnostic delivery or repair text. They do not independently prove a runtime miscompile is prevented. The field row checks one concrete inherited field's name and original declaration owner; it does not assert field order or prove that an entire runtime cycle is refused. Its comment explicitly identifies the internal traversal contract.
- All three rows are defended, so there is no undefended name/assertion mismatch to report. None promises a speed threshold. Twin and cost-row rules were not applicable.
- No setup was needed; warm env.sh worked and nproc=5. npm ci in stage3/api ran before baseline and reported 395 ms. Baseline binary time was 34.083 seconds. Six coverage command walls total 25.955 seconds. Compilation/native construction is included in replay command wall times, not separately instrumented. Only the requested package test suite ran; vet checked the mutated packages. No main push or pull request was made.

| ID | Base file:line | Menu/change | Failed top-level rows |
|---|---|---|---|
| D1 | internal/load/load.go:245 | return early: Return an empty formatted diagnostic for TS2511 (abstract-class construction), leaving checker results and private-scope diagnostics intact. | TestInheritanceKeepsCheckerConstructorRules |
| D2 | internal/lower/class.go:755 | change constant: Lose the explicit call syntax in the pre-super repair; the conditional row only requires the weaker call-super prefix. | TestInheritanceRefusesThisBeforeSuperReturns |
| D3 | internal/lower/cycles.go:236 | return early: Return after the first data field, losing an inherited field behind an own field, while a subclass with just an inherited generic slot can retain that slot. | TestInheritanceCycleFinderIncludesInheritedFields, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatZeroOneRefusesIsRefusedWithAFix |
| D4 | internal/lower/cycles.go:234 | return early: Stop on a property declared by another class, incorrectly treating the inherited-field boundary as the end of the shape. | TestInheritanceCycleFinderIncludesInheritedFields |
| D5 | internal/lower/cycles.go:239 | off by one bound: Omit the final selected field, which is the inherited parent after the child label in the direct traversal witness. | TestClassFeaturesStaticInterfaceCycle, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticSoundness, TestInheritanceCycleFinderIncludesInheritedFields, TestInheritanceGenericSoundness, TestInheritanceRefusesUnsoundOverrides, TestLiteralMethodCapturesCannotMakeCycles, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatZeroOneRefusesIsRefusedWithAFix |

Replay timings:
- D1: vet wall 0.229s; replay wall 49.585s; binary 37.673s.
- D2: vet wall 0.498s; replay wall 43.097s; binary 35.470s.
- D3: vet wall 0.588s; replay wall 43.402s; binary 35.226s.
- D4: vet wall 0.521s; replay wall 42.403s; binary 34.447s.
- D5: vet wall 0.468s; replay wall 42.131s; binary 34.400s.

```json
[
  {
    "test": "TestInheritanceKeepsCheckerConstructorRules",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestClassFeaturesPrivateChecker",
    "defense": "defended",
    "unique_mutant": "D1 internal/load/load.go:245",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/load/load.go:245",
        "change": "Return an empty formatted diagnostic for TS2511 (abstract-class construction), leaving checker results and private-scope diagnostics intact.",
        "rows_failed": [
          "TestInheritanceKeepsCheckerConstructorRules"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-class-inheritance/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D1.log 2>&1; class_inheritance_test.go:82: want abstract, got"
  },
  {
    "test": "TestInheritanceCycleFinderIncludesInheritedFields",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestInheritanceGenericSoundness",
    "defense": "defended",
    "unique_mutant": "D4 internal/lower/cycles.go:234",
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/lower/cycles.go:236",
        "change": "Return after the first data field, losing an inherited field behind an own field, while a subclass with just an inherited generic slot can retain that slot.",
        "rows_failed": [
          "TestInheritanceCycleFinderIncludesInheritedFields",
          "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
          "TestWhatZeroOneRefusesIsRefusedWithAFix"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/lower/cycles.go:234",
        "change": "Stop on a property declared by another class, incorrectly treating the inherited-field boundary as the end of the shape.",
        "rows_failed": [
          "TestInheritanceCycleFinderIncludesInheritedFields"
        ]
      },
      {
        "mutant": "D5",
        "file_line": "internal/lower/cycles.go:239",
        "change": "Omit the final selected field, which is the inherited parent after the child label in the direct traversal witness.",
        "rows_failed": [
          "TestClassFeaturesStaticInterfaceCycle",
          "TestClassFeaturesStaticParentCycle",
          "TestClassFeaturesStaticSoundness",
          "TestInheritanceCycleFinderIncludesInheritedFields",
          "TestInheritanceGenericSoundness",
          "TestInheritanceRefusesUnsoundOverrides",
          "TestLiteralMethodCapturesCannotMakeCycles",
          "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
          "TestWhatZeroOneRefusesIsRefusedWithAFix"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-class-inheritance/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D4.log 2>&1; class_inheritance_test.go:127: cycle traversal lost Base.parent when visiting Child"
  },
  {
    "test": "TestInheritanceRefusesThisBeforeSuperReturns",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestInheritanceConditionalThisRules",
    "defense": "defended",
    "unique_mutant": "D2 internal/lower/class.go:755",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/class.go:755",
        "change": "Lose the explicit call syntax in the pre-super repair; the conditional row only requires the weaker call-super prefix.",
        "rows_failed": [
          "TestInheritanceRefusesThisBeforeSuperReturns"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-class-inheritance/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D2.log 2>&1; class_inheritance_test.go:179: want pre-super this refusal with fix, got /tmp/adamic-gate/TestInheritanceRefusesThisBeforeSuperReturns4048356991/001/main.a:3:37: Adamic 0.1 refuses this before super returns; call super before using this"
  }
]
```

Publication note: the first push found existing remote defense history, including different inheritance rows and an earlier pre-super defense. This session's evidence was moved into constructor-fields-7b9d4272 before merging that history. Commands in the saved run records retain their original log/profile paths as executed. Replay scripts now resolve their own evidence directory so they cannot overwrite the earlier root-level evidence. No remote history was rewritten.
