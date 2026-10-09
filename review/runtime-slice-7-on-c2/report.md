# Slice 7 on c2: stopped at refusal-policy boundary

Base: 9674ce3f7854d009ec51212840d9629d81df6daf. Branch runtime/slice-7-on-c2.
Only the requested source commits were cherry-picked, individually with -x. No merge or main update.

| Source | Pick |
|---|---|
| a3f28d97 | d2676c26 |
| 257d6f3a | aa5134cf |
| 325aded7 | 1ed07ce5 |
| 23a4f2f0 | 45e6f880 |
| f2de9280 | 6292c5be |
| eccdb2f8 | 08bd0c97 |
| bbc3eb35 | 53d361fe |
| aaf58882 | eeca0e32 |
| 75f116e1 | d7dd19a9 |
| 73b6abff | 2f668a32 |

Integration repairs: 269b94ba (syntax, formatting, cell owner witness), 3e83e014
(graph count columns and independent allocation-tail witness), and this commit
(restored unconditional cycle checks and this report).

## Conflict resolutions

Each conflicted cherry-pick records its resolution in its commit message.

- a3f28d97, docs/runtime-statics.md: retain current audit entries and add Program member classification.
- native.go: retain Release and add ProgramRegion.
- runtime/adamic.h: retain graph bit 29, graph slow counting, current object metadata layout and full allocation sizing; add Program bit 28 and no-count paths.
- graph_regions.h: retain slice 6 graph declarations/prefix and add Program storage seam.
- heap.c: retain graph counting, cell-to-heap owner redirection, map relocation, child destruction and graph storage freeing; add Program adoption storage, no-count paths, region end and member accounting. Program adoption still requires references == 1 and refuses graph/shared/statement allocations.
- heap_parallel.h: keep graph allocator bounds and use the lower Program bound when enabled.
- share.c: retain graph sharing refusal and add Program sharing refusal.
- 257d6f3a, fresh/corpus_units_test.go: retain existing final units.
- ir.go: retain graph metadata and add Program metadata.
- lower/cycles.go: initially preserved default graph classification and Program checks behind programPlan; verification proved this gating wrong. Final code restores unconditional slot/captured-variable checks with Program skips inside them, as directed.
- lower/lower.go: retain c2 logical callable targets, add Program discovery, keep slice 5 async normalization, and refuse async Program functions after normalization.
- emit_arrays.go, emit_expressions.go, emit_locals.go, from.go: preserve graph emission and add Program allocation adoption.
- emit_objects.go: preserve c2 spread/tuple/readiness/representation emission, slice 4 conditional zeroing and graph adoption; add Program adoption. Do not import absent insertion-order APIs from source-branch ancestry.
- region.go, reuse.go: exclude both graph and Program members from statement-region/reuse paths; retain current metadata layout.
- counts.md in 257d6f3a and aaf58882: take current side for final generator; all existing rows remain.
- Core C witness adapts the environment owner to adamic_heap and independently pins the full current allocation tail rather than depending on absent insertion-order accessors.

## Verification

Submodules initialized at their pins through HTTPS; the full-history TypeScript
clone was stopped and replaced by a shallow recursive clone. Locked stage3/api
packages were installed with npm ci --ignore-scripts.

go build ./... passed under timeout 600 before the final unconditional-check
repair. go vet ./... passed under timeout 600 both before and after that repair.
The final build rerun also passed under timeout 600 after the repair. The corrected lower/oracle test binaries compiled.
All three initial test binaries compiled in one background go test -c command
under timeout 900; running jobs were checked at least once a minute.
Repository tracked Go files passed gofmt -l; git diff --check passed.

The requested native/oracle broad patterns hit their 85-second internal timeout
inside a 90-second external kill. The initial lower broad pattern completed with
a missing @types/node host failure; census tests skipped without their external
pristine TypeScript census root. Node dependencies were then installed.
Native explicit units use individual external 90-second kills. SignalAndExit was
split into its exact 30 leaves; native-union.json checks the planned top-level
inventory union. Actual completed and unfinished units are recorded separately;
this is not a claim that every planned unit ran. No completed unit exceeded 60s.

## Required 269b94ba refusal proof

The initial binaries have the exact lower production code from 269b94ba;
3e83e014 changed only a native C witness and the counts test schema.
Run every lower test containing adamic/cycle-capable: NestedCallbackCycleIsRefused,
AsyncGeneratedIdentityCannotBeClaimedBySource, PromisePayloadCannotHideUserCycles,
and NamespaceAmbientHostInitialization. Run oracle suites selected by Refus,
including every review/refused fixture, in ADAMIC_PROGRAM_REGION=0 and =1.
Each invocation has a 90-second external kill. Logs are included for both modes,
before and after restoring unconditional checks.

Before repair: default mode accepts cwd plus eleven review/refused programs;
Program mode accepts additional user-generated class/Promise back-references and
cwd. NestedCallbackCycleIsRefused in Program mode gets the canonical adoption
boundary refusal instead of the required cycle-capable diagnostic.
After repair: all selected lower checks and oracle refusal suites PASS in default
mode. Program mode still FAILS the selected lower checks and TestReviewProgramsRefuse:
Program membership skips accept selected cycles; the nested callback keeps the
canonical boundary diagnostic. The existing nested callback fixture covers a
captured strong back-reference without opt-in, so no new fixture was needed.

Named decision: retaining the requested Program member skips conflicts with
requiring every legacy refusal to remain a refusal with Program enabled. Do not
silently remove those skips or change refusal expectations. Stop with the
unconditional checks restored; further membership/refusal policy needs a decision.

Fixture 14 was attempted independently after installing host types. It fails at
14_getCurrentDirectory.a:16:24 because this c2 base refuses the non-null assertion !.
The optional storage/field-only mutant witness also fails lowering because delete
is unsupported on this base. Neither missing feature was imported from retired
source-branch ancestry. 73b6abff pins a constructor refusal; it contains no
production adopt-before-cache implementation.

## Counts and remaining work

Full TestCountsAreRecorded -update-counts was launched as a background generator
under timeout 900. First attempt failed from missing host types and was stopped.
After installing them and restoring unconditional checks, the generator encounters
cycle refusals for existing fresh_refused and graph-region fixtures; failures are
recorded in counts-failures.log. It was stopped at the named policy boundary.
counts.md is byte-for-byte equal to slice 6: no rows moved, and no Linux row was
removed. Counts were not successfully regenerated. Program rows and any changed
measurements remain uncommitted because there is no valid complete generator result.

K60/K144 counted expectations remain unchanged. Selection uses checker identities
and scalar/container analysis, with no name-based overrides. A full external
census was not run (the required pristine census root was absent).

No complete green native/lower/oracle gate is claimed. Unfinished native units,
corrected broad pattern reruns, successful counts regeneration, the external
census and a passing host 14 witness remain outstanding. Running verification
jobs were stopped before ending the turn. No merge to main occurred.

Native units completed: 35, passed: 35, unfinished: 21.

## Follow-up baseline cycle verification

See [the slice 6 baseline evidence and counted-equality stop](../runtime-slice-7-cycles/report.md). Slice 6 accepts all twelve in default mode. This follow-up changes no production source or fixture classification.
