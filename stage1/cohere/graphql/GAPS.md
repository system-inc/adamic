# What cohere's GraphQL parser needed that Adamic doesn't have yet

The port beside this file is cohere's `internal/format/graphql` lexer and parser (lexer.go, parser.go, token.go, block_string.go, character_classes.go): graphql-js 17.0.2's `parse`, as Prettier calls it, with Prettier's `parseComments`. It is stage 1's sixth slice, and the first that throws and catches. graphql-js throws a `GraphQLError` wherever a text fails to lex or parse, from as deep in its recursive descent as the failure is; cohere's Go panics with a `*SyntaxError` there and recovers it in `Parse`. The port throws `new Error(message)` from the same places and catches it once, in `parse`, so a failure unwinds every frame between, each holding the tokens, lists and nodes its level has made.

`graphql_test.go` holds it to Go cohere byte for byte, natively under ASan and UBSan with the leak check, on Node, and through the JavaScript backend; with `ADAMIC_GRAPHQL_LIBRARY` set, it holds Go cohere to graphql-js 17.0.2 itself on the same texts. `gaps_test.go` holds the two gaps below to this file.

## Why this slice

Ahra asked for a sixth slice that throws and catches, now that exceptions are on main.

- **format/graphql** (chosen): the throw is the library's own design, not a shape the port imposes. graphql-js throws from 35 places in the lexer and the parser, at every depth of a recursion that nests lists, objects, selection sets and list types; the Go mirrors it with panic and recover, and says so in its comments. A port that throws reads as both of them do. It is self-contained (2,435 lines of Go, no regular expressions), cohere's tests hold it to graphql-js with 94 inline fixtures, and its trees are what Prettier's GraphQL printer reads.
- **css/postcss** (runner-up): its parser panics and recovers too, but it is 2,328 lines with the SCSS dialect woven through, and it leans on the tokenizer the values slice has only part of.
- **css/selector** (runner-up): it throws as well, but it matches with regular expressions, which 0.1 doesn't have.

## 1. A function value that can throw (closed by stream C2's 5af7bfb, in integration 10)

Closed: function values may throw, and the program prints natively what it prints on Node. The port's workaround (grep 'gap 1') still stands; stream P2 undoes it.

graphql-js's parser hands each of its list helpers (`any`, `many`, `optionalMany`, `delimitedMany`) the parse function for the list's items, and every parse function can throw. Stage 0 refuses any function value that can throw, though the runtime never calls this one; the program calls it itself:

```ts
function total(items: readonly string[], measure: (item: string) => number): number {
	let sum = 0;
	for (const item of items) {
		sum += measure(item);
	}
	return sum;
}
try {
	console.log(`${total(['a', 'bb'], (item) => {
		if (item.length > 1) {
			throw new Error(`too long: ${item}`);
		}
		return item.length;
	})}`);
} catch (error) {
	console.log(error instanceof Error ? error.message : '?');
}
```

```
stage 0 can't lower a function value that can throw (the runtime calls them from its own loops, which a throw would have to leave) yet        (Node prints too long: bb)
```

The reason the message gives is the runtime's own loops (`map`, `sort` and the rest), which a throw would have to leave; a closure only the program calls has no runtime frame to leave. A rule that refused a throwing function value only where it reaches a runtime loop would take this program.

**Around it:** each list is its helper's loop, written where graphql-js calls the helper, and marked with the helper's name (`// optionalMany('(', parseVariableDefinition, ')')`). A dispatching method taking the item's kind in the function's place doesn't lower either, since it would call methods declared after it (gap 3 below, met again).

## 2. Throwing an Error made in another function (closed by codex/error-classes)

Closed: stage 0 now throws stored, returned and passed errors with nominal identity intact. The minimal gap fixture passes Node, native and the leak check. The port still constructs errors at its throw sites; changing that source is a separate port cleanup.

graphql-js's parser throws what `this.unexpected()` and `syntaxError()` return.

```ts
function unexpected(at: number): Error {
	return new Error(`Unexpected token at ${at}.`);
}
try {
	throw unexpected(3);
} catch (error) {
	console.log(error instanceof Error ? error.message : '?');
}
```

```
stage 0 can't lower throwing an Error that isn't made where it's thrown or caught by the catch around it yet        (Node prints Unexpected token at 3.)
```

**Around it:** `syntaxError` and `unexpected` return the message, and each of the 35 places throws `new Error(...)` of it. Since 0.1 has no class inheritance, the error can't carry graphql-js's `locations` either; Prettier reads only the first location into its message, so the message is the whole of what the port needs.

## Cooking a string without `String.fromCodePoint`

