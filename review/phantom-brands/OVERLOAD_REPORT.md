Built: module overloads admit phantom-only result differences with erased brand introduction.
Commit: this feature commit, based on a73f93c542ca79ca119a813cef24229f8180e745, on codex/phantom-brands only.
Checks: lower suite, affected Node differential oracle, counts, vet and formatting pass.
Mutants: eight executed, all caught by assertion failures, restored tests pass.
Uncovered: generic overloads, overloaded function values, methods/nested overloads, full repository gate.

The Path fixture keeps ensureTrailingDirectorySeparator's original implementation body and adapts the TypeScript any brand to void as ruled. It calls the Path and string overloads, and a generic-free variant whose two overloads both return Path. Both backends match original-source Node byte for byte, with native ASan/UBSan, release and leak checks through the oracle. It also exercises a safe number|string result widening, number/string argument dispatch and an existing required-undefined array brand result. Complete IR equality against the unbranded implementation pins erased overload headers and zero runtime brand operations.

This base had no module overload support: bodyless headers failed registration. Small hooks register the implementation once, prove each public parameter compatible with the implementation including writable slots/ownership, check every overload result even if unused, and fit direct calls to the real implementation representation. Phantom result compatibility projects away only accepted brand members and checks both directions, preserving literal, element and readonly constraints. Ordinary result widening still uses the existing proven relation. Generic overloads and reading overloaded functions as values remain explicit NotYet pending separate proofs.

The corrected negative pin is overload result 'a' against implementation result string. Its exact What is `an overload result "a" not proven by its implementation result string`; Fix is `return a type proven by the implementation, or use only an accepted phantom brand difference (adamic/overload-results)`. A brand over literal 'a' also stays refused. The any-member declaration pin has exact What `a primitive brand member __pathBrand whose type is not void`; Fix `make __pathBrand void (or optional and typed undefined) so the brand is phantom`. The review directory contains the literal counterexample and any-brand ensureTrailingDirectorySeparator source. No any brand was admitted or census source adapted in this unit.

Original-source Node stdout (stderr empty, exit 0):

```text
root1/ already/
plain2/
undefined
root1/
second3/
string wide4
5/
dispatch6/
2 7,8 undefined
```

Executed mutants and catches:

| Mutant | Assertion that failed |
| --- | --- |
| Accept a non-brand overload difference | TestPhantomOverloadLiteralResultRefused |
| Skip the ordinary result proof | TestPhantomOverloadLiteralResultRefused |
| Accept any branded result, losing literal constraints | TestPhantomOverloadBrandLiteralConstraintRefused |
| Keep a runtime Defined operation on brand introduction | TestPhantomOverloadResultCastsAreErased |
| Skip overload parameter proof | TestPhantomOverloadParameterProof |
| Accept an overloaded function as a value | TestPhantomOverloadFunctionValueNotYet |
| Register an overload header without its body | Native Node differential fixture |
| Accept a brand member typed any | TestPhantomOverloadAnyBrandRefused |

Every mutant exited 1 with a test assertion failure, not a build warning/error. The runner restores each production file in finally. Results, runner and individual logs are in overload-evidence.

Validation commands and outputs:

```text
go test ./internal/lower -count=1: ok 15.259s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPhantom|TestNativeAgreesWithNode/internal/oracle/testdata/(phantom_|native-sorted-array-brand.a|generic_|proven_|fresh_|invariance|weak_)' -count=1: ok 4.048s
After restoring mutants:
go test ./internal/lower -run 'TestPhantom|TestWhatStageZeroCannotLower' -count=1: ok 2.148s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPhantom|TestNativeAgreesWithNode/internal/oracle/testdata/phantom_' -count=1: ok 1.107s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts: ok 18.015s
go vet ./...: exit 0, empty output
gofmt -l cmd internal: empty output
git diff --check: clean
node --input-type=commonjs-typescript < internal/oracle/testdata/phantom_overload_results.a: exit 0
```

The new fixture counts allocations/frees/retains/releases/peak/regions are 29/29/16/51/7/0. No other count changed. All test output was logged.

Census: the unchanged archival adapted 78-root corpus and exact unique (kind,where,reason,text) method were reused. Before is the preserved a73f93c5 measurement from the prior unit, not the historical 222 total. After was rerun with a new scratch overlay/binary; production loader rejection remains enabled. LATENT_ASSERT_NO_OUTPUT=1 and the census audit pass, including the signature/body range misattribution mutant. Both configurations are measured on a checker-rejected program; these are latent findings, not successfully compiled tsc. Source hashes and exact diagnostic filters are preserved in overload-evidence/census-summary.json.

| Exact measurement | Before | After |
| --- | ---: | ---: |
| All unique findings | 3731 | 3729 |
| Historical broad 222-site diagnostic filter | 175 | 169 |
| Eight exact primitive brand diagnostics | 45 | 45 |
| Function without a body | 192 | 191 |

The historical broad filter contains unrelated named/intersection types; its six-site movement is not six proven phantom-brand implementations. New overload result refusals include original any-typed Path results, generic overload proofs and literal/structural narrowing. The full per-diagnostic delta is preserved. The positive adapted Path fixture moves from one bodyless-header refusal to zero.

Scratch commands: stage3/census/latent/make_overlay.py into /tmp/phantom-overload-after-overlay; go build -buildvcs=false -overlay=/tmp/phantom-overload-after-overlay/overlay.json ./stage3/census/latent/tool; LATENT_ASSERT_NO_OUTPUT=1 census binary /tmp/phantom-tsc/src/compiler /tmp/phantom-overload-after.jsonl; stage3/census/latent/audit.py on that binary. No census hooks were committed.

Toolchain setup succeeded: Node 0.025s, Go 0.027s, submodules 0.073s, markdown 0.084s (validation 0.012s), clang 0.190s, build 25.661s, deferred tests 25.827s, cache warm 25.830s, total 25.860s. nproc=5, cgroup CPU quota=4; Node24.19.0, Go1.27.1, clang20.1.8. Setup output is retained.

The latest explicit unit selected a73f93c5 as the base. Current origin/main was fetched (48c05d091f0a43c31cbe051b1d6578d99eeedf19) and inspected; it also lacked this overload guard. It was not merged over the explicitly requested base. Recursive optional submodule fetching was stopped after the main ref fetch completed. No push or merge into main or any area branch occurred, and no cohere code was copied. Only codex/phantom-brands is the push destination.
