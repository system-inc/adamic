u160 audited four rows at cf735d9fba9e38de6368575e5630e44375a86eaf.
Clean baseline passed; no skips; warm setup skipped; nproc=5.
Twelve production mutants, two weakened checks and one empty-entry probe ran over the whole package.
Two rows are sacred; two are witnesses; no row passed its own empty-entry probe.
M10 and M11 survived with observed output differences; repo-wide uniqueness remains untested.

```json
[
  {
    "test": "TestCompiler41231d51Shape",
    "package": "stage3/census/latent/refusalrewrite",
    "file": "stage3/census/latent/refusalrewrite/rewrite_test.go",
    "seconds": 0.008,
    "oracle": "Self-written: production refuse AST must be unchanged and latentRefuse must contain exactly four defers. Counting defers does not check owner behavior.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: rewrite_test.go:53: expected independently instrumented contracts and visit functions with scoped owners, got 2",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9",
      "M10",
      "M11",
      "M12"
    ],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestCompiler41231d51Shape",
      "TestMissingFunctionMutantFailsLoudly",
      "TestChangedVisitorMutantFailsLoudly",
      "TestVisitorsCollectContinueAndSkipDiagnosedBodies"
    ],
    "evidence": "ADAMIC_MUTANT=M2 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/refusalrewrite/ -run . > M2.log 2>&1; rewrite_test.go:53: expected independently instrumented contracts and visit functions with scoped owners, got 2",
    "timing_samples": [
      0.008,
      0.008,
      0.007
    ]
  },
  {
    "test": "TestMissingFunctionMutantFailsLoudly",
    "package": "stage3/census/latent/refusalrewrite",
    "file": "stage3/census/latent/refusalrewrite/rewrite_test.go",
    "seconds": 0.003,
    "oracle": "Self-written missing-function diagnostic substring and nil partial output. Exact label matters: M12 fails although rejection still works. Witness of Rewrite validation.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: rewrite_test.go:66: missing-function mutant not caught: output=10425 error=<nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9",
      "M10",
      "M11",
      "M12"
    ],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestCompiler41231d51Shape",
      "TestMissingFunctionMutantFailsLoudly",
      "TestChangedVisitorMutantFailsLoudly",
      "TestVisitorsCollectContinueAndSkipDiagnosedBodies"
    ],
    "evidence": "ADAMIC_MUTANT=W1 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/refusalrewrite/ -run . > W1.log 2>&1; rewrite_test.go:66: missing-function mutant not caught: output=10425 error=<nil>",
    "timing_samples": [
      0.003,
      0.002,
      0.003
    ],
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestChangedVisitorMutantFailsLoudly",
    "package": "stage3/census/latent/refusalrewrite",
    "file": "stage3/census/latent/refusalrewrite/rewrite_test.go",
    "seconds": 0.003,
    "oracle": "Self-written changed-child-walk diagnostic substring. Exact label matters: M12 fails although rejection still works. Witness of Rewrite validation.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: rewrite_test.go:82: changed visitor mutant not caught: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9",
      "M10",
      "M11",
      "M12"
    ],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestCompiler41231d51Shape",
      "TestMissingFunctionMutantFailsLoudly",
      "TestChangedVisitorMutantFailsLoudly",
      "TestVisitorsCollectContinueAndSkipDiagnosedBodies"
    ],
    "evidence": "ADAMIC_MUTANT=W2 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/refusalrewrite/ -run . > W2.log 2>&1; rewrite_test.go:82: changed visitor mutant not caught: <nil>",
    "timing_samples": [
      0.003,
      0.004,
      0.003
    ],
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestVisitorsCollectContinueAndSkipDiagnosedBodies",
    "package": "stage3/census/latent/refusalrewrite",
    "file": "stage3/census/latent/refusalrewrite/rewrite_test.go",
    "seconds": 0.253,
    "oracle": "Self-written ordered findings, continuation, diagnosed-body exclusion and first production refusal, executed in a synthetic Go program. Child go test must exit zero; logs distinguish compilation errors from semantic mismatches.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9"
    ],
    "unique_kills": [
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9"
    ],
    "last_proven_fail": "M9: rewrite_test.go:188: compiled behavior probe: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9",
      "M10",
      "M11",
      "M12"
    ],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestCompiler41231d51Shape",
      "TestMissingFunctionMutantFailsLoudly",
      "TestChangedVisitorMutantFailsLoudly",
      "TestVisitorsCollectContinueAndSkipDiagnosedBodies"
    ],
    "evidence": "ADAMIC_MUTANT=M9 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/refusalrewrite/ -run . > M9.log 2>&1; rewrite_test.go:188: compiled behavior probe: exit status 1",
    "timing_samples": [
      0.253,
      0.273,
      0.229
    ]
  }
]
```

