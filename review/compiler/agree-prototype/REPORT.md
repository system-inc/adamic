Built: source-row census for #41bkfdw, wave 2; all 345 rows are refusals, so zero conversions apply.
Commits: base 62286991694debbb14fdb981d8f0957636c90dbc; audit authority e7d2a4ed92c0f49668243ace1017bffc33bb338e.
Commands and outputs: focused six-test baseline passed, 13.496s; setup passed, 280.851s; lane evidence is in lane-checks.log.
Mutants: exact audit M20 survived this scope; own one-line refusal-explanation mutant failed both member-access rows at prototype_test.go:204.
Not covered: converted acceptance behavior or a converted-row mutant kill, because prototype_test.go has no acceptance rows; regex acceptance tests are outside this territory.

The ruling classifies a source by its assertion. Successful declaration lowering inside the inherited-member sweep is setup; its final assertion requires Refused and a nil expression. It is not an acceptance assertion or an IR-only ownership assertion. The five legacy members absent from es2024 require checker rejection when load rejects them. All of those rows belong to the refuse class.

| Test | Source rows | Classification |
|---|---:|---|
| TestInheritedLibraryReadsNeverLoadOwnFields | 288 | refuse |
| TestPrototypeMethodsAreRefusedWithReasons | 11 | refuse |
| TestNullishPrototypeReadsAreRejectedByChecker | 24 | refuse |
| TestPrototypeHazardsBehindObjectViewsAreNotYet | 12 | refuse |
| TestUnrepresentedPrototypeCallsAreNotYet | 8 | refuse |
| TestIsPrototypeOfReadsExplainThePrototypeRefusal | 2 | refuse |
| Total | 345 | 0 accept, 345 refuse, 0 IR-only |

Every concrete row is listed in rows.md. classify.py expands 12 receivers, Node's 12 Object.prototype members, both sweep access forms, and the nullish and string tables. No row was deleted or skipped at runtime; every refusal test is unchanged. No fixture was added, so counts.md was not regenerated. No golden IR snapshot or acceptance IR assertion was added.

Step 5 is inapplicable to this file. Conservative scope choice: do not create a new regex acceptance row or modify regexp_test.go to manufacture a conversion. The audit diff is authoritative and is preserved as M20.patch. It changes flags detection from v to u in escapeRegexSource. The corrected source observation is `[[/]\/`, one backslash before the final slash. Applying that exact patch and running all six unit tests exits 0. This is a survivor, not a claimed kill. This refusal-only unit cannot provide the requested converted-row failure, nor the second converted-row mutant proof.

As a control on the refusal tests, own-refusal-reason.patch removes only the phrase `Adamic has no observable prototype chain` from prototypeRead's isPrototypeOf explanation. TestIsPrototypeOfReadsExplainThePrototypeRefusal exits 1 in 0.072s, with both dot and bracket source rows failing their explanation assertion, not compilation. Both production files were restored byte for byte in finally blocks. The restored test run is recorded in restored.jsonl. No production or test changes are committed.

All shells source /workspace/adamic-tools/env.sh. GOPROXY was set to `https://proxy.golang.org|direct` before `timeout 300 bash cloud/setup.sh`. Setup succeeded. Its timing lines: Node ready 0.034s, Go ready 0.036s, clang ready 0.282s, Markdown dependencies installed step-duration 1.426s and ready 1.525s, submodules ready 17.642s, Go build ready 280.710s, test binaries deferred 280.821s, build cache warm 280.822s, done 280.851s. nproc is 5; cpu.max is 400000 100000, a four-CPU quota. Versions: Node 24.19.0, Go 1.27.1, clang 20.1.8. Setup's build cache warming overlapped the baseline compilation; no runtime speedup is inferred.

Exact focused command (baseline, M20, and restored):

```sh
go test ./internal/lower -run '^(TestInheritedLibraryReadsNeverLoadOwnFields|TestPrototypeMethodsAreRefusedWithReasons|TestNullishPrototypeReadsAreRejectedByChecker|TestPrototypeHazardsBehindObjectViewsAreNotYet|TestUnrepresentedPrototypeCallsAreNotYet|TestIsPrototypeOfReadsExplainThePrototypeRefusal)$' -count=1 -timeout 90s -json
```

The baseline had an outer timeout 180; the restored run has an outer timeout 120. The mutation driver has timeout 300 and gives each Go command timeout 180. Its additional focused command:

```sh
go test ./internal/lower -run '^TestIsPrototypeOfReadsExplainThePrototypeRefusal$' -count=1 -timeout 90s -json
```

All test output went directly to logs. Baseline top-level elapsed seconds: inherited-member parent 0.34 (its parallel subtests run after the parent yields), methods 4.54, nullish 9.09, hazards 4.93, unrepresented 3.10, explanatory reads 1.39. Every individual leaf is under 60 seconds; maximum baseline leaf is 9.09s. No test leaf was added or touched. No whole package or full gate was run; setup itself performs its prescribed repository build cache warming.

Before push, commit then run the prescribed lane command under an outer timeout:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

This delivers the complete refusal census and scope evidence toward the acceptance-agreement roadmap work. The brief gives no numbered roadmap step for #41bkfdw, so no number is inferred. The acceptance conversion and M20 behavioral witness belong to a unit owning regex acceptance tests.

Delivery verification: review commit e8165410 was merged with current main 60397548. The target prototype tests and production files did not change in that main update. The restored focused run passed in 3.796s, longest leaf 1.40s; M20's run passed in 10.377s, longest leaf 3.92s. The merged-main focused run is preserved in merged.jsonl.

The first lane attempt printed `fatal: invalid object name 'origin/cloud/merge-tree'`; the requested fetch populated FETCH_HEAD but not that remote-tracking ref. The original pipeline's final Python process exited zero despite the missing script, so that attempt is not a successful check. Explicitly fetched `refs/heads/cloud/merge-tree:refs/remotes/origin/cloud/merge-tree` and the corresponding fast-gate ref, then reran with bash pipefail. Actual lane output: `lane checks 0.4 s: gofmt and tools on 0 Go files, t.Parallel on 0 test packages`, exit 0. This is a review-only diff, so there are no changed Go packages to vet.
