# What stage 0 couldn't lower in cohere's media query parser, and a bug in Go cohere

The port beside this file is cohere's `internal/format/css/mediaquery` (index.go, nodes.go, parsers.go and whitespace.go), itself a port of postcss-media-query-parser 0.2.3, written as 0.1 Adamic. It is stage 1's second slice; the first is `stage1/cohere/gitignore`, whose GAPS.md this one follows and refers to.

Every place stage 0 refused the port, or took it and made C that clang refused, is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number (`grep -n "gap N" *.ts`). Each program typechecks under stage 0's options and runs on Node 24.21.0 with the output shown. They were found at main 2385966.

The programs are in `gaps/`, and `gaps_test.go` holds them to this file: an open gap must still be refused with the words recorded here (or, for gap 2, still make the C clang refuses), and a closed one must lower and print natively what it prints on Node, leaking nothing.

## Why this slice

Chosen from cohere's internal packages for being self-contained and held by tests cohere already has:

- **format/css/mediaquery** (chosen): 589 lines of Go, with no regular expressions, generators, goroutines or unsafe, and a standard library surface of `errors`, `strings` and `unicode/utf8`. cohere holds it to the JavaScript library it came from with an oracle test: 95 inline fixtures, golden trees and refusals. It is a real piece of Prettier's CSS path (the params of every `@media` and `@custom-media`), and a parser over strings, which is what most of cohere is.
- **format/css/values** (runner-up): postcss-value-parser, a tokenizer and a parser, about 1,100 lines, also with an oracle test against the library. Twice the size and the same shape, so it's the natural third slice. selector, its neighbour, needs regexp.
- **lint/ecmascript/text** (runner-up): the edit distance and grapheme counting the lint rules use. Pure and small, but held only by unit tests, and too small to tell us much.
- Looked at and left: format/yaml/cst returns iter.Seq; format/doc's string width matches emoji with regexp; the markdown entity tables are data, and constant data compiles slowly (gitignore's GAPS.md).

## 1. A generic function isn't instantiated per call (closed by stream P's generic functions (293976f, a1f90d0), in integration 11)

Closed: stage 0 now instantiates a generic function once per concrete type argument, and both gap
programs compile and print natively what they print on Node. Their gaps-test entries carry only the
expected stdout, which is how that test records a closed gap.

Before this closed, stage 0 lowered a generic function only once, so a return or parameter typed by
its type parameter was refused even though docs/0.1.md specifies monomorphization.

```ts
function identity<Item>(item: Item): Item {
	return item;
}
console.log(identity('a'));
```

```
stage 0 can't lower a function returning Item yet        (Node prints a)
```

```ts
function show<Item>(item: Item, describe: (item: Item) => string): string {
	return describe(item);
}
console.log(show(3, (value) => `${value}`));
```

```
stage 0 can't lower a value of type Item yet        (Node prints 3)
```

A generic function that only reads an array's length (`function count<Item>(list: readonly Item[]): number`) lowers.

**Around it:** the port checks the Go's in-range indexes with `list[index] ?? panic(...)` through one helper per element type, `nodeAt` and `modeAt` (parsers.ts), where one generic `element<Item>` would do.

## 2. `return undefined` from a function returning `string | undefined` is bad C (closed by stream B2's af5e41e, in integration 10)

Closed: stage 0 now returns undefined as a string's, and the program prints natively what it prints on Node. The port's workaround (main.ts, gap 2) still stands; stream P2 undoes it.

This one isn't refused: stage 0 lowers it, and the C it writes returns an `adamic_object *` from a function declared to return `adamic_string *`. clang refuses it under `-Werror`, so it never runs, but it is the one thing stage 0 promises never to do: reach clang with C that's wrong.

```ts
function nothing(): string | undefined {
	return undefined;
}
console.log(nothing() ?? 'none');
```

```
adamic: native: clang failed: exit status 1
main.c:15:9: error: incompatible pointer types returning 'adamic_object *' (aka 'struct adamic_object *') from a function with result type 'adamic_string *' (aka 'struct adamic_string *') [-Werror,-Wincompatible-pointer-types]
        (Node prints none)
```