| ID | File:line at starting origin/main | Change | Failed rows |
|---|---|---|---|
| M1 | rewrite.go:69 | change constant: clone.Name.Name = "latentRefuse" -> clone.Name.Name = "latentRefused" | Shape, Behavior |
| M2 | rewrite.go:161 | drop statement: defer func(){ latentFindingOwner = outerOwner }() -> (drop); Also drop the now unused outerOwner declaration at line 159. | Shape |
| M3 | rewrite.go:167 | drop statement: latentRecord(%s); %s = nil; node.ForEachChild(%s); latentStop = false -> latentRecord(%s); node.ForEachChild(%s); latentStop = false; Remove the unused corresponding fmt.Sprintf argument at line 167. | Behavior |
| M4 | rewrite.go:167 | change constant: latentStop = false }() -> latentStop = true }() | Behavior |
| M5 | rewrite.go:165 | off-by-one bound: len(l.program.LatentDiagnosticsIn(node.Body())) > 0 -> len(l.program.LatentDiagnosticsIn(node.Body())) > 1 | Behavior |
| M6 | rewrite.go:220 | drop statement: collectReturns(clone.Body) -> // collectReturns dropped | Behavior |
| M7 | rewrite.go:155 | drop statement: eraseWalks(walker.Body) -> // eraseWalks dropped | Behavior |
| M8 | rewrite.go:241 | change constant: Fun: ast.NewIdent("latentRecord"), Args: ret.Results -> Fun: ast.NewIdent("latentRecord"), Args: []ast.Expr{ast.NewIdent("nil")}; Discard the now unused ReturnStmt binding on line 240. | Behavior |
| M9 | rewrite.go:256 | drop statement: rewriteBlocks(child, rewrite) -> (drop); Discard the now unused BlockStmt binding at line 255. | Behavior |
| M10 | rewrite.go:158 | flip condition: if latentFullEnabled() && node.Kind == ast.KindFunctionDeclaration -> if latentFullEnabled() \|\| node.Kind == ast.KindFunctionDeclaration |  |
| M11 | rewrite.go:217 | off-by-one bound: badReturn \|\| outerReturns < 2 -> badReturn \|\| outerReturns < 3 |  |
| M12 | rewrite.go:13 | change constant: const function = "lowering.refuse" -> const function = "lowering.refused" | Missing, Changed |
| W1 | rewrite.go:54 | weaken missing-function validation: return nil, fail("expected exactly one function, found %d", len(matches)) -> return source, nil | Missing |
| W2 | rewrite.go:151 | weaken child-walk validation: if count != 1 { -> if count > 1 { | Changed |
| P1 | rewrite.go:37 | empty entry probe: func Rewrite(source []byte) ([]byte, error) { -> func Rewrite(source []byte) ([]byte, error) { 	return nil, nil; Replace whole Rewrite body with empty return to avoid unreachable code. | Changed, Missing, Shape, Behavior |

All source locations above refer to stage3/census/latent/refusalrewrite/rewrite.go at the stated starting commit. Shape, Missing, Changed and Behavior abbreviate the four rows in JSON order. M IDs are production mutants; W IDs only support witness verdicts; P1 is only a vacuity probe. No families, helpers, setup checks, panics of the parent binary, timed-out runs or skipped rows occurred. M5 panicked in the child behavior process, whose captured failure did not abort the package matrix. Every standalone diff passed git apply --check and go vet, and its whole-package replay matched the switched matrix.

Survivors:
- M10: go run ./stage3/census/latent/refusalrewrite/cmd -input stage3/census/latent/refusalrewrite/testdata/refusals-41231d51.go.txt -output <artifact> wrote `if latentFullEnabled() && node.Kind == ast.KindFunctionDeclaration {` before and `if latentFullEnabled() || node.Kind == ast.KindFunctionDeclaration {` after. This is unguarded generated control flow. The witness demonstrates generated output, not a runtime full-mode behavior comparison. See M10-witness-before.go.txt and M10-witness-after.go.txt.
- M11: the same command with two-outer-returns.go.txt succeeded before (exit 0, generated output saved) and failed after (exit 1, `latent refusal rewrite: lowering.refuse: unsupported outer refusal returns`). Unguarded acceptance of a supported two-return input. See M11-witness-before.go.txt and M11-witness-after.log.

Brief ambiguities and costs:
- The brief names no row list for a whole package, correctly resolved by go test -list. Four distinct bodies, no family merging was appropriate.
- Tests named MutantFailsLoudly validate production Rewrite input guards, rather than an external agreement comparator. Following the explicit witness rule, W1 returns the unchanged input when the missing-function guard fires; W2 permits zero child walks. Their failures alone support witness verdicts. Production diagnostic mutation M12 is recorded in the raw matrix but excluded from their verdict kills.
- The behavior test runs Go, but the brief's external-run category names Go cohere, not arbitrary Go execution. Its ordered expected findings are self-written, so all four rows are self.
- No external source is cited by these expected values; none was checked against an outside authority.
- The shape test's defer count catches M2, but does not prove owner restoration semantics. The behavior fixture pins latentFullEnabled to false and does not exercise full-mode restoration.
- A child nonzero exit alone is weaker than a diagnostic or value comparison. M7, M8 and M9 fail by compilation; M3 and M4 additionally demonstrate semantic findings mismatches. The complete child output is retained in each log.
- Drop-statement mutants required cleanup of dead bindings or fmt arguments. M2 also drops the unused capture; M3 removes the corresponding fmt argument; M8 and M9 discard unused AST bindings. These are compilation repairs, not inserted behavioral statements.
- Initial M8 and P1 replay diffs failed vet for an unused binding and unreachable code. Corrected them, reran vet, and independently replayed all final diffs. Final P1 replaces the entire Rewrite body with return nil,nil. The switch uses an entry return guarded by its selector.
- At most three mutants per row conflicts mildly with witness-only verdict rules. Twelve production mutants were selected across generation, traversal, bounds and diagnostics before any matrix failures were inspected; two separate guard weakenings test witnesses.
- A runtime guard survivor can be witnessed by its generated source output because Rewrite's output is Go source. M10's runtime effect itself was not measured.

CODE UNDER TEST: Rewrite and package helpers fail, ident, printed, statements, eraseWalks, collectReturns, rewriteBlocks. This is the static package call graph reached by these rows, including Rewrite's AST callbacks; no dynamic coverage profile was produced. cmd/main.go is only used to observe survivor outputs and was never mutated. ORACLE: the test's self-written AST counts, diagnostic substrings and ordered behavior findings, all retained unchanged.

Timing and scope: warm env verification succeeded; setup 0 seconds; stage3/api npm ci completed before baseline. See costs.json for separated build and recorded run driver times, runs.json for every final command cost, and the twelve isolated logs for binary timing samples. Native products were not involved. No other package's tests ran. Package uniqueness is proven only for the observed production matrix, not repo-wide. The source was restored and a clean post-audit baseline passed. No push to main and no pull request.
