Continued checked-views integration from d718a9ffa0432245b69a3545c3ca080db865e0bb.
Commits: merge ledger below; only codex/views-integration is eligible for pushes.
Validation: scoped backend tests pass; 152 new original-array probes and measured counts pass in 361.786s; requested checked-view/fixture oracle filter passes.
Mutants: ten array guard mutants and three callable boundary mutants were executed and caught in an isolated checkout, with all source mutations restored.
Limits: latest owner integration is unresolved; additional optional suites and the global count refresh reproduce existing starting-commit failures.

## Merge ledger

| Item | Fetched tip | Conflict decisions | Branch SHA after item | Failures |
| --- | --- | --- | --- | --- |
| Optional boolean | a2eb65ca76816f895f846f78a210bc1717a5f7c4 | Already an ancestor | d718a9ffa0432245b69a3545c3ca080db865e0bb | None introduced |
| Untagged object unions | cd32db4234fb0a65a6ad199474b802c9e7491054 | Already an ancestor | d718a9ffa0432245b69a3545c3ca080db865e0bb | None introduced |
| Lazy admission owner | d5b3c4a9bde6ee28fe38568ef4b720aa9c2b68e3 | Aborted before resolving 18 paths; Map adapter restoration must preserve boxed dispatch and ownership | d718a9ffa0432245b69a3545c3ca080db865e0bb | Not runtime-tested; unresolved merge boundary |
| Arrays and parser | a3b0e3570fdf83fa3071b7f34ff9780d5334d5f7 | Plan: retain existing duplicate sections, whitespace and both final reports; counts: retain every distinct row and reject differing values for one fixture | e636841dd3d2996e471d97dd23712e7dbaa6aaeb | Required filter passes; inherited failures below |
| Mixed unions, first tip | d78c3f5cf19959be2bdb2ba4b72951750303cc4a | One plan hunk: append both independent reports; update the primitive helper to preserve its wrong-boolean runtime refusal | This merge | Stale compile-only helper expectation fixed; inherited failures unchanged |
| Mixed unions, requested newer tip | 37763565c317a80c3b362e475edab8b61ab0eafd | Pending | Pending | Pending |
| Object and primitive unions | 8abb52a1518b9d8c3f0dbde993551d4f01ba10e2 | Pending | Pending | Pending |
| Intersections | 4c3c3009c1ab74e4da902ffa9357b81ec0cf7d95 | Pending | Pending | Pending |
| Callables | baa72933a1504a6d186e5f307a0a4c9684c95615 | Pending | Pending | Pending |
| Tuples | 21fd3587d6a32fef3f195a31be84605c02f58953 | Pending | Pending | Pending |
| Dictionaries | deec3c933543c7b5e7ba0d9db0964265d03c494c | Already an ancestor | d718a9ffa0432245b69a3545c3ca080db865e0bb | None introduced |

## Owner boundary

The earlier bd05075f rollback removed Map storage adapters and supporting entry
descriptors/tests. The latest owner tip adds nominal Map behavior on that base.
Its incoming native Map forEach calls closure code directly; current integration
uses viewCallableBoxedInvokeTypes. Converted key/value ownership also differs.
The merge was aborted rather than replace the current checked callback dispatch.
Restoring the general Union calling-convention refusal is not a resolution:
entry-live-mutation.a then refuses lowering and supported boxed callbacks regress.
The current unknown-producer refusals and original never-member helper refusal
remain. See docs/checked-views-blockers.md for the retained boundary.

## Commands and observations

All test output was redirected to files. Compressed exact logs are in
logs/continued, including mutant failure output and the restored never control.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Initial timings: Node 0.058s, Go 0.073s,
clang 0.492s, markdown ready 0.886s, submodules 194.800s. The initial build
encountered temporary owner-merge conflict markers. After aborting that merge,
setup succeeded: build 29.274s, cache ready 29.407s, done 29.438s; nproc=5,
cgroup cpu.max=400000 100000. Pinned stage3/api dependencies were installed with
npm ci --prefix stage3/api after the first backend run found missing Node types.

Pristine microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 was
checked out in /tmp/checked-views-upstream. The lane2 original15 through original18
prepare.cjs adapters emitted complete declarations into /tmp/checked-views-arrayN-decls.
Shared lane4b declarations were prepared before the lane4 and lane7 manifests.
No source was copied from cohere. The isolated starting checkout references the
existing cohere submodule and Node dependencies through symlinks.

