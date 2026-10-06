# What stage 0 couldn't lower in cohere's suppression index, and where Go cohere and ESLint disagree

The port beside this file is cohere's `internal/lint/suppression` (suppression.go, scan.go, parse.go and directive_subject.go) with the grammar it reads, `internal/lint/ecmascript/directives` (directives.go and subject.go), written as 0.1 Adamic. It is stage 1's fourth slice, after `gitignore`, `mediaquery` and `values`, whose GAPS.md files this one follows and refers to. The gaps were found at main 4ddd17f.

Every place stage 0 refused the port, or took it and made C that clang refused, is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number (`grep -n "gap N" *.ts`), and a workaround for another slice's gap names that slice's GAPS.md. The programs are in `gaps/`, and `gaps_test.go` holds them to this file.

## Why this slice

The first three slices were parsers: a string in, a tree out, nothing kept. Ahra asked for one that exercises what they didn't, either classes with Map-heavy state or a lint rule's pure core held by cohere's own rule fixtures.

- **lint/suppression** (chosen): the index every cohere lint run builds of a file's `eslint-disable` comments, and what each one withheld. About 950 lines of Go with `strings` alone, and it brings what the parsers didn't: a class whose state changes as it's asked (each `Suppresses` that a directive answers counts against it, so the order of the questions shows in the answers), a module-level `Map` written by registration and read by every query (`RegisterSubject`), a closure returned as a value and kept in a field (the line index), mutual recursion (the scanner and the template skipper call each other), a lexer that has to tell a regular expression from a division, and Go's own whitespace, which isn't JavaScript's. Its tests carry ESLint-measured expectations, and the test here reads its sources straight out of them.
- **A single lint rule's pure core** (runner-up): every rule's fixtures are whole sources run through typescript-go's AST, which a port can't take as input, so a pure core would be held by fixtures of my own extracting, not the rule's. Not the oracle Ahra asked for.
- **lint/report** (looked at): 125 lines of formatting, too little state.

## 1. Two functions that call each other (closed by stream P's signatures before bodies (24142eb), in integration 11)

Closed: stage 0 now writes every function's signature before it lowers any body, so a call knows what a function declared below it returns, and the gap program prints natively what it prints on Node. The port's workaround (gap 1) still stands; stream P2 undoes it.

gitignore's gap 3 is a call to a function declared later, taken for a void call. Its GAPS.md says that taking signatures from the checker up front "would also make mutual recursion possible, which no order can give today", and this is that program:

```ts
function isEven(value: number): boolean {
	return value === 0 ? true : isOdd(value - 1);
}
function isOdd(value: number): boolean {
	return value === 0 ? false : isEven(value - 1);
}
console.log(`${isEven(4)} ${isOdd(4)}`);
```

```
stage 0 can't lower a void call used as a value yet        (Node prints true false)
```

**Around it:** scan.go's `scanCode` calls `skipTemplate` for a template literal, and `skipTemplate` calls `scanCode` for each `${...}`, because an interpolation is code. The port writes `skipTemplate`'s loop inside `scanCode`'s template case (scan.ts), where the recursion is a function calling itself, which lowers.

## 2. `return undefined` from a function returning an array or undefined (closed by stream B2's af5e41e, in integration 10)

Closed: the program prints natively what it prints on Node. The port's workarounds (directives.ts and main.ts, gap 2) still stand; stream P2 undoes them.

mediaquery's gap 2 is `return undefined` from a function returning `string | undefined`, which lowers to C that clang refuses. It is not only strings: an array does the same.

```ts
function names(found: boolean): readonly string[] | undefined {
	if (found) {
		return ['a'];
	}
	return undefined;
}
const got = names(false);
console.log(`${got === undefined}`);
```

```
adamic: native: clang failed: exit status 1
main.c: error: incompatible pointer types returning 'adamic_object *' (aka 'struct adamic_object *') from a function with result type 'adamic_array *' (aka 'struct adamic_array *') [-Werror,-Wincompatible-pointer-types]
        (Node prints true)
```

An object or undefined (`Disable | undefined`, here) lowers and runs right.

**Around it:** the Go's `ParseEnable` returns `(rules, found)`, and `splitDirective` `(rest, found)`; the natural ports return the list or undefined and the string or undefined, which this gap and mediaquery's refuse. They return an `Enable { rules }` and a `DirectiveRest { rest }` instead (directives.ts), and `Recognize` returns `''` for a comment that is no directive.

