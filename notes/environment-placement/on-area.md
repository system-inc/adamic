# Environment placement on the current area

Base: runtime/area-take-batch at 6c891cca313e614ae46de794814bc7c0db4cae01.
Feature input: codex/environment-placement at a4ca1c6bb874f6f9b66e28481691a0b89fcda23a.
Coverage input: coverage/environment-placement at 066c2beef144ffa23bdd3635bc4788a9144b3c98.
Destination: runtime/envplace-on-area. No main push or forced push.

This is an adaptation of the existing allocation-site escape proof, not a redesign.
The area already lowers named nested functions to one AllocateEnvironment and one
FrameEnvironment layout. Placement now consults that statement's identity at
function entry, before captured parameters initialize their slots. No lowering
layout change was necessary. A frame stays counted if any retaining closure can
escape; at most 64 proven local cells use a fixed C aggregate, larger ones a region
owned by the call. Stores, returns, captures and keeping calls remain conservative.

Async functions keep the cells embedded in their counted scheduler frames. Their
allocation sites are excluded from placement; all async parameters are keeping,
and invoking an async carrier keeps its environment through frame->self. A carrier
assigned to an async local also escapes the entry call. Unknown void, promise or
union-returning carriers stay counted because they may use the async ABI. A known
synchronous, non-void result outside Promise and Union does not use that ABI.
Library callbacks check for async targets as well. The new environment_async.a
covers a synchronous parent's named async child and forwarding a nested closure
to an async function that reads it after suspension. Synthetic tests directly
assert counted async cells, carriers, callbacks and arguments.

GraphCell layouts retain their existing counted graph ownership. The environment's
cell vector is now a pointer for the common heap/stack/region representation, so
graph adoption repairs both that pointer and every interior cell's owner after
moving the record. The area's async initialize/drop helpers are retained and
shared by ordinary environments. Region allocation remains inline and region
teardown retains its holds_outside shortcut and weak-table walk; an environment
forces child cleanup because its slots can acquire counted values later.

The environment owner is held before its closures and locals, so reverse scope
cleanup releases carriers before destroying the slots they read. Stack cleanup and
region cleanup run on normal returns and exceptional exits. The imported fix makes
captured reference parameters own their entry reference independently of their
cell. The area allows borrowed cell stores; the legacy ownership mutant is now
rejected at parameter entry rather than by the scalar-store assertion. Other
borrowed cells keep the area's existing store rule.

Every oracle testdata/refusal path from all three inputs remains present. All 35
feature/coverage environment fixtures are registered, plus the new async fixture.
The carrier operand read is registered with the IR guard as lifetime evaluation;
callee target identity still comes exclusively from ClosureTargets.

Repository-wide vet exposed an unrelated stale volumeGenerated call in the area's
shipping profile test. Commit 91879f86 had moved those exact cases into generated.
The call now uses generated, as TestRulesAgree already does, preserving those cases.
This gate repair has its own commit.

Setup ran cloud/setup.sh --wasi-sdk with GOPROXY=https://proxy.golang.org|direct.
Every gate shell sourced /workspace/adamic-tools/env.sh. Node was v24.19.0,
Go go1.27.1 and clang 20.1.8, on Linux with a four-CPU quota.

## Counts

The complete table regenerated successfully: 713 to 749 rows, 36 additions.
Its regeneration commit changes only internal/oracle/counts.md.
Five existing rows reduce allocations:

| Fixture stem | Allocations | Frees | Peak live |
|---|---:|---:|---:|
| nested_captures | 22 -> 20 | 22 -> 20 | 7 -> 6 |
| nested_mutual | 10 -> 8 | 10 -> 8 | 5 -> 4 |
| nested_tdz | 2 -> 1 | 0 -> 0 | 2 -> 1 |
| nested_tdz_write | 2 -> 1 | 0 -> 0 | 2 -> 1 |
| nested_destructured_tdz | 2 -> 1 | 0 -> 0 | 2 -> 1 |

nested_mutual also reduces releases from 52 to 50. The three TDZ fixtures panic,
so their zero frees are counts at the panic, not finished-program leaks.
Twenty other existing rows gain equal retain/release pairs from captured parameter
ownership; their allocations, frees, peak and region counts are unchanged.

## Verification

All package runs used -count=1 and -timeout 30m. The runtime and oracle gates
used ADAMIC_GATE_UNCACHED=1. The oracle ran its ASan/UBSan malloc and slab variants,
release comparisons, leak checks, counted builds and ThreadSanitizer schedules.
Go-only analysis packages ran their complete existing tests; their compiled
witnesses use the repository's native sanitizer harness.

| Package | Passed top-level tests | Passed subtests | Skipped | Seconds |
|---|---:|---:|---:|---:|
| internal/native | 119 | 241 | 7 | 965.581 |
| internal/lower | 129 | 866 | 2 | 62.897 |
| internal/flow | 5 | 1985 | 0 | 150.377 |
| internal/fresh | 7 | 8 | 0 | 100.188 |
| internal/oracle | 94 | 2362 | 10 | 1105.850 |
| internal/ir | 4 | 0 | 0 | 10.883 |

Total: 5820 passed tests/subtests, 19 optional skips, zero final failures.
The skips are opt-in WASI, prebuilt release/profile and stage-3 hooks, historical
census/measurement hooks, and long-argument cases outside the available stack
limit. No required sanitizer or leak check was suppressed.

Commands:

- ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/lower ./internal/flow ./internal/fresh ./internal/ir -count=1 -timeout 30m -json
- ADAMIC_GATE_UNCACHED=1 go test ./internal/native -count=1 -timeout 30m -json
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -json
- go test ./internal/ir -count=1 -timeout 5m -json
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
- go vet ./...
- gofmt -l cmd internal stage1/cohere/lint/shipped_profile_test.go
- git diff --check

The initial combined package run completed lower, flow and fresh. Its IR guard
needed registration of the carrier operand evaluation and passed in the final
separate run. The old native process was stopped after starting the final-source
native rerun, which completed in full. The initial environment-focused run also
exposed the two legacy test adaptations described above; all final focused and
complete gates passed. The formatting, vet and diff checks produced no output.

TestEnvironmentEscapeStackMutant replaces the actual returned fixture's counted
record with call storage and changes its cleanup to explicit slot destruction.
It builds successfully, then ASan reports stack-use-after-return in the later
read closure. TestEnvironmentThrowReleaseMutant preserves output and passes the
ordinary sanitizer run but LeakSanitizer catches its missing exceptional cleanup.
TestEnvironmentCapturedParameterBorrowMutant recreates borrowed captured
parameters and is rejected at native parameter entry with the original named
parameter diagnostic. All three permanent mutants passed in the whole oracle.

Local evidence logs: /tmp/envplace-native-final.json, /tmp/envplace-packages.json,
/tmp/envplace-ir-final.json, /tmp/envplace-oracle.json, /tmp/envplace-mutants.json,
/tmp/envplace-async.json, /tmp/envplace-counts.log, /tmp/envplace-vet-final.log and
/tmp/envplace-format-final.log. Raw logs are not committed.
