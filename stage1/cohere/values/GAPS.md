# What stage 0 couldn't lower in cohere's CSS value parser

The port beside this file is cohere's `internal/format/css/values` (values.go, nodes.go, tokenize.go and parser.go), itself a port of postcss-values-parser 2.0.1, the parser Prettier's language-css hands every declaration value to, written as 0.1 Adamic. It is stage 1's third slice, after `stage1/cohere/gitignore` and `stage1/cohere/mediaquery`, whose GAPS.md files this one follows and refers to.

Every place stage 0 refused the port is below (gap 5 found at main 4ddd17f, the rest at 2385966), as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number (`grep -n "gap N" *.ts`), and a workaround for another slice's gap names that slice's GAPS.md. Each program typechecks under stage 0's options and runs on Node 24.21.0 with the output shown.

The programs are in `gaps/`, and `gaps_test.go` holds them to this file: an open gap must still be refused with the words recorded here, and a closed one must lower and print natively what it prints on Node, leaking nothing.

## 1. `indexOf` from a position (closed by stream P's indexOf and includes from a position (f34d884), in integration 11)

Closed: `indexOf` and `includes` with a position now lower, and the gap program prints natively what it prints on Node. The port's workaround (gap 1) still stands; stream P2 undoes it.

docs/0.1.md's library has `indexOf`. With one argument it lowers; with the position to search from, it's refused.

```ts
const text = 'a"b"c';
console.log(`${text.indexOf('"', 2)}`);
```

```
stage 0 can't lower indexOf with these arguments yet        (Node prints 3)
```

**Around it:** `indexOfFrom` (tokenize.ts) slices from the position and searches the rest. The tokenizer looks for a closing quote, the end of a comment and the end of a line that way, as upstream does with `css.indexOf(search, from)`. It copies the rest of the value each time, which upstream doesn't.

## 2. Boolean field gap retired

The refusal test and shape-encoded boolean workaround are removed. ValueNode now carries inline, quoted, isHex and isColor as boolean | undefined, and the renderer determines boolean field presence from those values. Shape still selects nonboolean fields and container kind; it no longer supplies boolean presence. The port's own differential test holds every parsed tree to Go cohere, source Node and both backends under sanitizers. The minimized program moved to internal/oracle/testdata/boolean_optional_values.a.

## 3. A `case` that names a constant (closed by compiler's area-next, on main as 41791f72)

Closed: a case that names a constant, a number or a string, now lowers, and the gap program prints natively what it prints on Node. The port's workaround (gap 3) still stands: the tokenizer's switches still write each case as its character code. A follow-up writes them as the constants' names, held to the CSS corpus.

```ts
const newline = 0x0a;
function describe(code: number): string {
	switch (code) {
		case newline:
			return 'newline';
	}
	return 'other';
}
console.log(`${describe(10)} ${describe(32)}`);
```

```
stage 0 can't lower a case that isn't a constant yet        (Node prints newline other)
```

A `const` string is refused the same way. A literal lowers.

**Around it:** the tokenizer's big switch writes each case as its character code with the constant's name beside it, `case 0x0a: // newline`. The constants are still declared and used in the `if`s, as upstream's are.

## 4. A `#private` method (closed by class features)

Closed: private methods now lower, and the gap program agrees with Node under the sanitizers and leak check. The parser now uses `#private` methods and fields. The following records the original refusal.

```ts
class Counter {
	#count = 0;
	#step(): void {
		this.#count++;
	}
	twice(): number {
		this.#step();
		this.#step();
		return this.#count;
	}
}
console.log(`${new Counter().twice()}`);
```

```
stage 0 can't lower a method with a computed name yet        (Node prints 2)
```

A `private step()` lowers.

**Original workaround, removed:** the parser used TypeScript `private` methods with `#private` fields.

## 5. An empty array literal as a default (closed by compiler's area-stack, views slice 1, Oct 8)

Closed: an empty array literal as a `??` default now lowers, and the gap program prints natively what it prints on Node. The port's workaround (gap 5, `noChildren` in parser.ts) still stands; retiring it is cohere's change. What it used to record:

```ts
function size(list: readonly string[] | undefined): number {
	return (list ?? []).length;
}
console.log(`${size(['a'])} ${size(undefined)}`);
```

```
stage 0 can't lower an array of never yet        (Node prints 1 0)
```

`(lists[1] ?? []).length` on a `string[][]` reads the same. The checker types the `[]` as `never[]` before the `??` widens it.

**Around it:** `noChildren` (parser.ts), a `readonly ValueTree[]` constant a leaf's tree starts from, found when the cycle rule (below) moved the children out of the nodes.

## The cycle rule's cost

Not a NotYet: a rule, the one mediaquery's GAPS.md records as gaps 4 and 5, "a tree with mutable child arrays is refused as cycle-capable", with their smallest programs. Upstream's Container is gap 4's program exactly. Here it reshaped the parser. Upstream's Container has `nodes`, and the parser appends to it long after it made the node: a func gets its arguments, the value gets its words. A mutable `nodes` field, or a mutable field holding a `readonly` list, can be made to hold its own node, and is refused (adamic/cycle-capable). Copying on every append, as the media query port does, would be quadratic in a long value.

So a container has an `id`, the parser keeps each container's children in `#children` by it, a `ValueNode[][]` (a ValueNode no longer reaches a list of nodes, so nothing there can close a cycle), and when the parse is done, `tree` builds the `ValueTree` the parse returns, whose lists are `readonly`. That's the arena docs/memory.md describes for stage 1, kept by hand: one parse's nodes live in one table until the parse ends. The timings didn't move (2.65 s before, 2.60 to 2.65 s after).

## The other slices' gaps, met again