The gitignore slice's gap 1 again: `String.fromCodePoint` and `String.fromCharCode` are in docs/0.1.md's library and don't lower on any branch, main at 4ff4657 included. A GraphQL string's `\u` escapes need them: graphql-js cooks each escape into the character it names, and a value can name any of the 1,112,064 scalar values. No table of them is small enough to embed.

**Around it:** every string the lexer gives, a token's value or an error's message, is in the output form the test compares: printable ASCII as itself but a backslash as two, and any other code point as `\u{HEX}` (lexer.ts, `escaped` and `written`). A cooked escape is written straight into that form from its code point, and a raw stretch of the source is put into it by `escaped`; the Go side writes Go cohere's cooked strings in the same form, so the comparison sees every character of every value. Names, numbers and keywords are ASCII letters, digits, `_`, `.`, `+` and `-`, which the form writes as themselves, so every comparison the parser makes on a token's value is unchanged. When `String.fromCodePoint` lowers, the values go back to being the text itself and only the driver writes the form.

A lone surrogate, which only an error message's slice through a surrogate pair makes (`"\u{1😀"` quotes the escape up to the cut), is written as U+FFFD: what Node writes for one when the message leaves as UTF-8, and what the Go's `sliceUnits` makes of it. Natively it is a lone surrogate until then, as on Node; the probe that checked this agreed byte for byte.

## The other slices' gaps, met again

- **gitignore's gap 2, the bitwise operators.** `readEscapedUnicodeVariableWidth` builds its code point as `(point << 4) | readHexDigit(code)`, a signed 32-bit result, and stops when it goes negative: a digit that isn't one, or a shift into the sign bit. The port writes it as arithmetic with both of those tests spelled out; the sign bit is the edge one mutant below sits on. `read16BitHexCode` ORs four shifted digits, where a -1 makes the whole negative; the port tests for a -1, then adds.
- **gitignore's gap 3, a call to a method declared later.** It shaped this port more than any other. A recursive descent parser is mutually recursive: `parseValueLiteral`, `parseList`, `parseObject` and `parseObjectField` call each other, and so do `parseSelectionSet`, `parseSelection`, `parseField` and `parseFragment`. With gap 3 no order of methods lowers a cycle, so each cycle is one self-recursive method (`parseValueLiteral` and `parseSelectionSet`), with the others written inside it where they're called and marked with their names. The other methods are callee first.
- **suppression's gap 2, `return undefined` from a function returning an array or undefined.** graphql-js's optional lists (`parseDirectives`, `optionalMany` and the rest) return a list or undefined. The port's return the field's `Value`, a `List` or `Undefined`, and the extensions test `.kind` where graphql-js tests for undefined.
- **mediaquery's gaps 4 and 5, the cycle rule.** A node holding the list of nodes it was built from is cycle-capable to 0.1's rule, and a document's lists can be long, which rules out growing each by copying, as the mediaquery port does. Each node's children are indexes into the document's table of nodes, as the values port keeps them. And graphql-js's tokens are a doubly linked list, each holding the token before it: a cycle by design, which the rule rightly refuses. The parser never walks the list backwards, so the lexer keeps the tokens in an array, in the list's order, and `parseComments` walks the array.

## What lowered as written

Every throw and the catch, on the first build. `throw new Error(...)` from 35 places, among them functions and methods with no other way out (`readString` ends in a throw), a ternary choosing which message to throw, and throws inside `switch` cases inside loops; the throw leaving through every method of the recursion, from as deep as 500 frames (the deep cases below), each frame holding arrays of node indexes, tokens and strings, with nothing leaked on any of the 3,426 texts, 1,374 of which throw; `catch (error)` narrowed with `instanceof Error` and read for its `message`; and the try's check that nothing it reaches is a library call that would throw on Node but panic natively, which the port passes with `toString(16)` (a constant radix) and `padStart` inside it.

Around the exceptions: a `switch` on character codes with cases falling through to one body and `continue` inside the `switch` reaching the loop around it; `charCodeAt` past the end giving NaN, which every comparison then rejects as graphql-js relies on; `codePointAt` past the end giving undefined; `do ... while` loops; nested ternaries; a string-literal union of 23 token kinds compared with `===` and switched on; a discriminated union of five object shapes as a field's value; `Number.MAX_SAFE_INTEGER`; and `slice` clamping an index past the end.

## No bug in Go cohere this time

Go cohere, the port and graphql-js 17.0.2 give the same answer, tree, comments or message, on every one of 30,426 texts (seed 20261005 with 30,000 generated): 19,426 trees with 1,088,017 nodes and 38,868 comments, and 11,000 refusals.

## What the cases reach

