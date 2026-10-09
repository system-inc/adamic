# Serial lowering chain

Base: origin/cloud/land-stack-lowering-b-2ae8eb95, 2ae8eb953f8ea20d2c4a600bef9d0f81408f54da.
Seven requested branches are merged serially at their exact supplied SHAs. The checked-any branch is compiler/checked-any, confirmed by git ls-remote.
Each green member is committed, lane-checked from the repository root and pushed before the next member.
Every control log is direct test output. JSON results record named tests and seconds; each count ledger records every addition and changed existing row with its reason. No authored test leaf is added here.

Tool setup: Go 0.024s, Node 0.026s, submodules 0.067s, markdown 0.089s, clang 0.176s, build 73.925s, cache 74.084s, done 74.120s. nproc=5, cgroup quota=4 CPUs. GOPROXY uses proxy.golang.org with direct fallback. Setup passes.

## Member 1: compiler/c-portability-main ebad12cf

Only conflict: internal/oracle/counts.md. Regenerated once on Linux, preserving both parents' rows; one row added, no existing row changed.
Clang/GCC signed-char C11 sanitizer and Node controls pass. The standard uncached fixture also agrees in both backends. All three retained GCC warning mutants are caught without -Werror. TestConstantPortabilityClang/GCC, TestLongStringBytesMutant, TestStaticNumberInitializerMutant, TestStaticBooleanInitializerMutant, TestLongUnicodeLiteralByteCoverage and TestLongUnicodePortabilityClang/GCC are recorded in member-1-results.json.
Named native checks pass in 4.198 s and the oracle in 13.361 s; the longest named leaf is 4.15 s. Counts regeneration passes in 72.440 s (the existing full counter sweep is unchanged).
The merge's integration lane output and prefix push are recorded in member-1-lane.log and member-1-push.log.


Member-1 compiler prefix: 4e04e0f232b48954262b9aa986187776c4a62e93.
Lane output: lane checks 2.1 s: gofmt and tools on 47 Go files, t.Parallel on 8 test packages; vet 8 packages.
The source candidate's historical portability report/archive is preserved under source-member-1, following the evidence layout rule.

## Member 2: compiler/exceptions-21-main 205586a0, stopped

Only textual conflict: internal/oracle/counts.md. Both compiler sources otherwise merge automatically.
Named exception controls, sixteen registered exception fixtures and five semantic handler IR mutants pass against Node in both backends with sanitizers. The named cold-run leaves TestStep21ReadinessIsTerminal and TestStep21TypeScriptExtension take 12.96 s and 13.03 s; all named exception leaves stay under 60 s. Detailed events are in member-2-tests.jsonl.
A prior-step18 spot-check exposes the source runner's shift from global exit-70 normalization to language exit 1. Two test assertions were adapted temporarily to raw Node exit/stdout/TypeError text, and those controls pass in 0.437 s; no backend stop or refusal assertion was relaxed. The patch is preserved as evidence only and is not part of the green prefix.

The required once-only counts regeneration fails in 51.817 s on 81 existing fixtures. Every compiler diagnostic is the same representation interaction (plus clang's error limit): the earlier optional-chain checkDefined path assigns adamic_error_new's adamic_object * into adamic_thrown and accesses adamic_thrown->slots; step 21 changes adamic_thrown to adamic_heap *. Clang reports incompatible pointer types and no slots member. Member-2-blocker.json lists all 81 affected fixture paths. No generated count changes are accepted.

Small reproducer: internal/oracle/testdata/047cb0d_n_arrayindex.a, preserved byte for byte in member-2-reproducer.a.txt. It narrows an array, calls a function that sets it undefined, then reads list[0]. The generated required-read error is the path that does not compile. This is a compiler rejection, not an observed wrong-output exit-0 run.

Following the explicit chain stop rule, the member-2 merge is aborted. The temporary test assertions are restored. The green compiler is identical to member-1's committed prefix. No exception merge commit exists. Members 3 through 9 are not attempted: generic-body-relations 2cc109e5, generics-scout-main 48391dcf, iteration-main 518eca83, namespace-value 26ccff9c, compiler/checked-any 57ea02f8, project-references-main e2aa3750 and refusal-rulings-main 8799f518. The checked-any and two added branch SHAs were verified by fetch/ls-remote.

The final evidence-only commit preserves this report, raw test events, the failed once-only count log and prefix verification. Integration lane checks are run on that committed tip before pushing. No new Go test is authored; the attempted two test edits are absent from the delivery. Full repository packages/gate, WASI and stopped members' later validation are not claimed run.


## Additional queued members

Member 10: compiler/assignment-proofs-main a6f05066, a plain merge with the source's six fixture count rows.
Member 11: no branch merge. Apply only 4e973695's internal/native/tsgo.go diff using git show 4e973695 -- internal/native/tsgo.go | git apply -3, then compare against the current main build path and focused typeaware tests and measure seconds before/after. Skip if that reuse no longer applies. These members follow member 9 and are not attempted because the chain stopped at member 2. No typeaware patch, applicability finding or timing improvement is claimed.

## Member 3 resumed: compiler/generic-body-relations 2cc109e5

No textual conflicts. The enum refusal stays Refused and only its diagnostic changes, preserving all Node observations. Both new fixtures agree with Node in both backends, uncached with native sanitizers. All three production overlay mutants are caught by semantic assertions. Counts regenerated once: two additions, no existing row changed, 56.758 seconds. The opt-in census is skipped without its config; added t.Parallel because its explicit output path is supplied by its caller, not shared test state. Historical evidence moved to review and document links updated.
