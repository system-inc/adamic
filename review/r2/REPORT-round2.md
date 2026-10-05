# Stream R2, round two: P's gap closures and P2's media query slice

Reviewed `origin/cloud/20d7ptt` at 09342fd (stream P: signatures before bodies, declared functions as values, return panic, fromCharCode and fromCodePoint, the bitwise operators, tuples as values) and `origin/claude/port-cohere-second-slice-050g6a` at e7612ee (stream P2: the media query parser). Nothing on either branch or on main was changed. The probes are in `review/r2/p/`, and each runs with `./review/r2/run3.sh <probe>` from a checkout of P's branch (copy `review/r2` into it), three ways as before: the source on Node, native under ASan and UBSan with leaks checked, and the JavaScript backend on Node. Each probe below that fails on P's branch is NotYet on main: `adamic js` there refuses it, so every finding here is new with the branch.

## Confirmed findings, P's branch

### 1. A tuple read as an array: a silent wrong answer (tuples as values)

The checker lets a tuple go where a readonly array of its elements goes. On P's branch a tuple is an object of fields "0", "1", ..., and nothing stops it from arriving where an array is expected, so the array code reads an object.

`tuple_length_silent.a`:

```ts
const pair: [number, number] = [3, 4];
const measure = (values: readonly number[]): number => values.length;
console.log(`length ${measure(pair)}`);
```

```
== node     exit=0  length 2
== native   exit=0  length 94844204055328
== backend  exit=0  length undefined
```

ASan says nothing, because the read stays inside the object. The optimized build (`adamic build`) printed `length 94120448228576`, exit 0. Both backends are wrong, and differently from each other. `tuple_through_closure.a` (join and indexing through the same function value, and a tuple inside `(readonly number[])[]`) is a heap-buffer-overflow under ASan in `adamic_array_join`, and a segfault (139) in the optimized build. `tuple_as_array.a` (a tuple passed straight to a named function's `readonly number[]` parameter) reaches clang as C it refuses (`incompatible pointer types passing 'adamic_object *' to parameter of type 'adamic_array *'`), which the doctrine says a NotYet must never do.

Main refuses the same programs with "stage 0 can't lower a value of type [number, number] where an array goes yet". The branch's tuple literal (`tupleLiteral`, made an object wherever the checker or the context says tuple) gets past that guard. A fix: wherever a tuple-typed value meets an array-typed slot (an argument, an assignment, an element, a closure's parameter, a return), refuse it as main does, or convert it to an array there.

### 2. A declared function read as a value twice isn't the same function (declared functions as values)

`function_identity.a`: `const first = double; const second = double; first === second`.

```
== node     true 2,4,6 10
== native   false 2,4,6 10
== backend  false 2,4,6 10
```

Exit 0 on all three sides. `functionValue` makes the forwarder function once, but each read is a new `MakeClosure`, which allocates a new closure natively and a new `AdamicClosure` in the backend. So `===`, and `includes` or `indexOf` of a function in an array, answer differently from JavaScript. A fix: make each forwarder's closure once, as an immortal static (it captures nothing), and have every read return that one. Or refuse `===` on function values until then.

### 3. A constructor calling a method declared below it, which reads a field not yet set (signatures before bodies)

Once every method's signature is known before the constructor is lowered, a constructor can call a method declared below it. TypeScript's definite-assignment check doesn't follow `this.field` reads inside a called method, so the method can read a field the constructor hasn't set yet. In JavaScript that field is `undefined`.

`constructor_number_later.a`:

```ts
class Scaled {
	readonly doubled: number;
	readonly base: number;
	constructor(base: number) {
		this.doubled = this.twice();
		this.base = base;
	}
	twice(): number {
		return this.base * 2;
	}
}
console.log(`${new Scaled(21).doubled}`);
```

```
== node     exit=0  NaN
== native   exit=0  0
== backend  exit=0  0
```

Both backends give a silently wrong answer: a number slot not yet written reads as 0 rather than undefined. With a string field (`constructor_calls_later.a`) Node panics with `TypeError: Cannot read properties of undefined (reading 'length')`. Native is a UBSan null-pointer access in string_index.c under the sanitizers, and a bare segfault (139) in the optimized build. The backend panics as Node does.

