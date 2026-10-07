Integrated lazy admission first from the lane-1 base ab4d6f902.
Owner tip e6aec805 includes lane-2 merge 592f1f71 and shared-flow merge c01ae313.
Focused package tests and vet pass; filtered Node comparison excludes one reproduced baseline failure.
Nested-loop rollback mutant is caught by string_views_policy.a: Node exits 0, mutant native exits 70.
Full repository gate and production tsc census were not rerun; baseline failures remain listed below.

Merge order follows the user's owner-first ruling. This first merge is clean,
with no conflicting hunks and no bulk or whole-file resolutions.

The integration repair in internal/lower/readiness.go applies the existing
markProgramViewArrayUse policy inside recursive struct traversal, so nested
for-of statements receive the same policy as top-level statements. Ordinary
nested loops previously retained preliminary view metadata despite no admitted
array view. The filtered Node fixtures library_array_copy_within.a,
string_views_methods.a and string_views_policy.a prove the repair. The rollback
Go overlay removes exactly that traversal repair; string_views_policy.a then
executes valid C and exits 70 with uncertified storage, against Node exit 0.
The owner's demanded callable-read refusal and tests are retained unchanged.

Validation commands (each writes full output to its named log):

```
go test ./internal/lower ./internal/ir ./internal/native -run 'TestView|TestLazyView|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View|Test.*Phantom|TestNativeAgreesWithNode/internal/oracle/testdata/.*(view|union|brand|optional|array)' -skip 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_holes_callbacks.a' -count=1 -timeout 30m
go vet ./...
```

The unskipped filtered oracle was run first: library_array_holes_callbacks.a
refuses at line 12:37 with adamic/no-type-predicate for the inline predicate
(value) => value !== undefined. The same refusal is reproduced in detached
c01ae313. This baseline fixture is explicitly excluded from the green rerun.
A broader package filter also reproduces TestPhantomArrayCastsAreErased,
TestPhantomArrayRequiredCastsAreErased and TestPhantomArrayProofs/cycle failures
on c01ae313. The parent also fails the tuple_representation refusal wording pin.
These failures were not suppressed by compiler or test changes.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Node ready 0.072s, Go 0.098s,
clang 0.504s, Markdown 0.948s, submodules 662.275s, Go build 1177.123s,
setup done 1177.419s; nproc=5 (CPU quota 4). Stage3 API pinned Node declarations
were installed with npm ci --prefix stage3/api. Early gate attempts before
submodule readiness failed on missing cohere/TypeScript/tsc/go.mod; subsequent
concurrent cold gate attempts were interrupted and not counted as evidence.

Existing carried evidence has whitespace errors; historical logs and raw
fixture evidence are preserved. No cohere code was copied.
