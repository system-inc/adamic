Built a production census keyed by all 173 ledger IDs, with one shared admission decision and all 67 optional contracts integrated.
Integration commits: ea4ecb4e, 117360ea, d7e67e20 and 9d324128; the final handoff commit also holds this report and the integration repairs.
The exact 79-root run reports 168 scheduled contracts, five named remaining errors and zero ordinary production errors; affected package and filtered oracle gates pass.
Two census mutants were caught; the integrated optional oracle also runs its construction, write, storage, collection and callback mutants.
Five catch use-site guards, whole-compiler emission and the complete repository gate are not claimed.

The input ledger is a1a16427a46149435e25f847c45ad136f1f1e55c,
`stage3/ledger/checker-259/rows-2026-10-08.csv`. `rows.csv` retains its actual
IDs, locations and codes. The 79 implementation roots in `roots.json` are the
ledger roots used by stricter-indexed-all. `source-hashes.json` records the exact
adapted input used here, from the existing stage 3 adapted tree.

`production.json.gz` contains the complete production decisions and diagnostics;
`states.csv` gives a compact, one-line-per-ID census:

| Kind | Scheduled contract | Remaining error |
| --- | ---: | ---: |
| Indexed presence | 99 | 0 |
| Optional presence / representation | 67 | 0 |
| JSON.stringify defined result | 2 | 0 |
| Catch use | 0 | 5 |
| Total | 168 | 5 |

The remaining IDs are D108, D128, D132, D152 and D153. Each retains its diagnostic
and a named unsupported-guard reason. The optional branch was integrated through
8c81fd11, including the final nested, collection and callback contracts. The catch
branch was integrated through a77b4c46: tagged thrown values and Error identity
are available, but its five use-site guards have not landed. A final fetch still
found that tip.

`Program.scheduleOptionSite` in `internal/load/option_disposition.go` is the sole
production admission decision. Normal Load and the census use it. The census
matches each decision to a manifest row using file, UTF-16 line/column and
checker code, then stores it under that row's original ID. IDs never authorize
admission. Duplicate IDs/sites, missing rows and additional production sites
fail closed. Ordinary diagnostics remain separate and still reject compilation.

For the whole-program worker:

```go
rows, err := load.ReadOptionLedger(input)
checked, report, err := load.LoadOptionLedger(paths, adaptedRoot, rows)
```

The report is available even when the remaining errors reject compilation. A
rejected Program is never exposed for emission. This is a complete diagnostic
census, not a bypass of pending guards. In particular, the 168 scheduled
contracts are not 168 emitted whole-program checks: whole-program emission is
still blocked by the five catch errors. The indexed-all worker can consume the
same ID-keyed decisions without reimplementing the loader's policy.

The standalone census command is:

```sh
go run ./stage3/stricter-options/whole-program \
  /tmp/stricter-options-adapted-current \
  stage3/stricter-options/whole-program/rows.csv \
  > /tmp/stricter-mode-whole-79.json \
  2> /tmp/stricter-mode-whole-79.log
```

It checks the entire 79-root program in one invocation, records every stricter
row, and exits successfully for a complete census with no ordinary errors. Its
stderr still names the five remaining checker errors. Native build admission
continues to reject those errors. The recorded counts are exactly 173 rows,
168 scheduled contracts, five remaining errors and zero ordinary errors.

`--explain-checks` now includes actual optional ObjectCall presence, spread,
storage, array and callback guards, and required-result coalesces. Pending
loader contracts remain uncounted. The mixed integration fixture proves the
actual IR report has indexed-presence=1, json-stringify-defined=1,
optional-write=1, catch-error=0 and trusted=0.