The emitted function is `adamic_object * adamic_temporary_1 = adamic_retain(NULL); return adamic_temporary_1;`. The `undefined` literal takes the representation of `undefined` on its own, not the function's return type. Returning a local declared `string | undefined` that holds undefined compiles and runs right, and so does `let found: string | undefined = 'a'; found = undefined;`.

**Around it:** the driver's `escapeOf` (main.ts) returns `''` for a character written as itself, where it would return undefined.

## 3. `?.` on an array or a string (closed by stream P's `?.length` (293976f, a1f90d0), in integration 11)

Closed: `?.length` on an array or a string that may be missing now lowers, and the gap program prints natively what it prints on Node. The port's workaround (gap 3) still stands; stream P2 undoes it.

docs/0.1.md has `?.`. On an object it lowers; on an array or a string it's refused.

```ts
function size(list: readonly string[] | undefined): number {
	return list?.length ?? 0;
}
console.log(`${size(['a'])} ${size(undefined)}`);
```

```
stage 0 can't lower optional chaining on a value yet        (Node prints 1 0)
```

On a `string | undefined`, `text?.length` reads `stage 0 can't lower optional chaining on a string yet`.

**Around it:** the driver's count mode (main.ts) narrows `parsed.value.nodes` into a local and tests it for undefined.

## 4. A tree with mutable child arrays is refused as cycle-capable (closed by stream B3's fresh-write relaxation (8f30be5, 952ccbe), in integration 11)

Closed: the cycle finder lets a cycle-capable slot stand when every write into it is proven not to reach its holder (docs/memory.md, "Relaxing the finder for fresh writes"), and this tree's are. The gap program prints natively what it prints on Node and leaks nothing. The port's workaround (gap 4) still stands; stream P2 undoes it.

Not a NotYet: a rule of 0.1, the cycle finder's (`internal/lower/cycles.go`, docs/memory.md "Cycles"), which landed on main at 4ddd17f and refused this port as first written. A node whose children are a mutable array of nodes can be given itself (`root.nodes.push(root)`), a cycle reference counting can't free:

```ts
class TreeNode {
	readonly name: string;
	readonly nodes: TreeNode[] = [];
	constructor(name: string) {
		this.name = name;
	}
}
const root = new TreeNode('root');
root.nodes.push(new TreeNode('a'));
root.nodes.push(new TreeNode('b'));
console.log(`${root.nodes.length}`);
```

```
Adamic 0.1 refuses TreeNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free; declare the elements weak, Weak<TreeNode>[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly TreeNode[] (adamic/cycle-capable)        (Node prints 2)
```

## 5. ... and so is a tree built from a local array of children (closed by stream B3's fresh-write relaxation (8f30be5, 952ccbe), in integration 11)

Closed, as gap 4: every push of a child into the local array is proven fresh, and the gap program prints natively what it prints on Node and leaks nothing. The port's workaround (gap 5) still stands; stream P2 undoes it.

The same rule refuses the way nearly every parser builds a tree bottom-up: collect the children in a local array, then make the node with them as `readonly nodes`, never writing the array again. The finder reads types, not what happens to a value, so the local `TreeNode[]` is cycle-capable whatever follows:

```ts
class TreeNode {
	readonly name: string;
	readonly nodes: readonly TreeNode[];
	constructor(name: string, nodes: readonly TreeNode[]) {
		this.name = name;
		this.nodes = nodes;
	}
}
function build(): TreeNode {
	const children: TreeNode[] = [];
	children.push(new TreeNode('a', []));
	children.push(new TreeNode('b', []));
	return new TreeNode('root', children);
}
console.log(`${build().nodes.length}`);
```

```
Adamic 0.1 refuses TreeNode[], an array whose elements can reach back to an array like it: ... (adamic/cycle-capable)        (Node prints 2)
```

A rule that let a local mutable array be handed to a `readonly` slot as its last use (moved, not shared) would accept this program and still refuse `children.push(root)` after it. That's for @system_adamic and stream B2 to weigh; the port doesn't wait on it.

