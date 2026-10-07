Built: preserve structural accessor targets when class dispatch is finalized again; no Error representation change.
Commits: merges origin/main c7991b900362796aefd111474e65eb5398e91953 into codex/error-classes-counts; no main or area branch writes.
Validation: uncached internal/lower, internal/ir, whole internal/oracle, and regenerated counts; all passed; results below.
Mutant: restore the old destructive finishClassCalls behavior at the existing pass order; class_features_accessors.a panics in CallTargets.
Not covered: no full repository gate beyond the three requested packages; no library-error-value implementation changes.

The dependency merge reproduced `panic: ir: virtual call has no target set` on
`class_features_accessors.a`. `finishAccessors` registered structural accessor targets,
then `guardRuntimeRanges` called `finishClassCalls`, which erased that registration
before calling `CallTargets`. A second call from exceptions would erase it again.

`finishClassCalls` now initializes the target map only when absent, then adds each
newly discovered class implementation with its existing deduplication. Class discovery
is additive, and the pass preserves accessor registrations for all subsequent readers.
This fixes the target lifecycle at its owner, without changing pass order or introducing
an error-class special case. `TestFinishClassCallsPreservesAccessorTargets` exercises two
finalizations and checks both structural accessor targets and deduplicated class targets.

The merge conflict in narrowed.go keeps main's element-access diagnostic and the
branch's existing conversion of property/element failures to catchable nominal TypeError.
The counts conflict was resolved by regeneration. No edits to error_classes.go or its
name/message/cause prefix were made.

Main's new IR reader audit initially rejected twelve groups of existing branch readers.
The audit now records narrow file/function/field entries with reasons: direct generated
guard body inspections, direct ReadyErrors constructors, named sort comparators under
Callback nil, and assertions inspecting guard identity or receiver binding. Virtual call
analysis continues to use CallTargets. The audit still rejects unlisted readers and stale
entries; its parser and enforcement logic are unchanged.

Counts regenerated successfully: 369 rows, 27 new main fixtures. Exactly four existing
feature-branch rows changed, all due to incoming main changes. Values below are
allocations, frees, retains, releases, peak live, and in regions.

| Fixture | Before 66766ab | After | Cause |
|---|---|---|---|
| string_index.a | 89,89,29,126,7,0 | 89,89,45,142,7,0 | f65fc05 adds narrowed element-read checks and ownership holds; equals current main. |
| undefined_keys.a | 122,122,173,238,24,0 | 122,122,196,261,24,0 | f65fc05 preserves undefined in accepting contexts and checks narrowed element reads; equals current main. |
| set_undefined.a | 26,26,29,46,7,0 | 26,26,31,48,7,0 | f65fc05 element/narrowing correctness; equals current main. |
| class_as_interface.a | 372,372,362,526,60,0 | 372,372,351,515,60,0 | 6965ac2 calls proven singleton methods directly, removing eleven ownership holds; the branch retains its existing nominal exception overhead. |

`class_features_accessors.a` remains 67,67,57,112,21,0. The target-preservation fix
adds no count increases relative to the prior feature branch.

Commands and logs:

- Initial setup: `/tmp/adamic-accessor-setup.log`; warm build failed because narrowed.go still contained merge markers. Conflict resolved before testing.
- Repeated `bash cloud/setup.sh`: `/tmp/adamic-accessor-setup-fixed.log`; Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 180s, total 180s. `nproc`: 5; CPU quota 4; memory 17.6 GB.
- Reproduction: `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_features_accessors.a'`; exit 1, intended panic. `/tmp/adamic-accessor-before.log`.
- Mutant: the same fixture command with `-overlay=/tmp/adamic-accessor-overlay.json`, restoring unconditional target-map clearing; exit 1, the same intended CallTargets panic, without compilation failure. `/tmp/adamic-accessor-mutant.log`.
- New assertion: `go test -count=1 ./internal/lower -run '^TestFinishClassCallsPreservesAccessorTargets$'`; exit 0. `/tmp/adamic-accessor-targets.log`.
- Counts: `go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts`; exit 0. `/tmp/adamic-accessor-counts.log`.
- First requested gate: lower passed 33.126s, the whole uncached oracle passed 102.836s, IR failed only the unlisted reader audit. `/tmp/adamic-accessor-gate.log`.
- Final requested gate: `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/ir ./internal/oracle`; `/tmp/adamic-accessor-final-gate.log`; exit 0: lower 50.293s, IR 2.990s, oracle 199.155s. The oracle compares source Node, native ASan/UBSan and release output, and generated JavaScript with Node; successful native programs must leak nothing.
- `go vet ./internal/lower ./internal/ir`: exit 0; `/tmp/adamic-accessor-vet.log`. `git diff --check`: clean.