This is the class-field version of the module dead zone, which the branch gets right. `later_reads_global.a` and `later_reads_number.a` call a later-declared function that reads a `const` not yet initialized, and all three sides panic with the same ReferenceError. A fix along the same lines: a field read inside a method called from the constructor before every field is assigned would need the same dead-zone check (JavaScript's answer is undefined, not an error, so it's a read of undefined, which a `number` slot can't hold). The conservative fix is to refuse a call to a method from a constructor before every field is assigned, with the fix "assign the fields first".

### 4. A tuple literal leaving out an optional element: native panics as a compiler bug (tuples as values)

`tuple_workout.a`, last lines: `const optional: [number, string?] = [5]; console.log(`${optional[1] ?? 'none'}`)`. Node and the backend print `none`. Native exits 70 with `adamic: panic: compiler bug: a field the checker proved is there is missing`. It's loud, not silent, but the program is valid. The literal's own type is `[number]`, so `tupleLiteral` builds one field, and the read of "1" through the declared `[number, string?]` finds none. Everything else in that probe matched all three ways: swaps by destructuring assignment, destructuring a returned tuple and reassigning it, sorting `[string, number][]`, a tuple as a Map value, a hole (`const [, second]`).

## What turned out fine on P's branch

- **The bitwise operators.** `bitwise_sweep.a`: `~`, `&`, `|`, `^`, `<<`, `>>`, `>>>` over 39 values (±2^31, ±2^32, ±2^53 and past, 1e21, the largest double, the smallest subnormal, NaN, both infinities, -0, fractions, shift counts past 31 and negative), 1,521 pairs. All three sides match byte for byte, and -0 comes out +0 from every operator. **Mutant:** `>>` without its sign extension (`shifted |= 0;` in bitwise.c) changed native's output only, so the sweep caught it.
- **fromCharCode and fromCodePoint.** `from_codes_sweep.a`: ToUint16's wrap (-1, 65601, 1.9, NaN, 1e21), a high and a low half that meet across arguments and across concatenation, `fromCodePoint` of lone halves, U+10FFFF and -0, the empty calls, and indexing into the results. All matched.
- **Order of evaluation.** `order_of_arguments.a`: arguments with effects to fromCharCode and fromCodePoint, and as both operands of `<<`, `-`, `|` and `>>>`, plus all six bitwise compound assignments. All left to right on all three sides.
- **Functions as values across modules.** `modules_value/main.a`: an imported function passed to `map` and to a function parameter. Matched.
- **Dead zones through later-declared functions.** See finding 3: right for module globals.
- **A function that never returns, used as a value.** `never_in_coalesce.a` (`ages.get(k) ?? fail('why')` with `fail(why): never`) is NotYet ("a void call used as a value"). That's safe, but it's the commonest use of such a helper, and worth closing.

## P2's media query slice: does the test hold the port?

Much better than the gitignore test did. With the branch's files laid into a main checkout (the slice touches nothing outside stage1/ and THIRD_PARTY_NOTICES.md), the test passes: 4,103 cases (3,961 parsed, 142 refused), every tree the same from Go cohere, natively, on Node and through the backend. Its five mutants are all caught. I added seven (`review/r2/mediaquery/mutants.go.txt`, `python3 review/r2/mediaquery/apply.py <repository>`; the log is `review/r2/mediaquery/run.txt`):

| Mutant | Result |
|---|---|
| the last word after a media type typed a keyword (parsers.ts) | caught |
| the third of four words typed a media type (parsers.ts) | caught |
| a curly brace not entering a level (parsers.ts) | caught |
| U+2029 not whitespace (whitespace.ts) | **survives**, natively and on Node |
| U+2000 not whitespace (whitespace.ts) | **survives**, natively and on Node |
| DEL written as itself in the output's quotes (main.ts) | **survives** |
| U+001F written as itself (main.ts) | **survives** |

The parser's logic is held. Its character sets aren't: the generator's pieces hold U+2003 and U+200A from the U+2000 to U+200A range, and U+2028 but not U+2029, so no seed can reach the others. The same gap means the comment in whitespace.ts ("stage 0's trim is then held to Node's on every whitespace character above") is only true for the characters in the pieces. Two of the survivors are in the driver's quoting, not the port, but a wrong quote would hide a wrong value. The fix is cheap: add every one of `\s`'s 25 characters to `adamicPieces` (U+2000 to U+200A whole, and U+2029), plus `\x7f` and `\x1f`.

A note for anyone running it from a worktree: the test overlays cohere by its path, so when `cohere` is a symbolic link, go resolves the real directory, the overlay never applies, cohere's side test is never compiled, and the test fails with "answers-true.txt: no such file or directory". That cost me a run. It isn't a fault in the test as it runs in the repository.

## What I didn't cover

- P's rewritten gitignore port itself (glob.ts and gitignore.ts reworked onto the closed gaps). Its test is unchanged on that branch, so round one's five surviving mutants should still survive there. I didn't rerun them.
- P's speed measurements (09342fd) and the method-read refusal (ddfed83): read, not probed.
- Generic functions as values, and functions with default or rest parameters as values: the branch says NotYet for the second, and I didn't try the first.
- P2's comparison with the library itself (`ADAMIC_MEDIA_QUERY_LIBRARY`): skipped, no npm install here.
- macOS: everything ran on Linux x86_64.
