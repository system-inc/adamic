# Shared empty arrays through readonly views

October 8, 2026.

Built: readonly conditional array destinations retain the checker's contextual
view; never has no readable values, while writable array invariance stays intact.
Commit: 809f155f; pinned host scratch merge 343445be contains c97402ba.
Commands: lowerer/native packages, uncached oracle, counts, Node, vet, formatting.
Mutants: writable acceptance, readonly refusal, lost context/layout, and three
independent ownership regressions, all caught as detailed below.
Limits: fixture 25 now stops at 1150:32; the full repository gate and all 25
host fixtures were not rerun. No 25/25 claim is made.

## Target at 1108:34

The source is:

```typescript
export function toSorted<T>(array: readonly T[], comparer?: Comparer<T>): SortedReadonlyArray<T> {
    return (array.length === 0 ? emptyArray : array.slice().sort(comparer)) as readonly T[] as SortedReadonlyArray<T>;
}
```

SortedReadonlyArray extends ReadonlyArray. A checker inspection of the actual
pinned host fixture gives:

```text
operand 25_readDirectory.a:1108:34
conditional inferred T[]
conditional context readonly T[]
readonly=true
```

The arm was incorrectly judged against its parent's inferred mutable T[] even
though the cast gives it a readonly contextual destination. Conditional array
arms now query the parent's checker context before falling back to that inferred
type. A shared never[] may be read as readonly T[]; it has no readable elements.
Writable element comparisons still run the reverse relation before recursive
reading comparisons, so never[] cannot acquire inhabited writable elements.
This is the existing invariance rule, not a new variance exception.

The empty never[] literal has real array identity and zero slots. It uses an
empty array allocation with no reference elements; it does not create any fake
value of T or copy the shared constant when returning a readonly view.

## Fixtures and ownership

readonly_never_array.a holds module-level emptyArray: never[], generic returns
of readonly T[] and ReadonlyArray<T>, the reported parenthesized readonly assertion
shape, string and number instances, and an object-constrained generic. Each
empty and populated branch is compared to source Node in both backends.

A typed copy of an empty readonly view can legally become writable. A probe
using runtime-built strings exposed an ASan heap-use-after-free in concat and
missing ownership in slice-and-push/splice. The empty array's original numeric
storage flag was inherited by a populated string copy. Native first writes now
set the ownership flag when the array has zero elements, after evaluating all
operands. Concatenation chooses the first populated input's ownership convention.
No element representation, identity, or ownership of a populated array is changed.
The fixture holds slice/push, concat and splice against Node and sanitizers,
then verifies the original shared never[] remains empty. No IR or JavaScript
backend change was needed. Protected emit.go, lower.go, native.go and
oracle_test.go were not edited.

Source Node exits 0 and prints:

```text
0
Kirk
0
4,5
0
Ahra
0
owned:slice
owned:concat
owned:splice
0
0
```

writable_never_array_refused.a tries to return shared emptyArray as writable T[].
Its source Node prints 0 and exits 0; Adamic refuses before backend emission:

```text
main.a:3:12: Adamic 0.1 refuses a never[] seen as writable T[], which can write T into a shared never[]; declare the result readonly T[], or return a fresh [] (adamic/invariant-mutable)
```

The lowerer test pins Refused, the complete What and Fix, and the line/column.

## Mutants

Run testdata/run-readonly-never-array-mutants.py. Each source is restored in a
finally block; each run writes a separate /tmp/readonly-never-mutant-*.log.

| Mutant | Catcher |
|---|---|
| Accept writable never[] by skipping the reverse element comparison | Pinned refusal test: got <nil>, the unsafe view incorrectly lowered |
| Reject readonly never values by removing the bottom-type read rule | Oracle's object-constrained generic refuses during Lower |
| Ignore a conditional's checker array context | Oracle refuses the asserted readonly fallback during Lower |
| Remove slotless empty never[] literal layout | Oracle stops at array of never during Lower |
| Lose first-push ownership | LeakSanitizer on the runtime-built slice string |
| Lose first-splice ownership | LeakSanitizer on the runtime-built inserted string |
| Take concat ownership from its empty prefix | ASan heap-use-after-free in adamic_array_join |

No compiler-warning kill is accepted. The writable mutant is caught entirely
in lowering, without invoking clang. The original required writable refusal is
independent of the runtime fixes.

## Commands and setup

With /workspace/adamic-tools/env.sh sourced, test output is redirected to logs:

```sh
bash cloud/setup.sh > /tmp/readonly-never-setup.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/readonly_never_array.a > /tmp/readonly-never-node.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/lower/testdata/writable_never_array_refused.a > /tmp/readonly-never-writable-node.log 2>&1
python3 internal/lower/testdata/run-readonly-never-array-mutants.py > /tmp/readonly-never-mutants.log 2>&1
go test ./internal/lower ./internal/native -count=1 -timeout 30m > /tmp/readonly-never-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/readonly-never-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/readonly-never-counts.log 2>&1
go vet ./... > /tmp/readonly-never-vet.log 2>&1
gofmt -l cmd internal > /tmp/readonly-never-fmt.log
git diff --check
```

Setup succeeded: Go/Node 0.023s each, submodules 0.058s, markdown 0.066s,
clang 0.151s, Go build 33.314s, cache warm 33.523s, total 33.549s.
nproc=5, cgroup quota 4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8.
Current origin/main remains 74fb6490 and is already merged into this branch.
Counts pass in 40.309s; only the new row changes: 31 allocations, 31 frees,
29 retains, 57 releases, peak 9, regions 0. Vet and formatting pass with no output.

## Actual host fixture's next stop

On local scratch branch codex/generic-function-value-host-scratch, which contains
pinned c97402ba and the final implementation:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/host/25_readDirectory.a$' -count=1 -timeout 30m > /tmp/readonly-never-host25.log 2>&1
```

The test exits 1 in 0.341s at Lower. The 1108:34 refusal is cleared; the next
stop is 1150:32:

```typescript
const includeFileRegexes = patterns.includeFilePatterns && patterns.includeFilePatterns.map(pattern => getRegexFromPattern(pattern, useCaseSensitiveFileNames));
```

Diagnostic: readonly string[] | undefined seen as RegExp[] | undefined, which
can write RegExp where string is read (adamic/invariant-mutable). This unit
records that next logical-array contextual-type stop without suppressing it.
Fixture 25 still does not finish lowering, so its native/JavaScript equality is
not certified. The scratch branch was not pushed to the library branch.

Final package results:

```text
ok github.com/system-inc/adamic/internal/lower 15.969s
ok github.com/system-inc/adamic/internal/native 181.929s
ok github.com/system-inc/adamic/internal/oracle 186.640s
```

The oracle run is uncached. Current main 74fb6490 is an ancestor of the final
branch. Working source was restored after every mutant; final diff checks pass.
