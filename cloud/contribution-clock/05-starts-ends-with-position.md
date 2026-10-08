# startsWith and endsWith with a position are refused

**Today on origin/main 74fb6490a.** This program:

```ts
console.log(`${'hello'.startsWith('ll', 2)} ${'hello'.endsWith('he', 2)}`);
```

is refused:

```
adamic: r5.a:1:16: stage 0 can't lower startsWith with these arguments yet
```

`endsWith` with a second argument is refused in the same way (`endsWith with these arguments`).

**Node prints** `true true`.

**Where.** internal/lower/object.go: `stringMethods` at line 1515 (`"startsWith"` and `"endsWith"` at lines 1526 and 1527 take one argument), and `func (l *lowering) stringCall` at line 1542, which refuses at line 1550. The runtime's `adamic_string_starts_with` and `adamic_string_ends_with` take no position (internal/native/emit_strings.go lines 90 to 93), so lower the position away instead of touching the runtime C:

- `s.startsWith(search, position)` is `s.slice(clamp(position)).startsWith(search)`.
- `s.endsWith(search, end)` is `s.slice(0, clamp(end)).endsWith(search)`.

Here `clamp(x)` is `min(max(ToIntegerOrInfinity(x), 0), s.length)`. Clamp before slicing, because slice reads a negative index from the end. internal/lower/library_string.go already has every piece: `stringHelper` at line 234 (it evaluates each operand once into a helper's parameters), `stringInteger` at line 248 (NaN to 0, then trunc), and `stringIndexMethod` at line 252, whose clamp is the model. Write a sibling of `stringIndexMethod`. The helper's `Returns` and the `ir.Call`'s `Returns` must be `ir.Boolean`, because `stringHelper` defaults them to `ir.String`. Make both shapes `{ir.String, ir.Number}` with one optional argument, and dispatch to the new helper from the switch after line 1573 when two arguments are given.

**The fixture.** `internal/oracle/testdata/library_string_starts_ends_position.a`, registered by `internal/oracle/library_string_starts_ends_position_test.go`:

```ts
const text = 'héllo, wörld';
const positions = [-5, 0, 1, 2, 7, 8, 11, 12, 99, NaN, Infinity, -Infinity, 2.9];
for (const position of positions) {
	console.log(`${position} ${text.startsWith('é', position)} ${text.endsWith('é', position)} ${text.startsWith('wö', position)} ${text.endsWith('hé', position)} ${text.startsWith('', position)}`);
}
console.log(`${'😀x😀'.startsWith('x', 2)} ${'😀x😀'.endsWith('x', 3)} ${'abc'.startsWith('abcd', 0)} ${'abc'.endsWith('c', 2)}`);
```

Node prints:

```
-5 false false false false true
0 false false false false true
1 true false false false true
2 false true false true true
7 false false true false true
8 false false false false true
11 false false false false true
12 false false false false true
99 false false false false true
NaN false false false false true
Infinity false false false false true
-Infinity false false false false true
2.9 false true false true true
true true false false
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_string_starts_ends_position.a' -count=1` passes.
- Mutant "unclamped start": pass the raw position to slice. The -5 row then prints `true` for `startsWith('wö', -5)`, which disagrees with Node.
- Mutant "position dropped": ignore the second argument. The 1, 2, 7 and 2.9 rows then disagree.
- `go test ./internal/lower` passes.

**Size.** About 20 lines plus the fixture. Confidence medium-high. It is the largest brief here, but it copies an existing helper's shape.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
