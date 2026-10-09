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
