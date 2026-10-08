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
