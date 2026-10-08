# Contextual generic function operands

October 8, 2026.

Built: checker-driven specialization of generic function operands through
conditionals, parentheses, ||, ?? and &&, with lazy function logical lowering.
Implementation: f5c4115f and e3e1aa9a; main merge 88862ee9 brings 74fb6490.
Validation: lowerer and oracle packages, Node comparisons, counts, vet, formatting,
and pinned host-proof scratch integration; final results below.
Mutants: direct-only specialization, swapped || operands, inverted && condition.
Limits: existing polymorphic-value and function-identity boundaries remain;
fixture 25 still refuses a separate mutable generic-array view; the full
repository gate and all 25 host fixtures are not rerun here.

The original refusal compared identity's unspecialized T parameter against
string in the function-variance check. Each operand now asks the pinned checker
for its contextual type and instantiated signature. The same resolved signature
feeds function specialization and the relation check. No checker implementation
is copied. The linkname shims reference the existing cohere dependency.

Callable context is queried for indirect operands without widening the context
rules for non-callable casts, destructuring and literals. A preliminary broad
context query failed the counts gate on existing enum, destructuring and empty
array fixtures. Restricting this path to callable destinations restores them.
Counts regeneration changes only the new fixture row: 18 allocations, 18 frees,
67 retains, 95 releases, peak 7, regions 0.

The fixture generic_function_value_arms.a covers the reported conditional both
ways, nested conditionals, parentheses, ||, ??, &&, and number/string instances.
Side-effect counters prove one left-operand evaluation and lazy fallback calls.
Function || reuses Coalesce; && reuses Conditional and IsUndefined. Functions
are truthy when their nullable closure pointer is present. Both backends share
the existing IR; no backend edits are needed.

The permanent operand-context test asks the checker for each identity operand's
signature, proving seven string contexts and one number context. The direct-only
mutant is caught there as a missing concrete operand signature. Swapping || or
inverting && is caught by stdout differences from source Node in both backends,
with no compiler-warning or sanitizer kill accepted as evidence. The runner
restores source files in finally blocks.

Setup succeeded: Go 0.023s, Node 0.024s, submodules 0.065s, markdown 0.067s,
clang 0.155s, Go build 32.231s, cache warm 32.407s, total 32.433s.
nproc=5, cgroup quota 4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8.
Environment: /workspace/adamic-tools/env.sh. Logs: /tmp/generic-arms-*.log.

Commands (all test output redirected to logs):

```sh
bash cloud/setup.sh > /tmp/generic-arms-setup.log 2>&1
source /workspace/adamic-tools/env.sh
node --input-type=module-typescript < internal/oracle/testdata/generic_function_value_arms.a > /tmp/generic-arms-node.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/generic-arms-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/generic-arms-gate.log 2>&1
python3 internal/lower/testdata/run-generic-function-arm-mutants.py > /tmp/generic-arms-mutants.log 2>&1
go vet ./... > /tmp/generic-arms-vet.log 2>&1
gofmt -l cmd internal > /tmp/generic-arms-fmt.log
git diff --check
```

Scratch branch codex/generic-function-value-host-scratch merges pinned host
proof c97402ba with this implementation. Merge 83b3fe9d preserves both the host
branch's readonly-array fallback checks and the new callable contextual path.
The scratch worktree references the identical pinned cohere submodule through
a symlink; its pinned stage3/api dependencies were installed with npm ci
--ignore-scripts. A temporary oracle registration runs actual fixture
stage3/fixtures/host/25_readDirectory.a against source Node, the JavaScript
backend, native release, sanitized malloc/slab builds and leak checking:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/host/25_readDirectory.a$' -count=1 -timeout 30m > /tmp/generic-arms-host25.log 2>&1
```

The scratch branch is local, never pushed to the library branch. The temporary
registration is not part of the feature branch. Previous host 24/25 status is
the user's observation; this unit directly tests fixture 25, not all 25.

## Results and the remaining host stop

Full uncached touched-package gate: lower 63.956s, oracle 254.686s, both pass.
Counts: oracle 18.428s, pass. Repository vet and formatting: exit 0, no output.
All three new source mutants were caught again after the main merge. The merge
changes stage1 and evidence, not the lowerer/oracle implementation.

The actual scratch fixture **does not complete lowering**. Its former
1088:12 identity conditional passes; the next refusal is at 1097:26:

```text
a value of type NonNullable<T> seen as T, a type parameter whose constraint
any can be written, so it can write what NonNullable<T> can't hold
(adamic/invariant-mutable)
```

The retained source is:

```typescript
export function flatten<T>(array: T[][] | readonly (T | readonly T[] | undefined)[]): T[] {
    const result = [];
    for (let i = 0; i < array.length; i++) {
        const v = array[i];
        if (v) {
            if (isArray(v)) {
                addRange(result, v);
            } else {
                result.push(v);
            }
        }
    }
    return result;
}
```

Observation: the evolving result array is inferred from truthy pushes as
NonNullable<T>[], then addRange receives it through writable T[]. The existing
reverse mutable-element relation refuses that view before monomorphization.
Inference: for a concrete string instantiation the distinction collapses, but
allowing the view for an unbound T without rechecking each instance could hide
an undefined-admitting instantiation. A separate sound fix must defer this
specific generic obligation and recheck it under the concrete mapper. This
unit does not suppress the check. Scratch oracle exits 1 in 0.586s at Lower;
therefore no fixture-25 native/JavaScript equality or 25/25 claim is made.
The temporary scratch registration remains available for that follow-up.

After merging current main, the generic function-value focused packages pass:
lower 0.426s, oracle 0.638s. Both original and new function-value fixtures run
against source Node with both backends and the native oracle variants.

Source Node for actual fixture 25 exits 0 and prints:

```text
a.ts,z.ts,alias/b.ts,alias/deep/c.ts
a.ts,z.ts
src/b.ts,src/deep/c.ts,a.ts
plain.js
```

New fixture source Node output, matched by both backends:

```text
KIRK
kirk
AHRA
ahra
NESTED
nested
nested
OR
or
NULLISH
nullish
absent
AND
first
SECOND
third
FOURTH
missing
SIXTH
6,3
40,41
```
