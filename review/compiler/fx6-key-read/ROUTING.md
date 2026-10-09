Built one checked member boundary for dot access, finite element keys, literal-typed keys and object destructuring.
The refusal was pushed separately as 339c86b3; the routing commit and final main-merge SHA are reported in the final response.
Full lowering passed in 83.214s; selected oracle views/review probes and counts passed in 70.032s; reader guard and vet passed.
Two valid runtime mutants bypassed element and destructured-union checks; both were caught because exit 0 replaced the ruled exit 70.
Spread, in, Object.keys, Object.values and Object.entries remain unrouted; p19 and p72 are not separately activated.

This advances task #1fk58py, item 132 and the amended P0 view-syntax ruling. The separately pushed first commit protects literal keys by refusal. This phase removes that refusal where finite object member reads now preserve the same checks as dot reads.

`readViewMember` owns the read's View, ViewType, ViewTypeID, ViewWhere, receiver type, semantic contract, allowed literals and readiness metadata. It also preserves the existing optional-slot handling. Dot reads call it through the existing `readObjectField` wrapper. Finite keys supply the selected field symbol and receiver type directly, after holding the receiver and key in their existing evaluation order. Object destructuring supplies the declared field and receiver certificates to the same helper. Tuple destructuring keeps its previous path and is outside this phase.

The p01, p05, p06 and p07 originals now run as active agreement review fixtures with intended/expected sidecars for the P0 ruled stop. Their exact source Node behavior is observed and must differ from the pin, as the existing review harness requires. The compiler backends must print no output and exit 70 with the pinned field diagnostic. Correct controls retain bracket syntax and run through `lowersAndAgreesWithNode`. Literal-typed keys and destructuring have their own controls and ruled-stop tests. The destructured union fixture proves the membership contract, not just a primitive-kind check. Integration's original p49 destructuring probe receives the same ruled-stop sidecars.

The focused command was:

```text
timeout 180 go test ./internal/lower ./internal/oracle -run 'TestView.*Element|TestCheckedView(Element|LiteralTypedKey|Destructured)' -count=1 -v -timeout 90s
```

All sixteen leaves passed. Their observed seconds are in routed-final-tests.log; the slowest was 1.29s. The new wrong-view leaves are separate top-level tests with t.Parallel first. The first oracle trial used relative fixture paths, which the cache rejects; the corrected helper uses filepath.Abs and all subsequent runs passed.

The final pre-merge validation commands wrote to logs:

```text
timeout 300 go test ./internal/lower -count=1 -timeout 240s
ok github.com/system-inc/adamic/internal/lower 83.214s

timeout 480 go test ./internal/oracle -run 'TestCountsAreRecorded|TestCheckedView(Element|LiteralTypedKey|Destructured|Objects|Interfaces)|TestReviewProgramsAgreeWithNode/fxspptb_oct9_views_p(0[1567]|49)_' -count=1 -timeout 6m -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 70.032s

timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
ok github.com/system-inc/adamic/internal/ir 34.875s

timeout 180 go vet ./internal/lower ./internal/oracle
exit 0, no diagnostics

git diff --check
exit 0, no diagnostics
```

counts.md was refreshed and has no diff. These new fixtures are lowering/refusal/ruled-stop fixtures, not new counted execution rows. Neither full native/oracle packages nor the full repository gate were run.

Run `timeout 400 python3 review/compiler/fx6-key-read/run-routed-mutants.py` after sourcing `/workspace/adamic-tools/env.sh`. Each mutation is independent and restored in a finally block. The reverse source diffs are saved under this directory.

- element-bypass.diff replaces the shared helper call with the old bare Property. TestCheckedViewElementP05 fails: native and release print a tiny run-specific number then true, while unchecked JavaScript prints NaN then false. All three exit 0 instead of the required checked exit 70. This reproduces the native disagreement with source Node shown by the first-phase exact-p05 witness.
- destructure-bypass.diff replaces the shared helper call with the incomplete Property. TestCheckedViewDestructuredUnion fails: all three backends print true and exit 0 instead of stopping at the union membership check with exit 70.

Both final mutants were caught by runtime comparisons, with valid generated C and no sanitizer report. The runner initially looked for a stdout comparison label, but the oracle compares exit codes first; that assertion was corrected. An initial destructuring mutant left fieldSymbol unused and was rejected as a build failure. It was rerun with a blank use so the final mutant could only be caught by the runtime check. The final two mutants both returned test exit 1 and the runner returned 0.

The first-phase runner and REPORT.md are historical evidence for commit 339c86b3. The current implementation uses the routed runner. Raw logs are explicitly archived in this routing commit because the repository ignores *.log by default.

The fetched main tip for final delivery is fd8cd003554c7fd9a59bd0e81f4e5ed9361a9e60. Its intervening changes concern stage-1 tests and review fixtures, not the modified lowering code. Pending sidecars for the four now-active original probes stay removed when merging main. Post-merge checks and lane-check output are reported in the final response.

Not covered: a shared checked enumeration boundary for object spread, in, Object.keys, Object.values and Object.entries; tuple/destructuring consumers outside object member bindings; and separate activation of the callable-union p19 and mixed-slot p72 probes. Their remaining pending state is not claimed as fixed. No backend emitter or protected orchestration file was edited, and no cohere source was copied.
