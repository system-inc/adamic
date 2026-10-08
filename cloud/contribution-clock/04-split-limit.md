# String split with a limit is refused

**Today on origin/main 74fb6490a.** This program:

```ts
console.log('a,b,c,d'.split(',', 2).join('|'));
```

is refused:

```
adamic: r4.a:1:13: stage 0 can't lower split with these arguments yet
```

**Node prints** `a|b`.

**Where.** internal/lower/object.go. The `stringMethods` table at line 1515 gives `"split"` one required argument at line 1528. `func (l *lowering) stringCall` at line 1542 refuses any other count at line 1550. A RegExp split already takes a limit (internal/lower/regexp.go line 194 onward, through the runtime). For a string separator, the spec's split with a limit is the split without one, then `.slice(0, ToUint32(limit))`. Lowering can express that with no runtime change. Make the shape `{ir.String, ir.Number}` with one optional argument. In the switch after line 1573, when `split` has two arguments, return an `ir.ArraySlice` whose `Array` is the one-argument `ir.StringCall` split. Its arguments are `0` and `ir.Binary{Operator: ir.ShiftRightUnsigned, Left: limit, Right: 0}`, which is ToUint32. Both emitters already handle `StringCall` split (internal/native/emit_strings.go line 64) and `ArraySlice`. Optionally, update docs/0.1.md line 129, which lists `split(separator: string)`.

**The fixture.** `internal/oracle/testdata/library_string_split_limit.a`, registered by `internal/oracle/library_string_split_limit_test.go`:

```ts
const text = 'a,b,c,d';
const limits = [0, 1, 2, 3, 10, -1, 2.7, NaN, Infinity, 4294967297];
for (const limit of limits) {
	const parts = text.split(',', limit);
	console.log(String(limit) + ' ' + String(parts.length) + ' [' + parts.join('|') + ']');
}
console.log('héllo wörld'.split('', 3).join('/'));
console.log(String(''.split(',', 5).length) + ' ' + String(''.split('', 5).length));
const kept = 'x-y-z'.split('-', 2);
kept.push('w');
console.log(kept.join('+'));
```

Node prints:

```
0 0 []
1 1 [a]
2 2 [a|b]
3 3 [a|b|c]
10 4 [a|b|c|d]
-1 4 [a|b|c|d]
2.7 2 [a|b]
NaN 0 []
Infinity 0 []
4294967297 1 [a]
h/é/l
1 0
x+y+w
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_string_split_limit.a' -count=1` passes, including the leak check on the intermediate array.
- Mutant "limit not ToUint32": slice to the raw limit instead of `limit >>> 0`. The rows for -1, Infinity and 4294967297 then disagree with Node.
- Mutant "limit ignored": return the one-argument split. Most rows then disagree.
- `go test ./internal/lower` passes.

**Size.** About 6 lines plus the fixture. Confidence high.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
