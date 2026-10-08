# Nullable generic array views

October 8, 2026.

Built: NonNullable<T>[] viewed as writable T[] is checked after substitution.
Commits: ab33596b implementation; 0e8a6b52 closure/null boundaries and counts.
Commands: setup, full lowerer and uncached oracle packages, counts, Node, vet,
formatting, bypass mutant, and actual fixture 25 on pinned host proof.
Mutant: bypass the instantiated array comparison; unsafe fixture lowered and
its pinned refusal assertion failed with got <nil>.
Limits: fixture 25 moves to the next refusal, 1108:34, never viewed as T;
the full repository test gate and all 25 host fixtures were not rerun.

## Change

Declaration-time array invariance defers only the exact NonNullable<T>[] -> T[]
relationship for a function type parameter. It does not change generic-class,
field, map, or unrelated array variance rules. A deferred declaration is not
permission to emit code: each specialized function is rechecked with its concrete
checker mapper before lowering its body, including obligations inside closures.
The mutable element comparison retains null and undefined.

At T = string both arrays hold string, so a writer can push through the T[] view.
At T = string | undefined the wider view could put undefined into a string[];
Adamic refuses the instantiating call, identifies the argument and gives the fix.
Explicit concrete type arguments are also checked during the refusal pass before
an unsupported result representation can hide a provable refusal. Explicit
arguments still containing an unbound parameter wait for their outer instance.
No checker code was copied from cohere.

The positive fixture is internal/oracle/testdata/generic_nullable_array_view.a:

```typescript
function write<T>(values: T[], value: T): void { values.push(value); }
function collect<T>(result: NonNullable<T>[], next: T): T[] {
    write<T>(result, next);
    return result;
}
console.log(collect<string>(["first"], "second").join(","));
```

Source Node prints first,second. Native release, sanitized malloc/slab builds,
leak checking and the JavaScript backend agree. The refused fixture calls the
same collect at string | undefined and supplies undefined; source Node itself
prints first, and exits 0. Adamic instead gives the pinned call-site diagnostic:

```text
main.a:6:13: Adamic 0.1 refuses type argument T = string | undefined breaks invariance: a value of type NonNullable<T>[] seen as T[], which can write string | undefined where string is read; use a type argument that includes neither null nor undefined, or keep the array view readonly (adamic/invariant-mutable)
```

The lowerer tests pin Refused, its full What and Fix and the instantiating use's
line/column. Additional tests require string | null and a captured closure's
nullable view to be refused. Negative fixtures intentionally have no backend
execution: the required result is a compile-time refusal.

## Mutation proof

run-generic-array-view-mutant.py returns from widenedArguments for the raw
NonNullable<T>[] -> T[] pair without comparing the instantiated types. It runs
TestGenericNullableArrayViewRefused, which fails solely because Lower incorrectly
succeeds: want pinned call-site nullable type argument refusal, got <nil>.
There is no C compilation, warning or sanitizer kill in this mutant. The runner
restores the source in a finally block. It was caught before and after the
null/closure boundaries were added.

## Commands

With /workspace/adamic-tools/env.sh sourced, all test output goes to logs:

```sh
bash cloud/setup.sh > /tmp/generic-array-setup.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/generic_nullable_array_view.a > /tmp/generic-array-node-string.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/lower/testdata/generic_nullable_array_view_refused.a > /tmp/generic-array-node-undefined.log 2>&1
go test ./internal/lower -count=1 -timeout 30m > /tmp/generic-array-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/generic-array-final-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/generic-array-counts.log 2>&1
python3 internal/lower/testdata/run-generic-array-view-mutant.py > /tmp/generic-array-final-mutants.log 2>&1
go vet ./... > /tmp/generic-array-final-vet.log 2>&1
gofmt -l cmd internal > /tmp/generic-array-final-fmt.log
git diff --check
```

Setup succeeded: Node 0.018s, Go 0.022s, submodules 0.061s, markdown 0.063s,
clang 0.142s, Go build 32.137s, cache warm 32.328s, total 32.356s.
nproc=5, cgroup quota 4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8.
Current origin/main remains 74fb6490, already merged into this branch.
Counts pass in 30.386s and change only the new row: 2 allocations, 2 frees,
3 retains, 3 releases, peak 2, regions 0. Final full lowerer package passes in
12.707s; vet and formatting exit 0 with no output.

## Pinned host integration

Local scratch branch codex/generic-function-value-host-scratch contains pinned
host proof c97402ba and the feature implementation. Merge 3b77114c checks the
initial implementation; merge 0ed5fe48 checks the final null/closure boundaries.
Both preserve the host branch's readonly-array and Node filesystem checks.
The scratch branch was not pushed to the library's branch.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/host/25_readDirectory.a$' -count=1 -timeout 30m > /tmp/generic-array-host25-final.log 2>&1
```

Observation: the refusal pass now gets past flatten's 1097:26 array view, and
stops at 1108:34 in toSorted:

```typescript
export function toSorted<T>(array: readonly T[], comparer?: Comparer<T>): SortedReadonlyArray<T> {
    return (array.length === 0 ? emptyArray : array.slice().sort(comparer)) as readonly T[] as SortedReadonlyArray<T>;
}
```

The next diagnostic refuses never viewed as T, whose constraint any can be
written (adamic/invariant-mutable). This is a separate emptyArray fallback gap.
Fixture 25 still does not finish lowering, so no native/JavaScript equality or
25/25 claim is made. Its source Node baseline remains exit 0 with the four
readDirectory output lines recorded in GENERIC_FUNCTION_VALUE_ARMS.md.

Final uncached oracle package passes in 125.397s. The final scratch test exits 1
in 0.693s solely at Lower's next 1108:34 refusal. The preceding implementation's
oracle run also passed uncached in 139.446s; the final run supersedes it.
