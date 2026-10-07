# What stage 0 couldn't lower in cohere's gitignore matcher

The port beside this file is cohere's `internal/gitignore` (gitignore.go and glob.go) written as 0.1 Adamic. Every place stage 0 refused it is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number, so when a gap closes, the place to undo it can be found with `grep -n "gap N" *.ts`.

Each program here typechecks under stage 0's options and runs on Node 24.21.0 with the output shown. Stage 0 refuses it with `NotYet` at the place shown: first at main 7055396, and checked again at 80c3098. None of them reached clang, and none compiled wrong.

The programs are also in `gaps/`, and `gaps_test.go` holds them to this file. An open gap must still be refused with the words recorded here. A closed one must lower and print natively what it prints on Node, leaking nothing. So the stream that closes a gap sees this test fail, and the message tells it to mark the gap closed and undo the port's workaround.

Every gap is closed. Gaps 5, 6 and 7 closed when the language-gaps stream landed on main at 80c3098, and this stream closed 1, 2, 3, 4, 8 and 9 in stage 0 itself, each with an oracle fixture and a mutant it catches. Each time the port went back to writing those places as the Go does, so it now reads side by side with the Go, with no workaround left. Closed gaps stay below, marked so, and the record is whole.

Found while closing these, and closed since by this stream (with P2's list): generic functions, instantiated once per set of what their type arguments are held as; `return undefined` from a function returning `string | undefined` (or an array, a function, an object), which reached clang as bad C; `?.length` on an array or a string; and spread arguments to `Math.max`, `Math.min`, `Math.hypot`, `String.fromCharCode` and `String.fromCodePoint`. Still open, and refused rather than miscompiled: a call through `?.` (`text?.toUpperCase()` segfaulted natively on main), a spread argument to any other call, a generic function as a value, `.length` or an array's methods on a tuple, and a tuple where an array goes.

## 1. `String.fromCharCode` and `String.fromCodePoint` (closed by this stream)

Both are in docs/0.1.md's library.

```ts
console.log(String.fromCharCode(104));
```

```
stage 0 can't lower reading String yet        (Node prints h)
```

`String.fromCodePoint(233)` is refused the same way.

**Around it, until it closed:** a 256-character string constant read by index. **Closed:** both lower, each argument a number evaluated in order (runtime/from_codes.c). `fromCharCode` is ToUint16 exactly; adjacent surrogates join into the character they make, as in JavaScript; `fromCodePoint` panics with V8's `RangeError: Invalid code point ...`. Fixtures: `from_codes.a` (a sweep of ToUint16's edges and code points against Node) and `from_code_point_fails.a`. A spread argument isn't lowered yet, as no call's is (below).

## 2. The bitwise operators (closed by this stream)

All of them, though 0.1 has them with JavaScript's ToInt32 semantics: `>>`, `<<`, `>>>`, `&`, `|`, `^` and unary `~`.

```ts
const x = 200;
console.log(`${x >> 3}`);
```

```
stage 0 can't lower a BinaryExpression with a number and a number yet      (Node prints 25)
stage 0 can't lower a PrefixUnaryExpression on a number yet                 (for ~x)
```

**Around it, until it closed:** arithmetic for every shift and mask. **Closed:** all seven, and the six compound assignments, through ToInt32 and ToUint32 exactly (runtime/bitwise.c), with no C conversion of an out-of-range value and no shift of a negative one. Lowering them first printed `true` for `200 >> 3`: `ir.Binary.Type` typed every operator it didn't list as a boolean. It now names both lists and panics as a compiler bug for an operator in neither. Fixture: `bitwise.a`, every pair of 36 edge values through all six binary operators against Node.

## 3. A call to a function or method declared later (closed by this stream)

This is the one that changed the port's shape most. A call to a function declared further down the module, or to a method declared further down the class, is lowered before the callee's body. Its return type isn't known yet, so the call is taken for a void one.

```ts
function first(value: number): number {
	return second(value) + 1;
}
function second(value: number): number {
	return value * 2;
}
console.log(`${first(3)}`);
```

```
stage 0 can't lower a void call used as a value yet        (Node prints 7)
```

The same happens for methods:

```ts
class Counter {
	count = 0;
	twice(): number {
		return this.once() + this.once();
	}
	once(): number {
		this.count = this.count + 1;
		return this.count;
	}
}
console.log(`${new Counter().twice()}`);
```

```
stage 0 can't lower a void call used as a value yet        (Node prints 3)
```

