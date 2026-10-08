# Scanner cast acceptance measurement

Acceptance inputs copied from ea1b2359 without merging its branch. Compiler base 561166d1, checked-views integration 432d4913. The ledger was fetched as an immutable object. Error remains pinned to library #ddwcejg; no library facts are fabricated.

Fresh baseline: all 20 Node goldens pass. Five of 20 runtime contracts pass: 01 positive and failing, 02 positive and failing, and 03 positive. Of the 18 contracts blocked on the acceptance-author branch, three pass here. 03 failing stops 70 at the specified read with the right stdout but uses the existing `field read failed:` diagnostic instead of the required `cast failed:` prefix. 04 positive and failing refuse the backing string/number field conversion during the view scan. 05 through 10 positive and failing refuse unchecked casts.

The supplied observer aborts at 01 failing because it only recognizes `adamicCast`; this compiler emits `adamicCheckedViewCast`. Its strict verifier was run unchanged and exits 1, but reads the historical observations, so its two-pass summary is not our baseline. worker-evidence/baseline-observations.json is a fresh independent audit with compiler/backend/source receipts; audit execution disables implementation mutants, not runtime checks. Its fixture hashes refer to the input fixtures before removing the now-obsolete 01 refusal headers.

Tagged contract group: TestStep09AcceptanceTagged holds all four tagged fixtures to Node, sanitized native, native release, and JavaScript, pins exact stop diagnostics, and checks leaks for successful executions. Its real IR mutants remove one cast-point check in each failing twin. Both backends still stop on a later checked read, but with a different diagnostic; the exact cast-point pin catches each omission. No unchecked-success mutant is claimed. No compiler change was needed for this group.

Commands (all output saved to logs):
- GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh: pass; setup 49.190s; nproc 5, CPU quota 4.
- go build -o /tmp/step09-worker-adamic ./cmd/adamic: pass.
- NODE_PATH=stage3/api/node_modules node stage3/fixtures/checked-casts/observe.cjs /tmp/step09-worker-adamic /tmp/step09-worker-baseline: exits 1, missing emitted-check mutant at 01 failing.
- node stage3/fixtures/checked-casts/verify.cjs --require-runtime: exits 1 on historical 18 blockers.
- node /tmp/step09-baseline-audit.cjs /tmp/step09-worker-adamic /tmp/step09-worker-audit: pass; fresh five contracts pass, fourteen lowering blockers, one runtime diagnostic mismatch.
- go test ./internal/oracle -run 'TestStep09AcceptanceTagged|TestScannerCastCounts' -v -args -update-counts: pass, 3.630s; four new canonical counts rows.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: fails on inherited fixtures, 50.949s; the focused refresh records this group's rows.

No full package suite or full gate was run. Untagged, finite-domain, generic and primitive contracts remain unfinished; the strict acceptance requirement is not green. The historical a-check and native-mutant receipts are not fresh worker validation.
