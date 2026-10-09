# Namespace test defense

Base: e77a4ae41f473c149aee910c51b73637686a804a. Both requested names exist on this commit.

## Code under test and oracle

Code under test: Adamic lower.Lower, module namespace classification, module-read proofs, namespace readiness and binding access, and emitted native/JavaScript programs. Oracle: unchanged source execution on Node. Comparisons require matching exit status, stdout and stderr; successful runs also check leaks. The boundary row additionally pins Node stdout to false:false, true:true, false:false. This is a self-written source observation pin, not a production expectation that a compiler mutant can invalidate.

## Coverage and execution difference

Per-row coverage used -coverpkg=./internal/lower,./internal/native,./internal/javascript. Each requested row was compared against the generic family's identical registered fixture(s). coverage-differences.json records zero exclusive covered production blocks for both rows. Four compressed profiles and their command output are retained. reached-functions.txt lists 426 measured reached production functions.

The module row runs the six module_namespace_reads fixtures; every fixture is also registered in the generic family, with checked=false. The boundary fixture stage3/namespace-live-export/live.a is likewise registered. The generic family compares the same Node observation to both emitted backends, checks leaks, and additionally compares release versus sanitized execution. The boundary calls nativelyUncached directly; the generic family calls natively, whose uncached path calls exactly that same helper with the same sanitize option. Mutant runs force ADAMIC_GATE_UNCACHED=1. These rows are not separate Node/native executor twins: each already executes both backends.

## Attempts and results

D01 flips module provider ordering; D02 flips library exclusion in namespace classification; D03 flips the proven-read condition. Each makes TestModuleNamespaceReadsMatchNode fail and also fails the generic family on the same fixture. D04 changes namespace completion to false; D05 selects the preceding readiness parameter instead of the qualified read's result; D06 flips namespace member scope classification. Each makes TestNamespaceLiveExportBoundary fail and also fails the generic family's live.a comparison. All six diffs passed go vet ./internal/lower/. All six passed git apply --check against the clean base before supplementary replay. Production files and all tests are restored unchanged.

rows.json groups failures by top-level row. results.json retains exact failed, passed and skipped subtests plus commands. Some witness rows fail because production lowering preconditions broke; those failures are not evidence that a witness's guarded comparator works. The ordinary generic agreement family's actual output comparisons alone suffice to show each attempted defense has a co-catcher. No mutant survived the bounded matrix. Neither row is defended after its three attempts. This is evidence about six faults, not permission to delete tests.

## Scope, baseline and limitations

npm ci ran in stage3/api. Warm tools were used; setup skipped; nproc=5. Whole-package baseline timed out at 90.119 seconds, with no assertion-failure events before timeout. It was narrowed to current namespace/import-cycle top-level tests and registered module/namespace generic fixtures, plus current stage3 namespace shape fixtures. Three expanded uncached clean runs passed. matrix-selectors.json and matrix-observed-tests.json specify exact selection and observed names; results.json records every pass and failure. Unselected package rows and repo-wide uniqueness are unknown. No bounded mutant run timed out or aborted on a Go panic.

## Brief costs and ambiguities

/tmp totals only 8.8 GB, making the requested 15 GB free threshold impossible. Its initial free space was 2.6 GB; /workspace had 14 GB. Only the identified previous unit's /tmp/unicode-defend scratch was removed (4 KB); unknown shared caches and all repository/tools were preserved. No full-disk failure was observed.

The latest origin/main differs from the audit base; this defense uses current names and current namespace registrations. Audit report artifacts are copied here for traceability. The full-package 90 second timeout forces a bounded matrix; it cannot establish package-wide uniqueness, but an observed generic co-catcher disproves uniqueness for each attempted fault without needing unseen rows.

A subtest path contains slashes. A parent-only generic regex would accidentally include unrelated fixtures or exclude all selected children. Family patterns explicitly match path components; exact observed children are saved.

The boundary's extra pin concerns source Node behavior. Mutating Node or the pin would violate the rule against oracle/harness changes, so its additional assertion cannot be defended by a production mutant here. Both row names accurately describe their output assertions; neither claims speed or a threshold it does not assert.

Initial six mutant cycles, including vet and two bounded test commands per mutant, took the aggregate seconds in base.json. Additional clean and shape replay logs retain the test binary's elapsed times. Coverage overhead and compilation time are included in process duration, not equated with test binary time. Baseline-wide profiling, other packages, arbitrary compiler faults and future changes were not covered.
