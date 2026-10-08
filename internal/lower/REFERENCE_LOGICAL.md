Built: reference && and || select lazy branches from pointer presence and evaluate the left once; mutable views remain checked.
Commits: b1ba9a7b implementation, 60e8e422 structural truthiness guard, 069dbff7 latest main merge, 187f9e43 flow fixture repair; October 8, 2026.
Checks: source Node, both backends, touched packages and analyses, uncached oracle, counts, vet and formatting; results below.
Mutants: ten caught, including duplicate left evaluation and empty-array falsiness; details below.
Limits: fixture 25 clears 1150:32 and next refuses at 1156:33; untagged unions holding both null and undefined remain NotYet when either absence can escape.

# Reference logical operands

The checker types the whole logical expression. The left of a reference && is
only a presence test: its elements do not become elements of the result. The
right operand is still checked against the receiving type, including writable
array invariance. The left of || can become the result and retains its view
obligations. Typed slot and conditional checks are preserved.

Arrays, objects that exclude primitives, Maps/Sets and closures are truthy
whenever their nullable pointer is present. Strings, numbers and booleans keep
their existing paths. Structural types such as {} also admit falsy primitives;
checker assignability excludes these from the reference truthiness proof.

The IR's lazy Coalesce has a ReferenceAnd mode. Native evaluates the left once,
then evaluates the right in an aside only if that pointer is present. An absent
left produces an absent result in the whole expression's representation. The
JavaScript backend emits && for that mode; reference || uses the existing
nullish branch, which has the same behavior for the proven left types.
The ordinary coalesce borrowing shortcut is excluded for ReferenceAnd because
its branch direction differs. Fresh results retain their ownership, and the
sanitizer and leak oracle hold runtime-built string elements.

A nullable reference type may hold null instead of undefined using NULL and
checker metadata. A type that can hold both requires a distinct absence tag
before either can escape. That representation was not added here. A slot read
holding both can still be consumed by || or ??, or inspected by typeof, whose
existing backend retains slot presence. Direct comparisons that would collapse
null and undefined receive NotYet. This boundary is held by a mutant. Stale
narrowing checks stay active when a scalar result cannot represent absence.

## Fixtures and observations

internal/oracle/testdata/reference_logical.a covers:

- arr && arr.map with present, absent and empty arrays, changing the element type.
- Object || fallback with both branches; nullable arrays, objects and closures.
- Map fallback, a nested &&/|| chain, counters for one left evaluation and a lazy right.
- Whole result representations number | undefined, string | undefined and array | string.
- Generic T/U mapping with number and object results; indexed array references.
- Owned mapped strings and nullable slots observed through typeof and consumed by ||.

Source Node exits 0. Key observations are present: for the empty map result,
4,2 for left/right counters, undefined,0,2 for the scalar result, and
object,undefined for a present null slot followed by a missing slot. Native
release, ASan/UBSan malloc/slab builds, leak checks and generated JavaScript
agree with source Node.

reference_logical_narrowed.a deliberately restores undefined after narrowing.
Source Node exits 0 and prints undefined. Both compiled backends fire the existing
inserted check, exit 70, and agree with each other. It is registered as checked,
so this difference is explicit rather than claimed as Node equality.

Lowerer tests retain mutable Dog[] to Animal[] refusals through both || and &&,
exclude primitive and structural primitive truthiness, and refuse an untagged
nullable lookup or parameter that could expose both absences.

## Mutants

Run internal/lower/testdata/run-reference-logical-mutants.py. An optional list
of mutant names selects a subset. Each source is restored in a finally block;
every run writes /tmp/reference-logical-mutant-<name>.log.

| Mutant | Observation that caught it |
| --- | --- |
| trace-refused-fixtures | Restored main's eight intentionally refused fixtures to the SSA glob; Lower fails before a graph can exist |
| double-left | Native counter 6,2 instead of Node's 4,2; stdout differs |
| empty-array-falsy | Empty array produces absent instead of present; stdout differs |
| lose-and-view | Valid readonly string[] left refused as the number[] result view |
| accept-writable-result | Mutable view test receives nil instead of Refused |
| borrow-and-as-coalesce | Indexed && chooses the coalesce branch; stdout differs |
| lose-narrowing-check | The deliberately checked fixture fails to exit 70 |
| collapse-lookup-absences | Nullable lookup incorrectly lowers; pinned NotYet test receives nil |
| use-string-presence | Primitive logical expression incorrectly lowers; NotYet test receives nil |
| treat-structural-primitives-as-references | {} admitting 0 incorrectly qualifies; NotYet test receives nil |

The first seven ran together. The borrowing, structural and refused-fixture
selection mutants ran after their dedicated witnesses were added. Every mutant was caught; no clang warning
or sanitizer fault was used as a substitute for the stated catcher.

## Commands and setup