## The other slices' gaps, met again

- **gitignore gap 8, a tuple as a value**: every Go function here with two or three results (`ParseDisable`, `ParseEnable`, `splitScope`, `splitReason`, `Recognize`, `splitDirective`) returns an object, or an object or undefined.
- **gitignore gap 9, `return panic(...)`**: the driver's `unescapeOf` ends with `panic(...)` as a statement.
- **mediaquery gap 2**: above, and the driver's `escapeOf`.

The cycle finder (mediaquery's GAPS.md, "The cycle rule's cost") had nothing to say here: a Directive holds strings and numbers, the Index holds lists of Directives and a closure over a list of numbers, and nothing reaches back.

## Go cohere and ESLint disagree on two whitespace characters

Not a stage 0 gap, and not the port's: the port answers as Go cohere does, which is what it's held to. But the port is where it showed, and it matters to cohere's own doctrine ("the parity doctrine is never worse than ESLint").

directives.go trims a comment's body, its rule names and its reason with Go's `strings.TrimSpace`, which trims what Go's `unicode.IsSpace` calls space. ESLint is JavaScript and trims with JavaScript's whitespace. The two sets differ by two characters: Go takes U+0085 (NEL) and leaves U+FEFF; JavaScript takes U+FEFF and leaves U+0085. The natural TypeScript port, calling `trim()`, answers as ESLint does; this port writes Go's trim as `goTrimSpace` (directives.ts), on purpose, and the test's U+0085 mutant shows the cases reach the difference.

Measured on ESLint 10.12.0 (`npm install eslint`; `testdata/eslint_whitespace.cjs`, the Linter API with `no-debugger` on), against Go cohere's `Build` and `Suppresses` on the same eight sources, each a comment and then `debugger;` on the next line:

| Comment | ESLint 10.12.0 | Go cohere |
|---|---|---|
| `/* eslint-disable no-debugger */` | suppressed | suppressed |
| `/*` U+0085 `eslint-disable no-debugger */` | reported | suppressed |
| `/*` U+FEFF `eslint-disable no-debugger */` | suppressed | reported |
| `/* eslint-disable no-debugger` U+0085 `*/` | reported | suppressed |
| `/* eslint-disable no-debugger` U+FEFF `*/` | suppressed | reported |
| `//` U+0085 `eslint-disable-next-line no-debugger` | reported | suppressed |
| `//` U+FEFF `eslint-disable-next-line no-debugger` | suppressed | reported |
| `/*` U+00A0 `eslint-disable no-debugger */` | suppressed | suppressed |

Six of eight differ, every one where the two whitespace sets do. cohere's tests measured ESLint 10.8.1; this is 10.12.0, and I haven't checked the earlier version. The fix in cohere is one function, a trim over JavaScript's whitespace in place of `strings.TrimSpace` in directives.go; when cohere makes it, the port's `goTrimSpace` becomes `trim()` and the U+0085 mutant changes sides. Sent to Ahra with this report, not fixed from here (this stream's access is the adamic repository).

## Not a stage 0 gap: Node's stdout into a Go pipe

With 20,000 generated sources (2.26 million lines of output), the test failed with Node exiting 70 and nothing on stderr. The port wasn't at fault: Node 24.21.0, writing many short lines to a pipe that Go's `os/exec` reads, fails a `writev` with `EINVAL`, and the oracle's runtime (`oracle/adamic.mjs`) turns any error on stdout into a silent exit 70. With no Adamic code at all:

```go
command := exec.Command("node", "-e", `
process.stdout.on('error', (error) => { require('fs').writeSync(3, String(error.stack)); process.exit(70); });
for (let index = 0; index < 2300000; index++) { console.log('x'.repeat(40)); }`)
var stdout bytes.Buffer
command.Stdout = &stdout // the report goes to an ExtraFiles file at fd 3
```

```
Error: write EINVAL
    at afterWriteDispatched (node:internal/stream_base_commons:159:15)
    at writevGeneric (node:internal/stream_base_commons:142:3)
```

Measured, each run once: 2,300,000 lines of 40 bytes fail; the same 92 MB as 23,000 lines of 4,000 bytes or 230,000 of 400 succeed, and so do 100,000 and 20,000 lines of 40. A shell pipe (`| cat`) never failed. My reading, an inference I haven't confirmed by counting: while the pipe is full, Node queues each `console.log` as its own buffer, and when it drains them in one `writev`, more than IOV_MAX (1,024 here) buffers is EINVAL. The oracle and every stage 1 test run Node this way, so a program that prints enough short lines fast enough fails on Node for no fault of its own. This test's default (2,500 generated sources, 282,774 lines) has passed every run. The fix is in the oracle's runtime, which isn't mine to change: making stdout blocking, or writing large outputs in fewer calls.

## What lowered as written

A class with mutable fields counted as it's asked (`Directive.applied`) and methods that read and write them, a second class holding a `readonly Directive[]`, a `readonly RuleReference[]` and a closure, all `#private`; a module-level `Map<string, boolean>` written by one function and read by another; a function that returns a closure over a `number[]` it built, kept in a field and called through it; `filter` with a closure over the class's own elements; a recursive lexer that reads a template's interpolations as code; string `switch`es over the scope words; `split`, `join`, `startsWith`, `endsWith`, `slice`, `indexOf` and `includes` throughout; and `Number.parseInt` on the driver's offsets.

## Performance, observed (not refusals)

On 20,125 sources (123 read out of cohere's tests, 2 fixed and 20,000 generated, seed 20261005, from the generator as it was before the class edges below), with 2,091,150 records asked and 2,262,905 lines answered, unsanitized `-O2`, each run twice:

| | Native | Node | Go cohere |
|---|---|---|---|
| the whole run | 5.56 to 5.63 s | 6.00 to 6.52 s | 1.45 to 1.52 s |

Native is a little faster than Node here, the first slice where it is. Go cohere's time is the same records answered and formatted into memory, so it does what the port does but read its cases and write its 70 MB: the most favourable to Go. Native was 6.1 s before the driver's `unescape` copied the runs between escapes whole, rather than appending a source a character at a time; under callgrind that loop had been a quarter of the run.

Under callgrind now, on a 3 MB slice of the cases, the index's own work (`scanCode`, the scanner) is a fifth, and the rest is the driver's reading and writing: `adamic_string_slice` 25% with what it calls, `adamic_string_at` 20% (`text[index]` makes a one-unit string every read, mediaquery's first shape), `adamic_string_concat` 19%, `adamic_string_split` 15%, and turning offsets into text 8% (`adamic_string_from_number`). mediaquery's GAPS.md has the shapes and their fixes.

## What the cases reach

Reviewer R's round six found the value and media query tests never reached the edges of their ports' character classes, so plausible bugs at those edges passed. This slice had the same classes and was checked for the same gap before anyone reviewed it. Six mutants, one at an edge of each class or keyword the scanner and grammar test: a `)` opening a pattern, a vertical tab skipped as space, a no-break space ending a directive word, a directive word read in any case, a string running past its line's end, and a regular expression's `[...]` ignored. Run against the cases as they were, two survived: the vertical tab and the no-break space, since neither ever stood where it mattered. The generator now writes every character after which a `/` opens a pattern, and the characters after which it divides, each before a pattern, with the whitespace the scanner skips and doesn't between (`adamicEdgeCode`), and whitespace other than a space or a tab after a directive word, and the directive words in mixed case; two fixed sources hold the two survivors. The pattern has to be `/ //x/`, not `/x/`: read as a pattern or as two divisions, `/x/` leaves the same comments behind, and the first try at this, with `/x/`, left the vertical-tab mutant alive. All twelve mutants are caught, natively and on Node.

## Not covered

- Real source. cohere counted 387 directives in the corpus it gates, but that corpus isn't in this container, and the cases are the 123 sources read out of cohere's tests, four fixed ones, and 2,500 generated.
- The nil-Index methods. The Go's methods answer for a nil `*Index`; the port's Index is always one `build` made.
- The registry's concurrency: Go writes `subjectRules` only from `init`, before any file is walked, and the port's driver registers before it builds an index, the same way; neither side is concurrent.
- The `fmt.Sprintf`-built sources in cohere's spelling tests, which go/ast can't read out as literals; the generator writes every spelling and scope instead.