- **gitignore gap 3, a call to a function or method declared later**: the parser's methods are ordered callee first, so `parseTokens` and `loop` come last. The Go's order is the library's, which is the other way round.
- **gitignore gap 4, a declared function as a value** (closed since): the Go passes `regexLastIndex` the regex as a function (`wordEnd`, `wordEndNum`, `atEnd`); the port passed its name and switched on it until the gap closed, and passes the function now.
- **gitignore gap 9, `return panic(...)`**: the driver's `field` (main.ts) ends with `panic(...)` as a statement.
- **mediaquery gap 2, `return undefined` from a `string | undefined` function is bad C**: the driver's `escapeOf` (main.ts) returns `''` for no escape, as the media query driver does.

Upstream's token is an array, `[type, value, startLine, ...]`, a tuple (gitignore gap 8). The Go made it a struct and the port follows the Go, so that gap isn't what shaped it.

## No bug in Go cohere this time

The port agrees with Go cohere on every case, and with the library itself: with `ADAMIC_VALUES_LIBRARY` set to a directory where `npm install postcss-values-parser@2.0.1` ran, values_test.go holds the port to the library run on Node, every node written with all of its own fields, whatever they are. It was run that way on this Linux container, and all 8,184 parses agree. cohere's port is careful work: its quirks (the NaN end column of a first bracket, an operator's sourceIndex taken from its end line, the space that compares a token's kind with `','`, which no kind is) are all the library's, and the port keeps each one, with a mutant that removes it.

## What lowered as written

A tokenizer over UTF-16 units with `charCodeAt` past the end giving NaN, which every comparison treats as no character, as upstream's does. `let next = NaN`, with arithmetic on it carrying NaN through to a printed column. A `do...while` loop with a `break` out of the switch around it. A closure pushing to an array it captured, and another building an error from `let`s it captured. A class whose fields are written while its methods run (the parser's position, current node, cache and spaces), holding a `readonly Token[]` of interface objects. A union of string literals as a switch's subject and as a parameter. `split` and `at(-1)`, `includes` on a `number[]`, and string `+=` throughout.

## Performance, observed (not refusals)

Measured on this Linux container, x86-64, unsanitized `-O2`, with the driver's `count` mode, which parses every value in both modes and prints one line, against Go cohere's Parse over the same values in a loop (the loop alone).

| Values (each parsed twice) | Native | Node | Go cohere |
|---|---|---|---|
| 100,092 generated | 2.65 s | 0.83 s | 2.39 to 2.44 s |
| the 62,803 of them that are ASCII | 1.34 s | 0.41 s | 1.37 to 1.69 s |

Native is level with Go cohere here, where on the media query parser it was four times slower: cohere's Go builds every node as an `estree.Node`, a list of properties with `any` values, plus maps for raws and source, and recovers every throw from a panic, and that costs it what Adamic's typed objects don't. Node is three times faster than both.

Under callgrind, on 10,000 ASCII values, the native run is allocation and counting: `adamic_release` 23% of the instructions in itself (35% with what it calls), malloc and free about 23% between them, and `adamic_string_slice` and `adamic_string_concat` 16% and 10% with what they call. Beyond the media query slice's three shapes (`text[index]` allocating, `+=` copying, short non-ASCII strings walked from the start), two more show here:

- **A field of an interface value is read through a call.** `adamic_object_field` (6%) finds a field of a structurally typed object through a one-entry shape cache, a call and a compare for every `token.kind` and `token.value`. A class's fields don't pay it. The parser reads its tokens' fields all the time; a Token class would avoid it, but a Token is what the Go and upstream have, an interface.
- **Freeing a node releases each of its fields one call at a time.** A ValueNode has twelve fields, and every parse frees the whole tree and every token when the case is done, all through `adamic_release` and its freeing list. An arena per parse (docs/memory.md, "Arenas") is the fix this was written for: one request, one file, freed at once.

## What the cases reach, and a review that found where they didn't

Reviewer R's round six (review/round6/REPORT.md, b1ef201) found the cases never reached the edges of the port's character classes or the case of its keywords, so eleven plausible bugs agreed with Go cohere: `alphaNum`'s digits stopping at 8 (which splits `#999`, an everyday color, into `#` and `999`), its lowercase stopping at y, its uppercase at Y, its taking `_`, `isColor`'s digits stopping at 8, a word ending at `=`, the strict calc check, the url argument check and the tokenizer's url argument each ignoring case, `unicodeRange`'s digits stopping at 8, and `--` a word only before more text. Each is now a mutant in the test, written against this port; each was run against the old cases and survived, as R found. R's values are fixed cases now (`#999`, `#z1`, `#Z1`, `#_a`, `a=b`, `CALC(1 -2)`, `URL(a b)`, `Url(//x)`, `u+99`, `--`), and the generator draws a quarter of its pieces from `adamicClassEdgePieces`: `#` before each end of each class and the characters just outside them, colors of every length, unicode ranges at each edge, exponents at both digits, every character that ends a word and its ASCII neighbours, and calc and url in three cases each. All eleven are caught, natively and on Node, and with R's fixed cases taken out, the generated cases alone catch all eleven too. The port itself agreed with Go cohere and the library before and after: R's 30,023 random values found no divergence, and neither did this.

## Not covered

- Real CSS. cohere's oracle test walks `.css` corpora with postcss for every value Prettier would parse; this container has none, and the cases are cohere's 80 fixtures, its 12 shape and refusal values, and 4,000 generated values, each parsed in both modes.
- Prettier's glue (parse-value.js, which turns the tree into Prettier's), which is language-css's.
- The order of a node's keys. Both sides write them sorted, as cohere's oracle compares them.
- Lone surrogates, and text that isn't UTF-8 (the Go reads a bad byte as U+FFFD), since the cases are UTF-8 text.
