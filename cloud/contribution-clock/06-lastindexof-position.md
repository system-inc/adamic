# lastIndexOf on a string with a position is refused

**Today on origin/main 74fb6490a.** This program:

```ts
console.log(`${'abcabc'.lastIndexOf('b', 3)} ${'abcabc'.lastIndexOf('c', NaN)}`);
```

is refused:

```
adamic: r6.a:1:16: stage 0 can't lower lastIndexOf with these arguments yet
```

**Node prints** `1 5`.

**Where.** internal/lower/object.go: `stringMethods` line 1529 (`"lastIndexOf"` takes one argument), and `func (l *lowering) stringCall` at line 1542, which refuses at line 1550. The runtime's `adamic_string_last_index_of` takes no position (internal/native/emit_strings.go line 71). Lower it as `s.slice(0, min(clamp(p) + search.length, s.length)).lastIndexOf(search)`, using the same `stringHelper` approach as brief 5, with the result type `ir.Number`. This is correct for an empty search, too.

The trap is that this method's NaN is not 0. `lastIndexOf` reads a NaN position, or a missing one, as +Infinity, so it searches from the end. So the position here is `isNaN(p) ? Infinity : trunc(p)`, not `stringInteger(p)`. If brief 5 has landed, add this as a case in its helper.

**The fixture.** `internal/oracle/testdata/library_string_last_index_of_position.a`, registered by `internal/oracle/library_string_last_index_of_position_test.go`:

```ts
const text = 'héllo, héllo';
const positions = [-5, 0, 1, 2, 7, 8, 11, 12, 99, NaN, Infinity, -Infinity, 2.9];
for (const position of positions) {
	console.log(`${position} ${text.lastIndexOf('héllo', position)} ${text.lastIndexOf('', position)} ${text.lastIndexOf('l', position)}`);
}
console.log(`${'abcabc'.lastIndexOf('c', 4)} ${'abcabc'.lastIndexOf('abc', 2)} ${'abc'.lastIndexOf('abcd', 9)} ${'😀x😀'.lastIndexOf('😀', 2)}`);
```

Node prints:

```
-5 0 0 -1
0 0 0 -1
1 0 1 -1
2 0 2 2
7 7 7 3
8 7 8 3
11 7 11 10
12 7 12 10
99 7 12 10
NaN 7 12 10
Infinity 7 12 10
-Infinity 0 0 -1
2.9 0 2 2
2 0 -1 0
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_string_last_index_of_position.a' -count=1` passes.
- Mutant "NaN as zero": use `stringInteger(p)`. The NaN row then prints `NaN 0 0 -1`, which disagrees with Node.
- Mutant "end not widened": slice to `clamp(p)` without adding `search.length`. A match that starts at or before the position but ends after it is then lost. For example, the 0 row then prints -1 for 'héllo', and most rows disagree.
- `go test ./internal/lower` passes.

**Size.** About 12 lines on top of brief 5's helper, or about 25 lines alone, plus the fixture. Confidence medium-high.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
- Counts. A new oracle fixture also needs its row in internal/oracle/counts.md, or TestCountsAreRecorded fails: run `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts` and commit only that fixture's row.