Every string constant in cohere's graphql tests, read out of their Go with go/ast: the 94 inline fixtures its graphql-js oracle test runs, the shapes, values, comments and errors its parser tests check, and the snippets its format tests print, with their names and expectations, which are texts too (414 in all). Then twelve deep texts: lists 500 deep, objects 300 deep, selection sets 300 deep and list types 400 deep, each whole, with a closer missing, and broken at the bottom, so a throw leaves through every level. Then 3,000 generated documents, a third of them wild. A tame document is built from valid pieces at the edges of each class: ignored characters (space, tab, comma, the byte order mark, each line ending), names at each end of the letter and digit ranges, every keyword, every one of the 21 directive locations, numbers in every shape, characters at the edges of UTF-8's and UTF-16's lengths and of the scalar values, every escape and the hex digits' cases and ends, surrogate pairs in both cases and at their ends, block strings with every indent and line ending, and comments. A wild one also takes the near misses just outside each class (a vertical tab, a form feed, U+00A0, U+0085, U+2028 and U+200B as separators; `@`, `[`, a backquote, `{`, `/` and `:` after a name; keywords in the wrong case; locations just off; every way to get a number wrong; escapes with `G`, `/`, `:`, `@` or a backquote among the hex digits, surrogates alone and out of order, points past U+10FFFF, the widest escape and one past it, points that reach the sign bit), and is often cut short or given a stray character. Of the 3,426 texts, 2,052 parse.

Two of graphql-js's messages are out of reach, and the test doesn't ask for them: "Invalid character within String" and "Invalid character", which need a lone surrogate in the source, and no UTF-8 input holds one.

Sixteen mutants, each caught natively and on Node: the byte order mark not ignored; a no-break space ignored; `[` taken for a letter; `G` taken for a hex digit; the sign bit not reached; U+DFFF not a trailing surrogate; a lone surrogate written as itself; a CRLF counted as two lines in an error's location; a block string's first line counted in its indent; an escaped triple quote keeping its backslash; a leading zero's error placed at the number's start; an invalid escape read as nothing rather than thrown; `unexpected` naming the current token rather than the one it was given; no directives written as an empty list; a fragment spread's absent arguments written as undefined; and `FRAGMENT_VARIABLE_DEFINITION` left out of the locations.

## Not a gap here: Node's writev at millions of lines

The suppression slice's GAPS.md has Node 24.21.0 failing a `writev` with EINVAL when it writes millions of short lines into a pipe Go reads, at 2.26 million lines there. This test can't reach it: its 3,426 texts answer in 11,068 lines, and the 30,426-text run timed below in 99,720.

## Performance, observed (not refusals)

On the 30,426 texts above, every side's output identical, unsanitized, each run three times, back to back:

| | Native | Node | Go cohere |
|---|---|---|---|
| the whole run | 3.07 to 3.17 s | 1.37 to 1.44 s | 1.44 to 1.51 s |

Go cohere's time is `Parse` on every text with every answer formatted into memory, not written out. Native is about twice Node and Go here, unlike the format walk and the suppression index, where it beats Node. Under callgrind, on 3,259 of the texts:

- **Numbers as text, 15% of the run.** The driver writes two offsets per node, and stage 0's runtime formats every number through V8's shortest-digits algorithm in its bignum form (`adamic_number_shortest_digits`, dtoa.c), with neither V8's fast path (Grisu) nor one for integers in front of it: about 2,000 instructions for each of the 214,602 offsets written. A throwaway driver that kept each offset's text in a table ran in 2.58 to 2.63 s. The table isn't kept: the cost is the runtime's, and an integer fast path there would take it for every program.
- **Making and freeing the trees, most of the rest.** malloc, free, `adamic_retain` and `adamic_release` are half the instructions left once the numbers are out of the way. Each node is a `GraphNode`, an array of fields, and a field object and a value object for each field, made one at a time and freed one at a time when the document goes, where Node bumps a pointer and collects the young generation in bulk.
- **Not the throws.** `syntaxError`, which runs once for each throw, is 0.3% of the run, where a third of the texts throw.

The driver's own reading of the cases was quadratic at first: each escape in a long line appended the text so far to itself (105,000 concatenations of about 1,500 instructions each). It now splits the line at its backslashes and joins once, which took the run from 3.2 s to 3.0 s.

## Not covered

- `maxTokens`, `noLocation` and the other options Prettier leaves at graphql-js's defaults; `parseValue`, `parseConstValue`, `parseType` and the schema coordinate entries, which Prettier never calls (the Go leaves them out too).
- Text that isn't UTF-8: the test refuses to write such a case, since Node and 'adamic' decode it with replacement characters and Go keeps its bytes.
- graphql-js's `locations` beyond the first, and the error's other fields, which Prettier doesn't read.

## Composed printer

The formatter is ported in [printer](printer/README.md), using this parser
unchanged. Its [gap report](printer/GAPS.md) records the compiler workaround,
exact Go parity and five upstream whitespace-only differences with Prettier.
