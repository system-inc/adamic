# A value compared with itself emits C that clang refuses (task 7c6b4pq)

**Today on origin/main 74fb6490a.** This program:

```ts
function main(): void {
	const zeros = new Array<number>(5).fill(0);
	console.log(`${zeros.fill(99) === zeros}`);
}
main();
```

fails to build:

```
adamic: native: clang failed: exit status 1
main.c:15:112: error: self-comparison always evaluates to true [-Werror,-Wtautological-compare]
```

`fill` returns its receiver, so both operands lower to the same C name, and the emitter prints `(adamic_local_0_zeros == adamic_local_0_zeros)`. It is not only arrays. In the fixture below, on 74fb6490a, a boolean, a Map and a closure compared with themselves are refused in the same way. `point === point` builds there, because `point` is captured by `read` and is read through its cell, but it stays in the fixture as a guard. Numbers are not affected: clang does not warn on a double compared with itself, because NaN differs from itself.

**Node prints** `true`.

**Where.** internal/native/emit_expressions.go, `func (e *emitter) binary` at line 633. Its last line, line 655, `return fmt.Sprintf("(%s %s %s)", left, cOperators[operator], right)`, is the fall-through that prints the tautology. The caller is the `case ir.Binary:` at line 56. When `left == right` (the same C text), the operator is `ir.Equal` or `ir.NotEqual`, and the operand type is not `ir.Number`, emit a constant that still evaluates the operand once, for example `((void)(x), true)` or `((void)(x), false)`. Identical C text on both sides means both are plain reads, because calls are already hoisted into uniquely named temporaries. The `ir.Number` case has to stay a real `==` for NaN. The string, maybe and union cases above line 655 already go through helpers and are not affected.

**The fixture.** `internal/oracle/testdata/self_comparison.a`, registered by `internal/oracle/self_comparison_test.go`:

```ts
// A value compared with itself: one C name on both sides, which clang refuses as a tautology
// unless the operands are numbers (NaN differs from itself).
function compare(): void {
	const zeros = new Array<number>(5).fill(0);
	console.log(`${zeros.fill(99) === zeros} ${zeros !== zeros.fill(1)}`);
	const flag = zeros.length > 2;
	const point = { x: 1 };
	const table = new Map<string, number>();
	const read = (): number => point.x;
	console.log(`${flag === flag} ${point === point} ${table !== table} ${read === read}`);
	const missing = Number('x');
	const count = zeros.length;
	console.log(`${missing === missing} ${count === count} ${missing !== missing}`);
}
compare();
```

Node prints:

```
true false
true true false true
false true true
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/self_comparison.a' -count=1` passes.
- Mutant "self-compare fall-through": delete the new `left == right` branch. The fixture then fails with clang's -Wtautological-compare error.
- Mutant "NaN folded": drop the `operandType != ir.Number` guard. The `missing === missing` column then prints `true`, which disagrees with Node.
- `go test ./internal/native` passes (allowing for the host failures noted at the top).

**Size.** About 5 lines plus the fixture. Confidence high.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