The cause is in `declareModule` (internal/lower/lower.go): it records every function with an empty `ir.Function` and then lowers the bodies in source order, so a call reads `Returns` before the callee's body has set it. The checker already knows every declared return type, so the signature could be taken from it up front. That would also make mutual recursion possible, which no order can give today.

**Around it, until it closed:** every file of the port was ordered callee first, so the files no longer read top to bottom in the Go's order. **Closed:** `signature` (internal/lower/lower.go) now writes every function's and method's parameters and result from the checker before any body is lowered: `declareModule` signs every module function first, and `instantiate` every method before the constructor or any method body. Mutual recursion compiles. Fixture: `internal/oracle/testdata/declared_later.a`. The port is back in the Go's order.

## 4. A declared function used as a value (closed by this stream)

```ts
function isEven(value: number): boolean {
	return value % 2 === 0;
}
const test: (value: number) => boolean = isEven;
console.log(`${test(4)}`);
```

```
stage 0 can't lower reading isEven yet        (Node prints true)
```

An arrow function in the same place lowers.

**Around it, until it closed:** `inClass = (character) => isLetter(character)` for the Go's `in = isLetter`. **Closed:** a module function read as a value lowers to a function value whose code forwards its arguments to the function (`functionValue`, internal/lower/expression.go), made once per function. One with an optional, default or rest parameter isn't made yet, since a function value is called with exactly the arguments its caller has. Fixture: `declared_later.a`. The port writes `inClass = isLetter` again.

## 5. `Array.from` (closed at 80c3098)

```ts
const members = Array.from({ length: 4 }, () => false);
console.log(`${members.length}`);
```

```
stage 0 can't lower reading Array yet        (Node prints 4)
```

**Around it, until it closed:** `new Array<number>(256).fill(0)`, which lowered. It held numbers because of gap 6. The port now writes `Array.from({ length: 256 }, () => false)`.

## 6. An element of a `boolean[]` (closed at 80c3098)

The Go's sets are `[256]bool`, and its position sets are bit sets. The natural port is a `boolean[]`, and reading one is refused:

```ts
const flags: boolean[] = [true, false];
console.log(`${flags[0] === true}`);
```

```
stage 0 can't lower an index into an array of booleans (boolean | undefined) yet        (Node prints true)
```

Taken into a variable first, `const first = flags[0]`, it reads `stage 0 can't lower a value of type boolean | undefined yet`.

**Around it, until it closed:** the byte sets and the position sets were `number[]`, with 1 for a member and 0 for anything else, which needed the gap 7 workaround as well. They are `boolean[]` again.

## 7. Comparing `number | undefined` with a number (closed at 80c3098)

```ts
const maybe: number | undefined = [1][0];
console.log(`${maybe === 1}`);
```

```
stage 0 can't lower a BinaryExpression with a value and a number yet        (Node prints true)
```

`string | undefined` against a string lowers (`words[0] === 'a'`). The number case does not.

**Around it, until it closed:** `(set[member] ?? 0) === 1`. Nothing in the port needs it now.

## 8. A tuple as a value (closed by this stream, and on cloud/collections-fs)

A tuple literal anywhere except directly inside `new Map([...])` was refused: returned, assigned, or as an element of a constant array. Stream C2 closed it: a tuple written out is a value wherever the checker types it a tuple, and `const [a, b] = tuple` (and `let`) destructures one, at a module's top level too.

```ts
function cut(text: string): readonly [string, number] {
	return [text.slice(1), 1];
}
const [rest, count] = cut('abc');
console.log(`${rest} ${count}`);
```

```
stage 0 can't lower a value of type [string, number] where an array goes yet        (Node prints bc 1)
```

```ts
type Pair = readonly [string, number];
const pairs: Pair[] = [['a', 1]];
```

```
stage 0 can't lower a value of type [string, number] where an array goes yet
```

The Go returns two results in several places: `strings.CutPrefix`'s `(after, found)`, and `(bool, Source)` from `Ignored` and `Excluded`. A tuple is how TypeScript says that.

**Around it, until it closed:** named fields. `Cut { after, found }` (strings.CutPrefix's own result names), and `Verdict { ignored, source }` (cohere's differential test names the pair `verdict`). While the cases were constants, a tree's entries were written `new Map<string, Entry>([...])`, the one place tuple literals lower. case.ts now builds the map with `set` as it reads them.

