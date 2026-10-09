# The global isNaN and isFinite are refused although Number.isNaN and Number.isFinite lower

**Today on origin/main 74fb6490a.** This program:

```ts
const value = Number('x');
console.log(`${isNaN(value)} ${isFinite(value)}`);
```

is refused:

```
adamic: r3.a:2:16: stage 0 can't lower reading isNaN yet
```

**Node prints** `true false`.

**Where.** internal/lower/object.go, `func (l *lowering) builtin` at line 573. Line 594 already routes the global `parseInt` and `parseFloat` to `numberCall` (line 763), because they are the same functions as Number's. TypeScript's lib types the global `isNaN(number: number)` and `isFinite(number: number)`, so tsc already rejects a string argument (TS2345), and the global behaves the same as `Number.isNaN` on every argument the checker lets through. Add `isNaN` and `isFinite` to the `isLibraryGlobal` check on line 594, and update its comment. `numberCall` keeps refusing a non-number argument as it does today. Optionally, list the globals beside `Number.isNaN` and `Number.isFinite` in docs/0.1.md line 128.

**The fixture.** `internal/oracle/testdata/library_math_number_global_is_nan.a`, registered by `internal/oracle/library_math_number_global_is_nan_test.go`:

```ts
const values = [0, -0, 1.5, NaN, Infinity, -Infinity, 0 / 0, Number('x'), Number.MAX_VALUE * 2];
for (const value of values) {
	console.log(String(value) + ' ' + String(isNaN(value)) + ' ' + String(isFinite(value)) + ' ' + String(isNaN(value) === Number.isNaN(value)));
}
const half = values.filter((value: number): boolean => !isNaN(value) && isFinite(value)).length;
console.log(String(half));
```

Node prints:

```
0 false true true
0 false true true
1.5 false true true
NaN true false true
Infinity false false true
-Infinity false false true
NaN true false true
NaN true false true
Infinity false false true
3
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_math_number_global_is_nan.a' -count=1` passes.
- Mutant "global isFinite unrouted": drop `isFinite` from the check. Lowering then refuses with `reading isFinite`.
- Mutant "isFinite as isNaN": route `isFinite` to `numberCall(node, "isNaN")`. stdout then disagrees with Node.

**Size.** 1 to 3 lines plus the fixture. Confidence very high. It may be smaller than a clock slot wants.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
- Counts. A new oracle fixture also needs its row in internal/oracle/counts.md, or TestCountsAreRecorded fails: run `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts` and commit only that fixture's row.
- Whole packages. Don't run a whole package (`go test ./internal/lower`, `./internal/native` and the like): the fast gate runs every touched package in about 20 s after your push and sends any red back to you with its first failure. Run your fixture, your tests and each mutant, refresh counts.md, then push.
