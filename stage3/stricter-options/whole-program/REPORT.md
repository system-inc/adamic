Integrated tagged catch values and typed-use guards into the production ID-keyed census for all 173 ledger rows.
Merged catch worker ded85b2d into codex/stricter-options-checks; the integration commit contains this report.
The unchanged 79-root adapted input reports 173 scheduled contracts / 0 remaining errors / 0 ordinary errors.
Two scheduler mutants and five executable catch mutants are caught; affected packages and catch oracles pass.
Scheduled contracts are lowering obligations; whole-compiler emission and the complete repository gate are not claimed.

The pinned input ledger remains a1a16427a46149435e25f847c45ad136f1f1e55c,
`stage3/ledger/checker-259/rows-2026-10-08.csv`. All 79 source hashes in
`source-hashes.json` remain unchanged. `production.json.gz` contains the complete
production decisions; `states.csv` preserves every original ledger ID and location.

| Kind | Scheduled | Remaining errors |
| --- | ---: | ---: |
| Indexed presence | 99 | 0 |
| Optional presence / representation | 67 | 0 |
| JSON.stringify defined result | 2 | 0 |
| Caught type / representation | 5 | 0 |
| Total | 173 | 0 |

D108, D128, D132, D152 and D153 moved from remaining-error to scheduled-check,
kind caught-type. The shared `Program.scheduleOptionSite` decides admission for
normal compilation and the census. No second bypass was added to Load. Identity
matching still uses file, UTF-16 line/column and diagnostic code, then retains the
original manifest ID. Missing, additional and duplicate identities fail closed.
Ordinary diagnostics remain errors and never become scheduled checks.

The catch representation preserves actual thrown values and the project's declared
checker type. A message/code read preserves its actual value, including undefined;
a definite typed use checks that value. Three of the five ledger patterns need
primitive guards; D152's code comparison and D153's returned record require truthful
representation. Thus five scheduled rows do not claim five emitted runtime guards.
The ruling and runtime limits are in `docs/caught-values.md`, and the worker's
site mapping is in `cloud/stricter-catch-report.md`.

Merge reconciliation preserves optional contracts, sparse and typed-array support,
filesystem Error identity, project options/lib and ordinary unknown reflection
refusals. Catch-property emission activates dynamic shape metadata, distinguishing
numeric and boolean slots. Nullable message/code data slots can enter project catch
storage when their null tag is represented. Fields admitting both null and undefined,
other nullable fields and broader nullable reflection keep their named refusals.
`--explain-checks` now includes actual caught-type guards in its summary. Its mixed
fixture reports indexed-presence=1, json-stringify-defined=1, optional-write=1,
caught-type=1, catch-error=0 and trusted=0. Pending loader contracts remain uncounted.

The exact census command, with the toolchain environment sourced, is:

```sh
go run ./stage3/stricter-options/whole-program \
  /tmp/stricter-options-adapted-current \
  stage3/stricter-options/whole-program/rows.csv \
  > /tmp/stricter-catch-census-final.json \
  2> /tmp/stricter-catch-census-final.log
```

The whole-program worker can merge the reported integration SHA and use the same
`load.LoadOptionLedger(paths, adaptedRoot, manifest)` API. A complete census is
available even if a future unsupported option remains an error; rejected Programs
are never exposed for emission. All 173 current contracts are scheduled, but
whole-program lowering may still refuse unrelated unsupported representations.

Tests and observations (all output saved to logs):

