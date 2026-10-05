# What stage 0 couldn't lower in cohere's gitignore matcher

The port beside this file is cohere's `internal/gitignore` (gitignore.go and glob.go) written as 0.1 Adamic. Every place stage 0 refused it is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number, so when a gap closes, the place to undo it can be found with `grep -n "gap N" *.ts`.

Each program here typechecks under stage 0's options and runs on Node 24.21.0 with the output shown. Stage 0 refuses it with `NotYet` at the place shown: first at main 7055396, and checked again at 80c3098. None of them reached clang, and none compiled wrong.

The programs are also in `gaps/`, and `gaps_test.go` holds them to this file. An open gap must still be refused with the words recorded here. A closed one must lower and print natively what it prints on Node, leaking nothing. So the stream that closes a gap sees this test fail, and the message tells it to mark the gap closed and undo the port's workaround.

Three of the first eight closed when the language-gaps stream landed on main at 80c3098 (gaps 5, 6 and 7), and the port went back to writing those places as the Go does. They stay below, marked closed, so the record is whole. The other five are open, and so is gap 9, found when the driver began reading its cases.

## 1. `String.fromCharCode` and `String.fromCodePoint`

Both are in docs/0.1.md's library.

```ts
console.log(String.fromCharCode(104));
```

```
stage 0 can't lower reading String yet        (Node prints h)
```

`String.fromCodePoint(233)` is refused the same way.

**Around it:** the port spells a string's UTF-8 as a byte string (bytes.ts, below), so it makes a character from a byte. It reads the character from a 256-character string constant, `'\x00\x01...\xff'`, by index. That lowers.

## 2. The bitwise operators

All of them, though 0.1 has them with JavaScript's ToInt32 semantics: `>>`, `<<`, `>>>`, `&`, `|`, `^` and unary `~`.

```ts
const x = 200;
console.log(`${x >> 3}`);
```

```
stage 0 can't lower a BinaryExpression with a number and a number yet      (Node prints 25)
stage 0 can't lower a PrefixUnaryExpression on a number yet                 (for ~x)
```

**Around it:** arithmetic. UTF-8's six-bit groups are `Math.floor(codePoint / 64 ** group) % 64` (bytes.ts, `sixBits`). The Go's case folding `character|0x20` becomes two ranges in `isLetter` and in the `xdigit` class (glob.ts). `100 << 20` becomes `100 * 1024 * 1024`. The `%q` escape's hex digits are `Math.floor(code / 16)` and `code % 16`.

## 3. A call to a function or method declared later

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

**Around it:** every file of the port is ordered callee first. Each declaration still names the Go function it reads as, but the files no longer read top to bottom in the Go's order. `(*Matcher).decide` sits ahead of `enter`, and `(*glob).reach` ahead of `matches`.

## 4. A declared function used as a value

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

**Around it:** the Go's `in = isLetter` (glob.go, addClass) is written `inClass = (character) => isLetter(character)`, and the same for `isDigit`.

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

## 8. A tuple as a value (closed on cloud/collections-fs)

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

## 9. `return panic(...)`

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

**Around it:** `panic(...)` as a statement (case.ts, `entryOf`). The checker accepts that, since nothing after a `never` call is reachable.

## Not a stage 0 gap: input

When this port began, 0.1 had no input, so the Go's reads through `os` (`os.Lstat`, `os.Stat` and `os.ReadFile` of ignore files and `.git` entries) became reads of a `WorkingTree` the program is given, and the cases were constants. `readTextFile` and `programArguments` have since landed on main. The driver now reads its cases from the file its argument names (`main.ts`, and the format in `case.ts`), so the port compiles once, whatever it's asked. Reading a real tree also needs lstat, stat, symbolic links and directory listing, which 'adamic' doesn't have yet. When it does, a `WorkingTree` read from disk can take the in-memory one's place, and nothing else in the port changes.

## Not a gap, a cost: bytes

A Go string is its bytes, and git matches bytes: `?` against `é` is false, because `é` is two bytes. 0.1's strings are UTF-16, and it has no byte array or encoder (no `Uint8Array`, no `TextEncoder`). So the glob spells its pattern and its text as byte strings, one character per byte (bytes.ts), and then reads as the Go does. The test's `bytes` tree holds this to git. The cost is a conversion of every non-ASCII text the glob matches; ASCII is its own byte string and isn't copied. A byte type, or `TextEncoder`, would remove the conversion.

## What lowered as written

What lowered as written, and is worth saying so: discriminated unions narrowed by `switch` (the Go's token is one), classes with `#private` fields written only while `enter` builds a new matcher, a generic `Result<Value>` union for the Go's `(value, error)`, `?? panic(...)` for every index the Go would bounds-check, closures assigned in a `switch`, `for...of` over strings by code point, a `Map` of a discriminated union, string `switch` throughout the glob compiler, a `Map<string, Matcher>` of entered directories for the walk, and the driver's parser over `readTextFile`'s text (`split`, `for...of` by code point, a union-typed `let` for the section it's in).

## Performance, observed (not refusals)

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

