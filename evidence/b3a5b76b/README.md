# Verification at b3a5b76b

Source head: `b3a5b76b91e61baea9db707dd121dc5664c473cc`, isolated worktree `/tmp/adamic-b3`.
Base: `ff41896cee93f832a4a8fa68c4c1822ddef5885e`, separate worktree `/tmp/adamic-main`.
Evidence destination: existing branch `evidence/compiler/generic-method-view`.
All compiler changes were temporary requested mutants, restored byte-for-byte. This evidence commit changes only evidence/b3a5b76b/.

| Step | Exit | Result |
|---|---:|---|
| 1 gofmt -l internal/lower/ | 0 | Empty output. |
| 1 go vet ./internal/lower/ | 0 | Passed. |
| 2 combined generic / NodeWatch pins | 1 | All 18 generic subtests passed. Three NodeWatch tests remain red. |
| 3 generic mutant, generic pins | 1 | All 14 rejection subtests red; four identical-view controls green. |
| 3 restoration / Git diff | 0 / 0 | Exact restoration. |
| 4 remove librarySymbol exception, TscShape | 1 | TscShape red, but also red in baseline; not a green-to-red proof. |
| 4 restoration / Git diff | 0 / 0 | Exact restoration. |
| 5 census head / base | 0 / 0 | Both pass with exact upstream TypeScript pin. |
| 5 whole lower, timeout 40m | 1 | Completed in 651.945s without timeout; only the same three NodeWatch tests failed. |
| 6 manifest | 0 | Generated at exact head. |
| 6 admission delta | 0 | Complete-corpus PASS, 4,957 programs, zero omissions, zero newly admitted, **three newly refused**. |

## Commands and logs

Go/clang/Node are selected by sourcing `/workspace/adamic-tools/env.sh`.
Every command's output and numeric exit are retained as named `.log` / `.exit` pairs. The two exact mutant patches are retained too.

```sh
gofmt -l internal/lower/
go vet ./internal/lower/
go test -count=1 -run 'TestAGenericMethod|TestNodeWatch' -v ./internal/lower/
# genericMethod return changed to continue:
go test -count=1 -run 'TestAGenericMethod' -v ./internal/lower/
# restore; remove " || l.librarySymbol(inside)":
go test -count=1 -run '^TestNodeWatchTscShape$' -v ./internal/lower/
# restore; census run at both revisions:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript-source go test -count=1 -timeout 40m -run '^TestProgramRegionCensusMembership$' -v ./internal/lower/
# same source pin exported for full lower:
go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/
python3 cloud/admission-corpus/manifest.py --sha HEAD > /workspace/adamic/evidence/b3a5b76b/manifest.json
go run ./cmd/adamic-admission-delta --base ff41896cee93f832a4a8fa68c4c1822ddef5885e --head b3a5b76b91e61baea9db707dd121dc5664c473cc --manifest /workspace/adamic/evidence/b3a5b76b/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

## Generic mutant red rows

Each fails for both `.a` and `.ts`:

- g1, T unused on both sides
- g4, T only in a callback's type
- g5, generic on the source side only
- g6, generic on the target side only
- g7, an interface seen as another
- o2, an overloaded generic view
- one class's generic method under two type arguments

The Box row gets the nominal-ancestry refusal instead of the required generic-method NotYet; the other rows lower the bad view. Exact names/messages are in generic-mutant.log and generic-mutant-red-rows.log.

## NodeWatch and full lower failures

Baseline pins and whole lower both fail:

- TestNodeWatchInheritedOnMutant
- TestNodeWatchInheritedOnStaysNotYet, subtest view
- TestNodeWatchTscShape

All report the generic `on` view NotYet, for ChildProcess or FSWatcher. Full messages and every `--- FAIL` are in lower.log, lower-failures.log and lower-failure-index.log. Removing the new exception also fails TscShape; because the unmutated test already fails, this mutant cannot validate the exception. [Read-only observations](observations.md) explain the likely library-predicate mismatch; no fix was attempted.

## Census environment

`stage3/source.json` requires upstream TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, not cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. The exact upstream v6.0.3 pin was cloned to /tmp/adamic-typescript-source. Both base and head census tests passed with ADAMIC_TYPESCRIPT_SOURCE set to it. Thus the earlier missing-source failure was environmental. The compiler-reading differences in the verbose logs are informational; the indexed/pairwise test passes.

## Admission

[All three new refusals with full diagnostics](new-refusals.md): host_watch/tsc_shape.a, host_watch/view_after_close.a, host_watch/view_other_event.a. All remain generic `on` view stops, contrary to the expected zero refusals. None is an override refusal or constrained-string handle view.

The final JSON contains the complete result. The gate passes because new refusals are informational. No omitted programs, no newly admitted programs. Runtime checks retain 50 accepted negative witnesses and 37 ruled runtime stops.

As in the prior successful run, npm provisioning used temporary ~/.npmrc `cafile=/etc/ssl/certs/ca-certificates.crt`, with TLS verification enabled. The prior configuration was restored. No tool/compiler code was changed.
