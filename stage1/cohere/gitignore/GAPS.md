# What stage 0 couldn't lower in cohere's gitignore matcher

The port beside this file is cohere's `internal/gitignore` (gitignore.go and glob.go) written as 0.1 Adamic. Every place stage 0 refused it is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number, so when a gap closes, the place to undo it can be found with `grep -n "gap N" *.ts`.

Each program here typechecks under stage 0's options and runs on Node 24.21.0 with the output shown; stage 0 (main at 7055396) refuses it with `NotYet` at the place shown. None of them reached clang, and none compiled wrong.

Two of the eight are already being filled by another stream (`Array.from`, and `boolean | undefined`, gaps 5 and 6). The rest are new.

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

## 5. `Array.from` (being filled by the language-gaps stream)

```ts
const members = Array.from({ length: 4 }, () => false);
console.log(`${members.length}`);
```

```
stage 0 can't lower reading Array yet        (Node prints 4)
```

**Around it:** `new Array<number>(256).fill(0)`, which lowers. It holds numbers because of gap 6.

## 6. An element of a `boolean[]` (being filled by the language-gaps stream, as `boolean | undefined`)

The Go's sets are `[256]bool`, and its position sets are bit sets. The natural port is a `boolean[]`, and reading one is refused:

```ts
const flags: boolean[] = [true, false];
console.log(`${flags[0] === true}`);
```

```
stage 0 can't lower an index into an array of booleans (boolean | undefined) yet        (Node prints true)
```

Taken into a variable first, `const first = flags[0]`, it reads `stage 0 can't lower a value of type boolean | undefined yet`.

**Around it:** the byte sets and the position sets are `number[]`, with 1 for a member and 0 for anything else (glob.ts, `ByteSet` and the position-set functions). That means the gap 7 workaround as well.

## 7. Comparing `number | undefined` with a number

```ts
const maybe: number | undefined = [1][0];
console.log(`${maybe === 1}`);
```

```
stage 0 can't lower a BinaryExpression with a value and a number yet        (Node prints true)
```

`string | undefined` against a string lowers (`words[0] === 'a'`). The number case does not.

**Around it:** `(set[member] ?? 0) === 1`.

## 8. A tuple as a value

A tuple literal anywhere except directly inside `new Map([...])` is refused: returned, assigned, or as an element of a constant array.

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

**Around it:** named fields. `Cut { after, found }` (strings.CutPrefix's own result names), and `Verdict { ignored, source }` (cohere's differential test names the pair `verdict`). The cases hold a tree's entries as a `ReadonlyMap` written `new Map<string, Entry>([...])`, the one place tuple literals lower.

## Not a stage 0 gap: input

When this port began, 0.1 had no input, so the Go's reads through `os` (`os.Lstat`, `os.Stat` and `os.ReadFile` of ignore files and `.git` entries) are reads of a `WorkingTree` the program is given. `readTextFile` from 'adamic' has since landed on main. Reading a real tree also needs lstat, stat, symbolic links and directory listing, which 'adamic' doesn't have yet. So the cases stay constants, written by the test into `cases.ts`. When the matcher can stat, a `WorkingTree` read from disk can take the in-memory one's place, and nothing else in the port changes.

## Not a gap, a cost: bytes

A Go string is its bytes, and git matches bytes: `?` against `é` is false, because `é` is two bytes. 0.1's strings are UTF-16, and it has no byte array or encoder (no `Uint8Array`, no `TextEncoder`). So the glob spells its pattern and its text as byte strings, one character per byte (bytes.ts), and then reads as the Go does. The test's `bytes` tree holds this to git. The cost is a conversion of every non-ASCII text the glob matches; ASCII is its own byte string and isn't copied. A byte type, or `TextEncoder`, would remove the conversion.

## What lowered as written

What lowered as written, and is worth saying so: discriminated unions narrowed by `switch` (the Go's token is one), classes with `#private` fields written only while `enter` builds a new matcher, a generic `Result<Value>` union for the Go's `(value, error)`, `?? panic(...)` for every index the Go would bounds-check, closures assigned in a `switch`, `for...of` over strings by code point, a `Map` of a discriminated union, and string `switch` throughout the glob compiler.

## Performance, observed (not refusals)

These lowered and answered correctly, but they're the numbers to beat on the way to "faster and leaner than Go cohere". All measured on this Linux container, x86-64, with the case set that includes cohere's own checkout as a real tree (`ADAMIC_GITIGNORE_COHERE_TREE=1`): 50 trees, 6,212 tree paths, 21 pattern lists and 421 globs, and output identical on every side.

- **Native is 5 times slower than Node.** The unsanitized `-O2` binary ran in 2.06 s, and Node ran the source in 0.42 s. Under callgrind, 42% of the native run's instructions are in `adamic_string_length` and 39% in `adamic_string_char_code_at` (internal/native/runtime/string.c). Each of them walks the string's UTF-8 from its first byte. So `text.length`, `text.charCodeAt(index)` and `text[index]` cost the length of the string, and a loop over a string's indexes is quadratic. The glob reads its pattern and its text that way, as the Go does. That's natural code, and every 0.1 program that reads a string by index pays the same. docs/memory.md already names the fix: an ASCII-only flag, so the common case is a load. Smallest program that shows the shape:

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

- **Constant data compiles slowly.** The cases are constants (7,360 lines of TypeScript). They become 3.4 MB of C, which clang takes 39 s to compile at `-O2`, and longer under the sanitizers. That's why the default test asks about this repository as its only real tree, and cohere's checkout only when `ADAMIC_GITIGNORE_COHERE_TREE` is set: with it, the test takes about five minutes. Reading the cases at run time, now that `readTextFile` has landed, would compile the port once.