**Around both:** the parser built each list of nodes in a local `MediaNode[]` with `push` and handed that array to the container as its `nodes`, which is gap 5's shape. Now the lists are `readonly MediaNode[]` from the start, and `appended` (parsers.ts) grows one by copying, `[...list, node]`: nothing writes a list once it exists. Each list holds a handful of nodes; on the generated params the copying costs about a tenth (2.15 s before, 2.3 to 2.5 s after). The finder also refused the Go's `mediaQueryElement`, rightly: with a `nodes` field it had every field a MediaNode has, so a MediaNode could be seen as one and given itself through that mutable field. Its nodes are now a local, `elementNodes`, reset with it. The values slice, whose lists can be long, keeps each container's children in a table instead (its GAPS.md, "The cycle rule's cost"). Every answer is byte for byte what it was.

## The first slice's gaps, met again

Two of gitignore's open gaps (stream P is closing them) shaped this port too, and are marked where they did:

- **gitignore gap 3, a call to a function declared later**: parsers.ts is ordered callee first (`parseMediaFeature`, `parseMediaQuery`, `matchUrlStart`, `parseMediaList`), which happens to be the Go's order already.
- **gitignore gap 8, a tuple as a value**: the Go's `matchUrlStart` returns `(int, string)`, and the port returns a `UrlStart { length, before }`.

## A bug in Go cohere, found by the port

`parseMediaFeature` in cohere's parsers.go walks the feature's bytes and appends each to the feature's name with `mediaFeature += string(character)`, where `character` is a `byte`. In Go, `string(byte)` makes a rune of the byte, so every byte of a non-ASCII character before a feature's colon becomes a Latin-1 character of its own: the no-break space `C2 A0` comes out as `"Â "`, and its sourceIndex counts the extra bytes. The library appends `string[i]`, one UTF-16 unit, and keeps the character.

The generated params found it, in case 461 of the test's seed, `<=世url ( )*/{﻿,)`. On that feature, `( )` with a no-break space between the parentheses, the port says what postcss-media-query-parser 0.2.3 itself says when run on Node:

```
media-feature "" @9 before=" " after=" "
```

Go cohere says:

```
media-feature "Â" @8 before="" after=" "
```

cohere's own oracle test didn't see it, because none of its non-ASCII fixtures puts the character before a colon inside parentheses (`(max-width:\xc2\xa0100px)` has it after). The fix is one line: `mediaFeature += stringNormalized[i : i+1]`.

**Fixed upstream:** cohere commit `61ed9f4a` preserves the feature's bytes. The bump to `cbe755d3` removes the parser-source overlay and compares the port with Go cohere unchanged. The port already kept the characters, so its behavior does not change. The earlier test ran unfixed and fixed Go on the same cases and found 8 differing answers among 4,103 cases. With `ADAMIC_MEDIA_QUERY_LIBRARY` set to a directory where `npm install postcss-media-query-parser@0.2.3` ran, the test also compares the port with the library itself.

## What lowered as written

