# Binary lowering evidence

Logical operator rule family: **1,202 table sites**, counted from `sites.csv`: 874 `&&`, 328 `||`. This is operator-family coverage on the saved checker-rejected census, not a measured whole-program success count or a claim that all 1,202 still failed at the newer compiler base. Some narrower logical shapes were already supported. The two required real-site replays demonstrate actual before/after progress on this branch.

Logical selection reuses the operand-selection IR used by coalescing. `Logical` selects by existing JavaScript ToBoolean support rather than by undefined presence. Both operands are represented alike before selection, and the checker-proven result is projected afterward. The left runs once; the right runs only on its branch. Existing freshness and flow analyses already merge both operand aliases. Native emission keeps an unselected falsy reference's temporary until statement cleanup: an empty string or boxed zero can still own counted memory. No condition admission changes.

Six fixtures cover the five table operand rows, each exercising both operators, plus primitive boxed unions and strings. They include undefined, false, zero, negative zero, NaN, infinities, runtime-built strings, object identity, fresh return values, right-side effects and a thrown right operand. Source Node, JavaScript backend Node, release native, ASan, UBSan and LeakSanitizer agree.

Commands, all output directed to logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/binary_logical' -count=1 -timeout 15m > /tmp/notyet-binary-logical-final.log 2>&1
python3 stage3/notyet-binary/run-logical-mutants.py > /tmp/notyet-binary-logical-mutants.log 2>&1
go test ./internal/lower -run TestMixedLogicalBoxedNullBoundary -count=1 > /tmp/notyet-binary-null-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts > /tmp/notyet-binary-logical-counts.log 2>&1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir > /tmp/notyet-binary-vet.log 2>&1
```

Final oracle passes in 0.944s, boundary test in 0.033s, counts refresh in 22.240s. Vet exits 0 with no output. Counts changes are only the six new fixture rows. The initial value/value fixture was rejected by TS2872 for direct always-truthy object literals; replacing those probes with function-produced objects fixed the fixture. Its corrected focused oracle and the final six-fixture oracle pass.

| Mutant | Catcher |
| --- | --- |
| Remove `&&`'s falsy-left branch test | Source Node versus native stdout |
| Invert `||`'s branch test | Source Node versus native stdout |
| Treat NaN as truthy | Source Node versus native stdout |
| Treat a present empty string as truthy | Source Node versus native stdout |
| Transfer the owned left temporary even when it is falsy | LeakSanitizer: 88 bytes in two allocations, the boxed zero and empty string |
| Emit `&&` for JavaScript backend `||` | Source Node versus backend stdout |
| Remove boxed-null boundary | `TestMixedLogicalBoxedNullBoundary`, unexpected successful lowering |

Every mutant exits 1 at the named check; none is counted as a build-error kill. The runner restores the source after each mutation. The null mutant disables the guard with `if false && common == ir.Union ...`, then runs the named boundary test; its log is preserved separately. A mixed logical selection that would box null alongside undefined remains NotYet, because the runtime has no distinct boxed-null tag. This is a representation limitation, not a new design refusal.

Both exact original-signature replay commands from README were run again after lowering. Each exits 1 because the original signature no longer reproduces. Complete JSON and stderr are in `evidence/after-*.json` and `evidence/after-*.log.txt`:

| Original table example | Next named stop |
| --- | --- |
| `binder.ts:423:18` | `binder.ts:432:37`: NotYet `for...of over an object` |
| `binder.ts:1698:21` | `core.ts:656:1`: Refused `overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] &#124; undefined` |

These selected units remain measurement-only; no native or JavaScript output is claimed for the complete tsc project.

## Strict equality across reference representations

The second group covers **60 original `!==` sites**. Each has differently represented references, notably `NodeArray<T>` versus `readonly T[]`. Both operands are boxed into the existing union representation and compared using its existing strict equality: strings by value, references by identity, scalar members by both value and type. Undefined retains its distinct missing value. Nothing is coerced to a number for equality.

`binary_equality_references.a` covers the new object/array comparison with each side independently missing, both missing, distinct references with the same length, and controls for shared array identity. It also pins equal runtime-built strings, unequal strings, scalar type distinctions, NaN, negative zero and union/scalar equality. It passes source Node, both backends and sanitizers.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/binary_equality' -count=1 -timeout 15m > /tmp/notyet-binary-equality-oracle.log 2>&1
python3 stage3/notyet-binary/run-equality-mutants.py > /tmp/notyet-binary-equality-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts > /tmp/notyet-binary-equality-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/binary_(logical|equality)' -count=1 -timeout 15m > /tmp/notyet-binary-final-oracle.log 2>&1
go test ./internal/lower -run TestMixedLogicalBoxedNullBoundary -count=1 > /tmp/notyet-binary-final-lower.log 2>&1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir > /tmp/notyet-binary-final-vet.log 2>&1
```

Equality oracle passes in 0.462s. The wrong-right-operand normalization mutant and boxed-pointer-equality mutant both exit 1 at source-Node/native stdout comparison. Neither fails compilation. Counts refresh passes in 22.097s and adds only the new equality fixture row. The final seven-fixture uncached oracle passes in 0.979s; the boundary test passes in 0.033s; final vet exits 0 with no output. Mutant logs are preserved with `.log.txt` names under `evidence/`, and both runners restore the source.

## Coverage and remaining work

The two implemented operator families cover **1,262 of 1,452 original sites**: 1,202 logical selections and 60 strict reference inequalities. This is a raw-CSV operator-family count, not a measured net decrease at the newer compiler pin or a count of complete tsc bodies that now compile. Both requested example sites show actual lowering progress to their next named stops.

The remaining **190 sites use `=` as an expression**: 87 reference/reference and 103 number/number. They remain compiler lessons. Returning the right operand without performing the store would silently miscompile. Lowering these needs expression-level stores with their evaluation order and mutation effects represented in freshness, flow and ownership analyses. That separate IR capability was not added here; the original NotYet stops remain. There are no arithmetic operators in these five raw rows, so no arithmetic ToNumber rule was added or claimed.

The existing documented October 7 ruling already admits non-boolean conditions; this unit changes no admission rule. Mixed selection requiring a boxed null distinct from undefined remains the explicit NotYet boundary proven by its mutant. A distinct null tag needs representation work or a ruling on the scope of that representation. No existing design refusal was removed.

No protected orchestration/emitter/native/oracle file was edited. No code was copied from cohere. No whole-package tests or full repository gate were run. The required counts refresh was the named TestCountsAreRecorded test. The reports and lowering groups were pushed separately to `codex/notyet-binary`, with no pull request and no push to main or an area branch.

Delivery main is `6f16a1693ff41bc102c4d9bfac83b6330277c479`, verified against the remote. It is already an ancestor of this branch; the final merge reports `Already up to date.`

## Assignment continuation

The subsequent assignment-value group is recorded in [ASSIGNMENTS.md](ASSIGNMENTS.md), including 43 fixtures, ten killed mutants, exact before/after replays, touched-package tests, 190 original operator-family sites and the known array-length target boundary. The earlier 190-site deferral above describes the logical/equality commits, not the current branch.

Current continuation and delivery status, including the c41c0e06 merge, exact per-group coverage limits, all mutants and pending conditional ownership, are in [FINAL.md](FINAL.md). Older delivery/deferred-work statements describe their respective commits.