Integration repairs preserve the current project options and lib, the iterator
runtime-fact prelude, truthful JSON contracts and indexed guards. Explicit extra
roots retain their audit. Host console declarations retain their signatures;
the pinned Node console also retains its lowering declaration identity. A Node
runtime rebuild retains the Adamic and Set prelude roots. Filesystem errors and
catchable defined-read TypeErrors now set the tagged exception carrier and its
independent pending flag. Filesystem Error identity recognizes its real producer
shape rather than accepting arbitrary objects. Typed-array recognition terminates
on recursive generic intersections and remains ahead of record representation
selection. Unsupported typed-array spread retains its named refusal before
inferred record storage is considered. The imported filesystem buffer fixture
narrows JSON.stringify with `?? 'undefined'`, preserving Node's console output.

Commands were run with `/workspace/adamic-tools/env.sh` sourced, and all test
output went to log files. Final successful coverage:

- `go test ./internal/load ./internal/lower ./cmd/adamic ./stage3/stricter-options/... -count=1 -timeout 30m`: load 24.052s, lower 47.639s, CLI 4.758s. The representation-only run initially found the spread refusal ordering, repaired and rerun below. `/tmp/stricter-mode-final-core.log`.
- `go test ./internal/lower ./stage3/stricter-options/... -count=1 -timeout 30m`: lower 38.848s, representation fixtures 37.801s, pass. `/tmp/stricter-mode-representation-final.log`.
- `go test ./internal/ir ./internal/native` as part of the affected gate: IR 40.651s and native 288.078s, pass. Other packages in that initial gate found the loader and lowering integration issues recorded above and were rerun after repairs. `/tmp/stricter-mode-green-packages.log`.
- `go test ./cmd/adamic -run TestExplainChecksCountsIntegratedContracts -count=1`: 2.140s, pass; also included in the final CLI package run. `/tmp/stricter-mode-explain-test.log`.
- `go test ./internal/oracle -run 'TestOptional.*Guard|TestOptionalImplementsRepresentation|TestNativeAgreesWithNode/internal/oracle/testdata/catch_values.a$|TestNodeFSFileAgreesWithNode' -count=1 -timeout 30m`: 25.822s, pass. Source Node, native under sanitizers and the JavaScript backend agree. `/tmp/stricter-mode-final-oracle.log`.
- `go vet ./internal/load ./internal/lower ./internal/ir ./internal/native ./cmd/adamic ./stage3/stricter-options/...`: pass, `/tmp/stricter-mode-final-vet.log`.
- `gofmt -l cmd internal` and `git diff --check`: empty, pass.

Fresh census mutants:

| Mutation | Witness that caught it |
| --- | --- |
| Admit an unsupported stricter row as a scheduled check | TestOptionLedgerContinuesAfterRemainingErrors fails its census counts and named catch-error assertion |
| Replace manifest row IDs with D000 | TestReadOptionLedgerMultiline fails its original-ID assertion |

Both mutant compilers built; both tests failed at the intended assertions. The
compiler was restored before final gates. Logs are
`/tmp/stricter-mode-mutant-admit-unconverted.log`,
`/tmp/stricter-mode-mutant-replace-ledger-id.log` and
`/tmp/stricter-mode-mutants.log`.

The filtered optional oracle reruns the worker's mutants: lost own-presence on
direct writes, fresh construction and leading spread; early outer-scope cleanup
caught by leak/count observations; undefined required producer result; wrong
object storage; lost optional write presence observed by Node; missing nested
and nullable fields; missing array-element fields; wrong callback storage and
lost callback result presence. Guard failures require native exit 70 with the
named message; representation mutants are held to independent Node observations.
The original worker's detailed proof is `stage3/optional-writes/REPORT.md`.
Existing indexed, sparse, typed-array and JSON witness/erasure-mutant tests are
rerun by the representation package. Their detailed proof remains in
`stage3/stricter-options/REPORT.md`. No new runtime guard semantics were invented
by the census.

The complete repository gate and complete allocation-count census were not run.
No per-site indexed witnesses were added: those remain with the sliced workers.
No branch other than codex/stricter-options-checks is pushed, and no PR is opened.