```sh
go test ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestLazyView|TestPrepareViewCallableRead|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload' -count=1 -timeout 30m
go test ./internal/oracle -run 'Test.*View|Test.*Union|Test.*Tuple|Test.*Dictionary|Test.*Callable|Test.*Optional|Test.*Array|TestNativeAgreesWithNode/.*(view|union|tuple|dictionary|callable|optional|array)' -count=1 -timeout 30m
go test ./internal/oracle -run 'Test.*View|TestNativeAgreesWithNode/.*(view|union|tuple|dictionary|callable|optional|array)' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCheckedViewRanked(15|16|17|18)(OriginalArrays|ArrayCounts)$' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

The backend run passed: lower 7.601s, native 7.669s, JavaScript 0.845s.
Original-array testing supplied ADAMIC_ARRAY15_ORIGINAL_DECLS through
ADAMIC_ARRAY18_ORIGINAL_DECLS and passed all 152 probes and count rows in 361.786s.
Successful probes include leak checks; all probes compare Node and sanitized/
release native plus JavaScript output, with exact refusal diagnostics.

The broader oracle run failed in 834.349s on fourteen optional-suite subtests.
Twelve expect older scalar/literal diagnostic wording, while the checked union
read emits matches-no-member diagnostics. The nominal-slot and
nominal-subclass-slot fixtures expect behavior outside the current nominal
read/source-certificate frontier. The exact starting SHA reproduces the same
fourteen subtests in 18.675s. No refusal or test expectation was weakened.
The separate requested checked-view/fixture filter has its own exit result.

Global counts failed in 66.224s on 41 fixtures and wrote no counts table.
An isolated checkout at the exact starting SHA failed on the identical 41
fixtures in 83.139s. The full fixture list is logs/continued/count-failures.txt;
exact compiler, clang, runtime and readiness failures are in both compressed logs.
The new 152 array rows were retained, then independently remeasured successfully.

## Executed mutants

In the isolated checkout, each original15 through original17 runner removed the
native and JavaScript numeric field checks. Witnesses were function-wrong-pos,
function-parameters-wrong-pos and signature-type-parameters-wrong-pos respectively.
All six failed their oracle on execution. Original18 additionally removed native
and JavaScript numeric checks (json-diagnostics-wrong-element) and join element
checks (config-files-join-bad); all four failed on execution.

run-callable-boundary-mutants.py replaced native and JavaScript unknown-producer
refusals with ordinary invocation; TestViewCallableBoxingUnknownProducer caught
both executing mutants. Broadening the scalar exemption was caught by
TestCheckedViewCallableRetainsNeverWiderHelper, and the mutant execution control
also passed with ADAMIC_CALLABLE_NEVER_MUTANT=1. All thirteen source mutations
were restored before any integration commit.

Array filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	366.910s

## Mixed union first tip

The only merge conflict is one plan hunk: retain the array certification report
and append the independent lane4 reconciliation report. All compiler hunks
merged automatically. Unknown/untracked receiver fallback and the existing
callable-only concrete-read exemption remain intact.

The new complete-declaration conversion test passes in 0.009s. Scoped backend
tests pass: lower 5.681s, native 5.244s, JavaScript 1.430s. Complete original
inputs activate brand, primitive, array and intersection oracles. That run
finished in 1457.212s with exactly one obsolete compile-only helper expectation.
Its other cases pass, including 66 new primitive first-member/boolean-first and
member/literal-bypass mutant executions on native and JavaScript.

The original helper-unsupported.a boolean is now rejected at its helper field
read with exit 70 and the exact string | number declaration. The corrected test
also passes ordinary and viewed valid string/number objects through the same
helper, compares source Node, release/sanitized native and JavaScript, and checks
successful-run leaks. Two member-check mutants execute the original wrong value
and print ordinary/true at exit 0, proving the retained runtime refusal.
The corrected helper test passes in 1.753s. No compiler guard was changed by this
expectation update. The fresh filtered oracle includes TestPrimitiveOrdinaryProperty.

The broader lower refusal corpus additionally has two stale known-Union-callback
expectations (a union a function value takes; Array.from's undefined as a union).
The exact starting checkout reproduces both. No callback guard or expectation
was changed for these failures. Counts refresh still fails on the exact same 41
fixtures as the starting commit, in 50.651s. Callable unknown-producer and scalar
never-exemption executable mutants were rerun against the mixed-union source;
all three were caught.

User steering during validation supplied newer lane4 tip
37763565c317a80c3b362e475edab8b61ab0eafd. It was fetched and will be integrated
before moving to lane4b.

Fresh mixed filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	637.595s