**Closed** twice, on this branch and on cloud/collections-fs, and joined when main 4ff4657 was merged: this branch's lowering of a tuple literal and of `const [a, , c] = tuple` is kept, since it also fills a left-out optional element with undefined and lowers `[a, b] = tuple`, and main's `.length` of a fixed-length tuple and its JavaScript (a tuple written as an array) are kept with it. An array literal the checker types as a tuple is an object of its elements, `"0"`, `"1"`, ..., as Map entries already were; a destructuring takes the tuple whole first, then each name from its field, read as the element's type and fitted to the name's. Fixture: `tuple_values.a`. Still refused rather than read as missing fields: an array's methods on a tuple, and a tuple where an array goes, since a tuple isn't an array yet. Holes in a pattern (`[, b]`) crashed stage 0 on main, in `for...of` too: typescript-go's hole is a binding element with no name.

## 9. `return panic(...)` (closed by this stream)

Found when the driver began reading its cases at run time. `panic` returns `never`, so TypeScript writes the end of a function whose `switch` didn't return as `return panic(...)`. Stage 0 lowers `x ?? panic(...)`, and `panic(...)` as a statement, but not `panic(...)` as a returned value.

```ts
import { panic } from 'adamic';
function kindOf(kind: string): number {
	switch (kind) {
		case 'one':
			return 1;
	}
	return panic(`no kind ${kind}`);
}
console.log(`${kindOf('one')}`);
```

```
stage 0 can't lower reading panic yet        (Node prints 1)
```