A class whose fields are written after it's made (the post-pass retypes nodes in place, and the colon gets its `before` once the feature's `after` is known), a recursive class (`MediaNode.nodes`), `readonly MediaNode[] | undefined` as a field, a generic `Result<Value>` union with `return parsed` passing an error through, `string | undefined` and `number | undefined` parameters with `??` defaults, a closure returning an object literal (the Go's `resetNode`), an array of interface objects pushed and popped as a stack, `break` out of a `for` from inside an `if`, `continue` in a long `if` chain, and the string library the parser leans on: `slice`, `startsWith`, `endsWith`, `trim` and `charCodeAt` over non-ASCII whitespace, and `text[index]` compared with one-character literals.

`trim` is worth a word: the Go writes its own, since Go has no `String.prototype.trim`; the port calls stage 0's, so the test holds stage 0's trim to Node's on every JavaScript whitespace character (U+FEFF, U+3000, U+2028 and the rest are among the generated pieces), and the U+FEFF mutant shows the cases reach it.

## Performance, observed (not refusals)

The port agrees with Go cohere and the library on every case; these are the numbers to beat. Measured on this Linux container, x86-64, unsanitized `-O2` (`adamic build`), with the driver's `count` mode, which parses every case and prints one line, against Go cohere's Parse over the same cases in a loop (the time of the loop alone, so Go's is the most favourable).

| Cases | Native | Node | Go cohere |
|---|---|---|---|
| 200,103 generated (non-ASCII on purpose) | 2.15 s | 0.69 s | 0.54 s |
| the 63,742 of them that are ASCII | 0.35 s | 0.18 s | 0.16 s |

Under callgrind, on 20,000 ASCII cases, the native run's instructions are allocation churn: `adamic_release` 17%, `adamic_string_concat` 13%, malloc and free 23% between them, `adamic_string_slice` 8%. Three shapes cause it, all in natural code the port shares with the Go:

- **`text[index]` makes a new string every read.** The parser walks every params string with `text[i]` and compares the character with `'('` or `','`. Each read slices a one-unit string onto the heap and frees it after the comparison. V8 hands back a cached single-character string. A table of immortal one-unit strings (ASCII, or all 256 of Latin-1) would make the read a load and the comparison a pointer test.
- **`value += character` copies.** The parser builds every element by appending one character at a time, as the Go does. Each `+=` allocates a new string and frees the old one. docs/memory.md's reuse in place (a string whose count is one, appended to in place) is the fix; V8 uses ropes.
- **A short non-ASCII string is walked from its start on every index.** string_index.c gives an index only to strings of 64 bytes or more, so below that `text[i]` and `charCodeAt(i)` walk the UTF-8 from the first byte each time. On the generated cases (mostly short and non-ASCII) `adamic_string_at` averages about 870 instructions a call, and is 37% of the run with `adamic_string_locate` under it. A cursor that a short string keeps without an index (the next index in a loop is one step from the last) would make a loop over one linear without the index's memory.

The driver's printing, which is not the parser, cost more than the parse at first: `quote` read every string unit by unit and appended one at a time (the second shape). It now copies the runs between escapes whole. Even so, writing the trees takes the full run to 4.7 s natively and 2.4 s on Node, and 10% of the native run is converting integers to strings through `snprintf` (`adamic_number_format`), which an integer fast path would remove.

## What the cases reach, and a review that found where they didn't

Reviewer R2 found the generator's pieces held 15 of JavaScript's 25 whitespace characters (of U+2000 to U+200A only U+2003 and U+200A, and U+2028 but not U+2029), and neither DEL nor U+001F, the edges of what the output escapes. So four mutants survived, and with the pieces as they were, each still survives: U+2029 dropped from `isWhitespace`, its U+2000 to U+200A range starting at U+2001, and the driver's quote writing DEL or U+001F as itself. The cohere side now names all 25 (`adamicWhitespace`), the near misses (U+0085, which Go's unicode.IsSpace takes, U+200B, and U+180E, whitespace before Unicode 6.3) and DEL and U+001F, puts every one among the pieces, and writes one fixed case for each, `_screen_and_(_min-width_:_1px_)_,_print_` with the character for every `_`, so that none rests on the generator's draw. All four mutants are in the test and caught, natively and on Node, by those fixed cases. The values slice's driver quotes the same way and had the same blind spot; it has the same two quoting mutants, caught the same way.

Reviewer R's round six (review/round6/REPORT.md, b1ef201) found one more here: a closing quote ending a string whatever quote opened it survived all 4,133 cases, since no case put a string holding the other quote inside parentheses before a colon. It was run against the old cases and survived; R's `({'a"b'}:x)` is now a fixed case, the pieces spell the keywords the parser compares in three cases each (`not`, `NOT`, `Not`, and so on) and carry strings holding the other quote in the shape the feature parser sees (`({'a"b'}:`), and the mutant is in the test and caught. With R's fixed case taken out, the generated cases alone catch it too.

## Not covered

- Real CSS. cohere's oracle test walks `.css` corpora with postcss to find real `@media` params; this container has none, and the port's cases are cohere's fixtures and generated params. A corpus would be a third source of cases, held the same way.
- Prettier's glue (addMissingType, addTypePrefix, the fallback on a throw), which is language-css's, not this package's.
- The order of a node's keys. The Go keeps the library's assignment order, and the port's class has one field order. cohere's oracle compares keys sorted, and so does this test (the Go side checks each node has exactly the port's keys).
- Lone surrogates. The cases are UTF-8 text, so no params holds one; the library walks UTF-16 units and the port does too, so a params with one would be read alike, but no case shows it.