- Affected packages: `go test ./internal/load ./internal/lower ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh ./cmd/adamic ./stage3/stricter-options/... -count=1 -timeout 20m`, pass. Load 39.925s, lower 215.741s, IR 25.172s, flow 341.834s, fresh 222.852s, CLI 44.759s, stricter-options 159.630s; JavaScript and census command have no package tests. `/tmp/stricter-catch-packages.log`.
- Final admission and mixed explain counts: `go test ./internal/load ./cmd/adamic -count=1 -timeout 10m`, pass (48.860s and 50.830s). `/tmp/stricter-catch-admission-explain.log`.
- Reconciled catch oracle: `go test ./internal/oracle -run 'TestCaught|TestNativeAgreesWithNode/internal/oracle/testdata/(caught_|catch_values.a)' -count=1 -v -timeout 20m`, pass in 15.003s. Source Node, JavaScript backend and sanitized native builds agree; definite string failures stop at the named check with exit 70. `/tmp/stricter-catch-oracle-reconciled.log`.
- Filesystem catch regressions: every `TestNodeFSFileAgreesWithNode` case passed in the initial oracle run. That run also found the nullable-view merge blocker, subsequently repaired and rerun above. `/tmp/stricter-catch-oracle.log` is an unsuccessful overall run and is not claimed green.
- Final lowering/refusal gate: `go test ./internal/lower -count=1 -timeout 20m`, pass in 69.433s; `/tmp/stricter-catch-lower-final.log`. Additional project null-tag refusals pass in 0.357s; `/tmp/stricter-catch-nullable-refusals-final.log`. The first additional witness met an unrelated template interpolation refusal; changing it to a typeof observation isolates the intended unsupported-tag assertion.
- `go vet ./internal/load ./internal/lower ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./cmd/adamic ./stage3/stricter-options/...`, pass; `/tmp/stricter-catch-vet.log`.
- Complete allocation-count regeneration: `go test ./internal/oracle -run '^TestCountsAreRecorded$' -parallel 4 -count=1 -timeout 20m -args -update-counts`, pass in 87.889s; `/tmp/stricter-catch-counts-final.log`. The table now has 803 rows: 11 additions (five catch fixtures and six existing library overlay fixtures missing from the previous table), zero removals and 73 changed existing rows. 72 changes affect only retains/releases; collections.a also decreases allocations and frees from 363 to 362. The initial regeneration found the nullable-view blocker and did not update the table; its unsuccessful log is `/tmp/stricter-catch-counts.log`.
- Final whole-program census exits 0, stderr is empty, all 173 original identities match and ordinary errors are zero. The saved production artifact equals the final run exactly after path normalization. `/tmp/stricter-catch-census-final.json` and `/tmp/stricter-catch-census-final.log`.
- Native field and release checks: `go test ./internal/native -run 'TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestOptionalWriteReservedSlotMatchesNode|TestRuntimeReleasePaths' -count=1 -timeout 10m`, pass in 6.793s; `/tmp/stricter-catch-native-final.log`.
- `gofmt -l cmd internal` is empty and `git diff --check` passes.

| Mutation actually run | What caught it |
| --- | --- |
| Restore catch rows to remaining-error | TestOptionLedgerContinuesAfterOrdinaryErrors fails the 4/0 census assertion; ordinary errors remain separate |
| Admit an unsupported stricter option | TestUnsupportedOptionContractRemainsError fails its named remaining-error assertion |
| Every caught value is Error | Source Node stdout disagrees in native and JavaScript |
| Claim every member read is a definite string | Native and JavaScript stop early with exit 70 while Node preserves undefined |
| Erase the definite-string use guard | JavaScript continues with undefined; native UBSan detects the null string access; both violate the pinned exit-70 contract |
| Clear maybeMessage.MayThrow | Native loses the outer caught TypeError output after finally; Node comparison catches it |
| Drop the null tag | Sanitized native finishes but stdout differs from Node |

Scheduler mutants compiled, failed at their intended assertions, and were restored
before final verification. Logs are `/tmp/stricter-catch-mutant-catch-admission.log`,
`/tmp/stricter-catch-mutant-unsupported-admission.log` and
`/tmp/stricter-catch-scheduler-mutants.log`. The five runtime mutants are executable
tests in `internal/oracle/catch_values_test.go` and passed in the reconciled oracle.
Prior indexed, holes, typed-array and JSON guard witnesses and erasure mutants were
rerun by the stricter-options package. Their detailed proof remains in
`stage3/stricter-options/REPORT.md`; optional contracts remain documented in
`stage3/optional-writes/REPORT.md`.

No new per-site indexed witnesses were authored. The complete repository gate,
test262 corpus and whole-compiler native emission were not run. Only
codex/stricter-options-checks is pushed; no PR is opened.
