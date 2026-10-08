# NonNullable<T> contribution clock

Status: cancelled as a census echo for the reserved kind `a value of type NonNullable<T>` (7 listed roots). No claim that these seven roots now lower, and no ABI was invented for an unsubstituted T.

Assumption: the brief permits cancellation when only the isolated generic context remains. Start from current origin/area/compiler b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861. The direct checked-type probe tests the actual checker intersection and substitutes string and object types through the existing checker mapper. representation already calls concrete first, so the checker removes the empty non-nullish view and storage is String or Object respectively. Without substitution the same checked type remains unrepresented. No production helper, admission guard or separate runtime helper was added.

The exact requested minimal fixture is registered as stopped, retaining its independent Node output text/true. Its next production stop is `a function returning NonNullable<T>`. The existing inferTypes helper cannot read back T from the intersection, even with explicit type arguments. That is generic inference/return work, outside this contribution's ownership. functions.go and generic.go were not edited. Ownership history was inspected with git log --remotes=origin/codex/notyet-* before any lowering work.

The separate concrete fixture adds a witness parameter of type T and supplies the same value twice. This gives existing inference concrete evidence without altering the NonNullable<T> parameter or result. It is a supported neighbor, not a claim that the exact requested minimal fixture passes native compilation. Both string and object specializations retain their values. The named mutant clock-nonnullable-t-drop-object-return replaces exactly one object specialization return with ir.Undefined{Of: ir.Object}; the string specialization is unchanged.

Reconstruction uses git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3 and that archive's apply.sh. All 81 byte lengths and SHA-256 hashes match the original tsc source-manifest.json from 3cead3fd. The guarded replay harness was transiently restored from 3cead3fd because the current compiler baseline lacks its entry point; no harness source is carried in this commit. It ran against the current compiler baseline and exact adapted source bytes, with the setup environment sourced.

The census harness attempts generic declarations as written without invented instantiations (lower.go.txt). At core.ts:246:19, countWhere<T> has no concrete caller substitution; its checked non-null array element is NonNullable<T>. The empty structural member {} excludes nullish values but can contain primitive values and cannot decide storage for T. This representation stop must remain in the isolated census context.

Validation (all normal checks exit 0):

- `go test ./internal/lower -run TestClockNonNullable -v`: direct checked-type string and object representation probes pass, and the unmapped intersection remains unknown.
- `go test ./internal/oracle -run 'TestClockNonNullable|TestNativeAgreesWithNode/internal/oracle/testdata/representation_clock_nonnullable' -v`: original Node text/true and its named production stop; concrete fixture Node/JavaScript/release/sanitized native agreement and leak checks; named sanitized object-return mutant prints text/false with empty stderr.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: passes; concrete fixture allocations/frees 2/2, retains/releases 4/6, peak 2, regions 0. Exact minimal fixture is stopped and thus not counted.
- Guarded exact-site replay exits 0, reproducing NotYet `a value of type NonNullable<T>` at core.ts:246:19. This is retained evidence, not a claim that the representation stop is gone. Full stdout is preserved compressed.
- Two additional temporary compiler-overlay mutants each exit 1 at the direct checked-type test: invent-object-abi is caught because isolated T & {} wrongly acquires Object storage; drop-concrete-substitution is caught because concrete string and object fail representation. Neither modifies production source or adds a new admission guard.

Raw successful checks and deliberately failing mutation checks are in evidence. The replay uses the brief's exact command and /tmp/notyet-representations-adapted input. No IR or backend changes and no separate runtime helper. No generic-return or other clock kind was implemented. The cancellation concerns the isolated census representation kind only; support for the exact minimal generic function remains for the generic inference/return owner.

## Fast-gate follow-up

The fast gate on 35fad67b failed in flow's TestEveryFunctionIsInSingleAssignment because flow independently globs top-level oracle fixtures and assumed they all lower. The exact cancellation fixture is deliberately registered as stopped and has no IR. Exclude only representation_clock_nonnullable_t.a from flow's shared programs list, with a comment explaining why. The concrete fixture remains included, and the oracle still independently verifies the exact source's Node output and expected generic-return stop. No production lowering, generic-return, backend, registry or counts changes are part of this follow-up.

The reported SSA test passes: 4,206 functions, 648 phis, 4,517 values and 12,659 uses checked. The oracle fixture, backend/leak checks and named object-return mutant also pass. A narrow trace-only selection verifies the concrete fixture; selecting only this immutable witness for mutation and liveness checks triggers those suites' non-vacuity assertions, so the complete flow package is run for final validation. Follow-up output is retained separately from the original contribution's checks.

Final follow-up validation: `go test ./internal/flow -count=1` passes in 100.123s, covering all four consumers of the shared fixture list. Focused oracle checks pass in 0.530s. No counts refresh is needed because fixture registration and runtime behavior are unchanged.