All test output goes to files, and every toolchain command sources
/workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/reference-logical-setup.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/reference_logical.a > /tmp/reference-logical-node.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/reference_logical_narrowed.a > /tmp/reference-logical-narrowed-node.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical' -count=1 -timeout 30m > /tmp/reference-logical-focused.log 2>&1
python3 internal/lower/testdata/run-reference-logical-mutants.py > /tmp/reference-logical-mutants.log 2>&1
python3 internal/lower/testdata/run-reference-logical-mutants.py borrow-and-as-coalesce > /tmp/reference-logical-borrow-mutant.log 2>&1
python3 internal/lower/testdata/run-reference-logical-mutants.py treat-structural-primitives-as-references > /tmp/reference-logical-structural-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/reference-logical-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/reference-logical-landing-gate.log 2>&1
python3 internal/lower/testdata/run-reference-logical-mutants.py trace-refused-fixtures > /tmp/reference-logical-flow-mutant.log 2>&1
go test ./internal/flow -count=1 -timeout 30m > /tmp/reference-logical-flow-final.log 2>&1
go test ./internal/oracle -run '^TestClassWrongOutput' -count=1 -timeout 30m > /tmp/reference-logical-main-refusals.log 2>&1
go vet ./... > /tmp/reference-logical-vet.log 2>&1
gofmt -l cmd internal > /tmp/reference-logical-fmt.log
git diff --check
```

Setup succeeded: markdown dependencies skipped 0.008s and ready 0.097s,
submodules 0.116s, clang 0.245s, Go build 32.547s, cache warm 32.778s,
total 32.803s. nproc=5, cgroup CPU quota 4; Go 1.27.1, Node 24.19.0,
clang 20.1.8. Setup did not warm test binaries.

Counts passed in 17.943s. Only two new rows were added: ordinary fixture
92 allocations / 92 frees / 91 retains / 190 releases / peak 13 / regions 0;
checked fixture 1 / 1 / 1 / 2 / peak 1 / regions 0.

The gate after the first main merge passed: lower 43.524s, ir 7.789s,
flow 158.286s, fresh 78.553s, native 253.503s, uncached oracle 254.831s;
JavaScript has no separate package tests. The final structural guard was
mutant-tested and main subsequently advanced to bfa0bfec, merged in 069dbff7.
The next landing gate on latest main passed lower 54.021s, ir 1.835s,
fresh 80.342s, native 251.477s and uncached oracle 240.260s, but flow failed in
164.975s. The eight fixtures newly refused by main remained in flow's broad
oracle fixture glob, so trace, range, liveness and SSA tests tried to construct
graphs for intentionally refused programs. This is an observed main integration
failure, not a logical-operand runtime discrepancy.

The flow fixture chooser now excludes exactly those eight named fixtures;
unexpected lowering errors remain failures. TestClassWrongOutput103/106/107/108
continue to pin each refusal and its source Node observation. No compiler rule
was weakened. Restoring the old fixture chooser is caught by the SSA test at
Lower. The complete flow package is rerun in /tmp/reference-logical-flow-final.log,
and the pinned main refusals in /tmp/reference-logical-main-refusals.log.
The complete flow rerun passes in 56.863s. The pinned main refusal oracle tests
pass in 1.087s. Vet and formatting produce no output; unit diff checks pass.
The final branch includes origin/main bfa0bfec, verified again by fetch and
merge-base before pushing. All changed packages have passed on this compiler
implementation. No runtime source changed after that landing gate; the final
repair changes only the test fixture chooser.

## Pinned host scratch

The local scratch branch codex/generic-function-value-host-scratch retains
c97402ba as an ancestor. Compiler conflicts preserve the host's dynamic type
handling, Node filesystem argument exemption and callable/array contexts, plus
the new logical lowering. Host fixture conflicts preserve pinned source files.
The actual fixture 25 blob remains bd24f7bd2ed712f6d07455130ab8bbe76c82b743,
identical to c97402ba, including its line numbers. No library branch was pushed.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/host/25_readDirectory.a$' -count=1 -timeout 30m > /tmp/reference-logical-final-host25.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs stage3/fixtures/host/25_readDirectory.a > /tmp/reference-logical-host25-node.log 2>&1
```

The 1150:32 refusal is cleared. The actual fixture still stops before complete
lowering. The next stop is:

```typescript
const results: string[][] = includeFileRegexes ? includeFileRegexes.map(() => []) : [[]];
```

At 1156:33, the diagnostic is:

```text
Adamic 0.1 refuses a never[] seen as writable string[][], which can write string into a shared never[]; declare the result readonly T[], or return a fresh [] (adamic/invariant-mutable)
```

Observation: this map callback returns fresh empty inner arrays. The current
refusal and its shared-array wording do not prove that the arrays are shared.
This unit records the next contextual nested-empty-array stop without bypassing
invariance. Fixture 25 still cannot reach either backend, so its native or
JavaScript output is not certified, and no 25/25 host claim is made.

Source Node exits 0 and prints:

```text
a.ts,z.ts,alias/b.ts,alias/deep/c.ts
a.ts,z.ts
src/b.ts,src/deep/c.ts,a.ts
plain.js
```

The full repository gate, a-check and all 25 host fixtures were not rerun.
Protected emitter/orchestrator/oracle files were not edited by this unit, and no
cohere code was copied. The main merge imports its own recorded source headers
and corpus filenames. The final scratch merge is 961afb64, containing the final compiler and flow
fixture repair. The filtered oracle again records 1156:33.
Scratch merge-wide whitespace warnings come from existing
main evidence files; the unit's source diff passes whitespace checks.