**Around it, until it closed:** `panic(...)` as a statement. **Closed:** `return panic(...)` lowers as the panic itself, and so does an arrow whose body is `panic(...)`. A function whose result the checker types `never` has no result to hold, as one returning `void` hasn't; on main, an arrow returning `panic(...)` crashed stage 0 with a nil dereference, pointing a refusal at the name an arrow doesn't have. Fixtures: `return_panic.a` (where it doesn't run) and `return_panic_fires.a` (where it does).

## Not a stage 0 gap: input

When this port began, 0.1 had no input, so the Go's reads through `os` (`os.Lstat`, `os.Stat` and `os.ReadFile` of ignore files and `.git` entries) became reads of a `WorkingTree` the program is given, and the cases were constants. `readTextFile` and `programArguments` have since landed on main. The driver now reads its cases from the file its argument names (`main.ts`, and the format in `case.ts`), so the port compiles once, whatever it's asked. Reading a real tree also needs lstat, stat, symbolic links and directory listing, which 'adamic' doesn't have yet. When it does, a `WorkingTree` read from disk can take the in-memory one's place, and nothing else in the port changes.

## Bytes, read in place

A Go string is its bytes, and git matches bytes: `?` against `é` is false, because `é` is two bytes. 0.1's strings are UTF-16 to a program. The port first spelled each pattern and each text as a "byte string", one character per byte (bytes.ts), which cost a conversion of every text the glob matched and was most of the distance to Go. Now `utf8Length(text)` and `utf8At(text, index)` from 'adamic' read a string's UTF-8 in place (runtime/utf8.c; docs/0.1.md's library), and the glob reads its pattern and its text byte by byte, as glob.go does; bytes.ts is gone. Natively nothing is copied, since a string is stored as UTF-8: a lone surrogate, stored as WTF-8, reads as U+FFFD's three bytes, as TextEncoder writes it, so both backends agree. Fixtures: `utf8_view.a` and `utf8_view_fails.a`. The test's `bytes` tree holds the glob's byte semantics to git.

## What lowered as written

What lowered as written, and is worth saying so: discriminated unions narrowed by `switch` (the Go's token is one), classes with `#private` fields written only while `enter` builds a new matcher, a generic `Result<Value>` union for the Go's `(value, error)`, `?? panic(...)` for every index the Go would bounds-check, closures assigned in a `switch`, `for...of` over strings by code point, a `Map` of a discriminated union, string `switch` throughout the glob compiler, a `Map<string, Matcher>` of entered directories for the walk, and the driver's parser over `readTextFile`'s text (`split`, `for...of` by code point, a union-typed `let` for the section it's in).

## Performance, observed (not refusals)

### At a0c4027, with halves joined only where pieces meet, and main 4ff4657 merged

Concatenation now looks for a surrogate pair's halves only where two pieces meet, and main brings integrations 7 and 8 (Perceus reuse and the rest). Same case set, best of seven at a load average under 1, all three outputs byte for byte the same: Go cohere 0.11 s, **the native port 0.236 s** (was 0.252 s), Node 0.392 s. **Go cohere is 2.1 times faster than the native port.** callgrind: 2.19 billion instructions, down from 2.34 billion. Two things moved them, and they're apart in the profile: `adamic_string_join_halves`, 76 million before, is gone from it (this branch), and `adamic_release` fell from 458 million to 378 million (the merge; releases belong to borrow inference). The top is now releases 17%, slice 8%, malloc and free 15%, and the glob's own `lastMatching` and `matches`.

### At 15e0b5f, with shared slices, appends in place and indexOf in place

Main 1d72913 merged. A slice of 64 bytes or more, and a quarter of its owner, reads its owner's bytes; `text += more` on a local appends in place when the local holds the only reference; `indexOf` from a position no longer slices (docs/memory.md, "Strings, specifically"). The same case set as below, best of seven at a load average under 1, all three outputs byte for byte the same:

| | time | peak memory |
|---|---|---|
| Go cohere | 0.11 s | not measured apart from `go test` |
| the native port | 0.252 s (was 0.295 s) | 10 MB |
| Node running the port's source | 0.373 s | 95 MB |

**Go cohere is now 2.3 times faster than the native port** (2.7 before this, 20 this morning). callgrind puts 2.34 billion instructions in the native run, down from 2.82 billion: `adamic_string_concat` is off the top of the profile and `adamic_string_slice` is down from 241 million to 177 million. What's on top now: `adamic_release` 20% (458 million, the same count as before, so it's the number of releases, not their cost, that a borrow inference would cut), malloc and free together 14%, and `adamic_string_join_halves` 3%, concat's scan of every byte for a surrogate pair's halves, which needs to look only where two pieces meet.

### At 7017551, with the glob reading bytes in place

Main 4ddd17f merged, and the glob on `utf8Length` and `utf8At` instead of bytes.ts. The case set has grown, with reviewer R2's cases: 55 trees, 6,360 tree paths, 21 pattern lists and 1,825 globs, 14,963 answer lines. Timed best of seven on the same container at a load average under 1, with all three outputs byte for byte the same:

| | time | peak memory |
|---|---|---|
| Go cohere (as below) | 0.11 s | not measured apart from `go test` |
| the native port (`adamic build`, `-O2`, reading the cases file) | 0.295 s | 10 MB |
| Node running the port's source | 0.370 s | 95 MB |

**Go cohere is now 2.7 times faster than the native port, where it was 6.6 times faster.** The native port beats Node by a quarter, in a ninth of its memory. callgrind puts 2.8 billion instructions in the native run, down from 7.4 billion, and none of the top of the profile is reading bytes any more. It's allocation now: `adamic_release` 16%, `adamic_string_concat` 11%, `adamic_string_slice` 9%, and malloc and free together 19%. Those are the walk's and the matcher's path strings, built by `+` and `slice` and released one by one, where the Go slices a string without copying it. A slice that shares its parent's bytes would close most of what's left. Also observed, and not in this case set's time: `text += character` in a loop is quadratic natively (each `+=` copies), where it is linear on Node; the case parser's `unescape` hit it on a 100 MiB file and now joins slices instead.

### On merged main at ddfed83, with every gap closed

Main now carries stream C's constant-time string index (integrate-2), and the port no longer works around anything. The test's case set (53 trees, 6,318 tree paths each asked twice, by path and by walk, 21 pattern lists and 421 globs; 13,464 answer lines) was timed best of seven on the same 4-vCPU container, at a load average under 1, with all three outputs byte for byte the same:

| | time | peak memory |
|---|---|---|
| Go cohere (cohere_side_test.go's answer step, reading the trees from disk and the cases as JSON) | 0.11 s | not measured apart from `go test` |
| the native port (`adamic build`, `-O2`, reading the cases file) | 0.730 s | 10 MB |
| Node running the port's source | 0.800 s | 94 MB |

**Go cohere is now 6.6 times faster than the native port, where it was 20 times faster.** The native port is a little faster than Node and uses a ninth of Node's memory. callgrind puts 7.4 billion instructions in the native run, down from 26.4 billion. 56% of them are under `utf8Bytes`: before the glob matches a text, it spells the text as bytes, and its first step is a `charCodeAt` loop asking whether the text is ASCII. Each step is constant time now, but it's a call and a few conversions per character, once for every text the glob matches. That's the bytes cost below, and it's most of the distance to Go, which gets a string's bytes for free. A byte view of a string, or `TextEncoder`, would close most of it; so would the glob converting only past its literal and suffix fast paths, as the Go's never converts at all.

### Before

These lowered and answered correctly, but they're the numbers to beat on the way to "faster and leaner than Go cohere". All measured on this Linux container, x86-64, with the test's case set, which includes cohere's own checkout as a real tree: 50 trees, about 6,200 tree paths, 21 pattern lists and 421 globs, and output identical on every side.

- **Go cohere is 20 times faster than the native port.** The same 50 trees, pattern lists and globs, answered byte for byte the same, take Go cohere 0.10 s (the whole `go test` of cohere_side_test.go's answer step, reading the trees' ignore files from disk and its cases as JSON). The native port takes 2.1 s, and Node 0.62 s. The rest of this section is where the native time goes.
- **Native is 3.5 to 5 times slower than Node.** With the cases compiled in as constants, the unsanitized `-O2` binary ran in 2.06 s and Node in 0.42 s. Read from a 447 KB cases file, it is 2.09 s and 0.59 s. Under callgrind (run on the constants build), 42% of the native run's instructions are in `adamic_string_length` and 39% in `adamic_string_char_code_at` (internal/native/runtime/string.c). Each of them walks the string's UTF-8 from its first byte. So `text.length`, `text.charCodeAt(index)` and `text[index]` cost the length of the string, and a loop over a string's indexes is quadratic. The glob reads its pattern and its text that way, as the Go does. That's natural code, and every 0.1 program that reads a string by index pays the same. docs/memory.md already names the fix: an ASCII-only flag, so the common case is a load. Smallest program that shows the shape:

  ```ts
  const text = 'a'.repeat(100000);
  let count = 0;
  for (let index = 0; index < text.length; index++) {
  	if (text.charCodeAt(index) === 97) {
  		count++;
  	}
  }
  console.log(`${count}`);
  ```

  Built with `adamic build` (`-O2`), it takes 11.1 s natively. Node runs it in 0.085 s.

  In the port, 78% of the native run's instructions are under `utf8Bytes` (bytes.ts). Before it spells a text as bytes, it checks whether the text is ASCII by this very loop, and the glob calls it once for every text it matches. With `length` and `charCodeAt` constant time, the loop is a few instructions per character.

- **Constant data compiles slowly.** When the cases were constants (7,360 lines of TypeScript), they became 3.4 MB of C, which clang took 39 s to compile at `-O2`, and longer under the sanitizers. With cohere's checkout among the trees, the test took five and a half minutes. The driver now reads them at run time, and the same test takes 27 s. It would still cost any program with a large table in its source.

## Path cleaning (added by stream P2)

`path.ts`'s `clean` split every path at its slashes and joined it again, so it copied even a path already clean, which is most of what it is asked. In the format walk (stage1/cohere/formatfiles), which cleans every path it builds, that was 40% of native's instructions. Go's `path.Clean` returns a clean path as itself.

First try, Go's way: one pass over the units into a lazy buffer that copies nothing until the output first differs from the input, comparing by `charCodeAt`. It answered as Go does, and the walk took 4.72 s where it had taken 2.7 s. Under callgrind, `adamic_string_char_code_at` was 47% of the run: each call goes out of line, and on a short string holding anything but ASCII it walks the UTF-8 from the start (mediaquery's GAPS.md, the third performance shape), about 200 instructions a unit. In stage 0 today a loop over a string's units, written in the program, loses to the string library's own calls, which run in C.

So `clean` asks first whether the path is clean, with the library's scans (`includes('//')`, `startsWith('./')`, `endsWith('/..')` and the rest, `isClean`), returns it as it is when it is, and takes it apart only when it isn't. `dir` cleans what comes before the last slash without that slash, which cleans to the same path and then is usually clean already. The walk takes 2.26 s (from 2.7), and this slice's own run is unchanged within noise (0.66 to 0.77 s, from 0.71 to 0.73; Node 0.77 to 0.80). Of what remains in `clean`, most is `adamic_string_index_of`, which compares at every position with a call to `memcmp`.

`path_test.go` holds `clean`, `base` and `dir` to Go's `path.Clean`, `path.Base` and `path.Dir` on 5,040 paths, natively, on Node and through the JavaScript backend: every string of up to seven units over `/`, `.` and `a`; every sequence of up to six of `/`, `.`, `..` and `a`; every string of up to four over `/`, `.`, `é`, `😀` and `.a`; and long ones by hand. Six mutants, each taking one test out of `isClean` or moving `dir`'s cut, are caught natively and on Node. The lazy buffer's four mutants were caught too, while it was the code.

