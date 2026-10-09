Built: three regex acceptance rows in internal/lower/regexp_test.go, each held to source Node through the agreement helper and sanitized native behavior; preserved all 345 prototype refusal rows.
Commits: previous delivery 402f3310; merged current main 12e77e89 for this addition; final delivery SHA is reported with the push.
Commands and outputs: focused regex baseline passed in 5.739s; restored focus, TestCallTargetReaders, and lane results are recorded below and in their logs.
Mutants: audit M20 fails the new Unicode row on native stdout; one-line overescaping fails all three new rows; both restored.
Not covered: production changes, new oracle fixtures, or general regex coverage beyond these witnesses. JavaScript recomputes source metadata and cannot alone kill M20.

## Initial census and verification

The following preserves the original refusal-only scope and its evidence. The authorized regex addition below extends that scope and supplies the previously missing M20 behavioral guard.


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

Additional requested guard: `timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s` initially failed in 17.970s on the existing `internal/lower/class_static_guard_test.go:TestClassStaticInitializerCallIsEmitted:Call.Function` reader. Fetched and merged current main cf79ecec, which contains the compiler-owned allowlist entry with reason `names the function under test`. No local allowlist entry was invented. Reran the exact guard: pass, 0.805s. Before/after logs are call-target-readers-before-main.log and call-target-readers.log. Repeated the six focused unit tests on that main; final-main.log records the result. Lane checks were repeated before updating the pushed branch.

## Authorized regex acceptance addition

Added three separate parallel top-level leaves beside the existing regex lowering tests in internal/lower/regexp_test.go. Each program constructs a regex with flag u, prints source and flags and a test result on slash-containing input, and calls lowersAndAgreesWithNode. Cases: `[[/]/` on `//`, `[/]/` on `//`, and an escaped slash on `/`. rows.md records the three acceptance rows in addition to the unchanged prototype census. No IR assertions or snapshots were added. No checked-in oracle fixture was added, so counts.md needs no new row.

Observation: the first JavaScript-only version of the three rows passed under M20 (regex-M20-javascript-only.log). JavaScript emission builds `new RegExp` from the original arguments or pattern and flags, and never uses RegExpNew.Source; JavaScript recomputes its metadata. Therefore the requested helper alone cannot expose the lowering error. Added a sanitized native run of the same lowered program against an independent source Node observation. This compares stdout, requires successful exit, and rejects stderr, including sanitizer reports. The test helper's comment explains why both backends are necessary. There are no hard-coded expected regex strings in test assertions.

Applied the authoritative M20 patch from e7d2a4ed, preserved as regex-M20.patch. The new TestRegExpUnicodeClassSourceAgreesWithNode fails with native stdout `[[/]/|u|true` versus Node `[[/]\/|u|true`, which has one backslash before the final slash. It exits 1, and the other two rows pass. The one-line own mutant changes the slash escape condition from `!escaped && depth == 0` to `depth <= 1`; all three new rows fail native stdout comparison. Native processes exit successfully, with empty stderr; neither mutant is a build or sanitizer kill. The driver verifies the intended failures and restores regexp.go byte for byte in finally blocks. Logs and patches are regex-M20.log, regex-overescape.log, regex-overescape.patch, and regex-mutant-driver.log.

Exact focused commands, each with an outer hard limit and output sent straight to a log:

```sh
timeout 120 go test ./internal/lower -run '^TestRegExp' -count=1 -timeout 90s -v
timeout 300 python3 review/compiler/agree-prototype/run-regex-mutants.py
timeout 120 go test ./internal/lower -run '^(TestRegExp.*|TestInheritedLibraryReadsNeverLoadOwnFields|TestPrototypeMethodsAreRefusedWithReasons|TestNullishPrototypeReadsAreRejectedByChecker|TestPrototypeHazardsBehindObjectViewsAreNotYet|TestUnrepresentedPrototypeCallsAreNotYet|TestIsPrototypeOfReadsExplainThePrototypeRefusal)$' -count=1 -timeout 90s -json
timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
```

The mutant driver runs only the three new leaves, with -count=1, -timeout 90s, -v and an outer timeout 120 per invocation. Baseline new leaf times, including the agreement calls and native build: Unicode class 5.72s, slash class 5.73s, escaped slash 5.73s. M20's failing leaf is 0.39s; the overescape leaves are 0.40s, 0.41s, and 0.41s. Native runtime build caching explains the lower later costs; no performance improvement is inferred. No new leaf approaches 60 seconds.

Merged current main 12e77e89 before final verification. The main update adds a native test and review evidence; the regex lowering production source is unchanged. The new tests read none of Call.Function, CallClosure.Closure, or ArraySort.Comparator, and add no allowlist entry. Final guard output is in regex-call-target-readers.log. Final focused clean output is in regex-restored.jsonl. No whole package or full gate was run.
