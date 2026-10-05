# Stream R, round six

This round has two parts:

- **P2's CSS value parser port**, `stage1/cohere/values` on claude/port-cohere-second-slice-050g6a (653c9be),
  plus the media query port beside it. The question is whether its test holds the port.
- **Stream C's inline charCodeAt and length, the 2^53 integer printing, and the size-class allocator**, on
  cloud/hx74vgw-alloc (3e98c99).

Everything ran on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. A sub-review did the values
half, under my instructions, in a worktree with a real cohere submodule. Its logs and probes are in
/tmp/claude-0/values/ in this container. I checked its central claim myself by grepping the test's inputs.
The C half is mine.

## Confirmed findings, worst first

### 1. The values test doesn't hold the port at the edges of its character classes

The comparison itself is honest. It compares the whole serialized tree byte for byte, error messages included,
against Go cohere's real `Parse`, on four runs: native sanitized with leak detection, native release, Node,
and the JavaScript backend. Nothing is filtered out, and all 7 of its own mutants are caught.

What it asks is too narrow, the same kind of gap R2 found in the gitignore test. Eleven plausible bugs pass the
whole test, natively and on Node. For each one, the sub-review added it to the test's own `mutants` list, ran
the test, and got "the mutant agrees with Go cohere":

| Mutant in the port | An input that catches it | Go cohere | Mutated port |
|---|---|---|---|
| `alphaNum`'s digits stop at 8 | `#999` | one word `#999`, isColor, isHex | word `#` plus number `999` |
| `isColor`'s digits stop at 8 | `#999`, `#123459` | isColor true | false |
| `alphaNum`'s lowercase stops at y | `#zz` | one word, isHex | `#` plus `zz` |
| `alphaNum`'s uppercase stops at Y | `#Zz` | one word | split |
| `alphaNum` takes `_` | `#_a` | `#` plus `_a` | one word |
| `wordEnd` also ends at `=` | `a=b` | sourceIndex 0, columns 1..3 | sourceIndex 1, columns 2..4 |
| the strict calc check ignores case | `CALC(1 -2)` | `ParserError: Syntax Error at line: 1, column 8` | parses |
| the url argument check ignores case | `URL(a b)` (strict) | words `a`, `b` | one word `a b` |
| the tokenizer's `isURLArg` ignores case | `Url(//x)` (loose) | `ParserError: Expected closing parenthesis...` | parses |
| `unicodeRange`'s digits stop at 8 | `u+99` | `u+99` | `u+9` plus `9` |
| `--` is a word only before more text | `--` | `TypeError: Cannot read properties of undefined (reading '0')` | word `--` |

`color: #999` is an everyday value, and a port that splits it into `#` and `999` passes the gate.

Why these get through: none of the test's fixed cases or generated pieces has `#` followed by `9`, `y`, `z`,
`Y`, `Z` or `_`, an `=` in a value, or uppercase `CALC(` / `URL(` / `Url(`. I confirmed the fixed-case side
myself: none of those strings appear in values_test.go, sample-cases.txt or the Go side.

The media query port has the same gap in one place:

- The mutant: a closing quote ends a string whatever quote opened it.
- It survives the test's 4,133 cases. This was run on Node only.
- `({'a"b'}:x)` catches it: Go gives a media-feature `{'a"b'}`, a colon, and the value `x`, while the mutant
  gives one feature, `{'a"b'}:x`.

**What it would take:**

- Add these as fixed cases:
  - values: `#999`, `#z1`, `#Z1`, `#_a`, `a=b`, `CALC(1 -2)`, `URL(a b)`, `Url(//x)` (loose), `u+99`, `--`;
  - media query: `({'a"b'}:x)`.
- The general fix: give the generator pieces that start at every edge of every character class, and both cases
  of every keyword the parser compares.

**The port itself, as it stands: no divergence found.** The sub-review ran 30,023 random values, covering
printable ASCII, controls, non-BMP characters and CSS fragments in both cases, plus 23 real-world values. All
gave the same bytes from:

- Go cohere;
- the port on Node;
- native sanitized, exit 0 with leak detection on;
- native release;
- the JavaScript backend;
- postcss-values-parser 2.0.1 itself.

On the media query side, the port on Node matched Go cohere on 30,000 random params. So the gap is in what the
test asks, not yet in the port.

## Stream C's changes: no findings

- **Inline charCodeAt.** The fast path runs only when the cached UTF-16 count equals the byte length. That is
  exactly the all-ASCII case, since every multi-byte UTF-8 sequence is fewer units than bytes. A fraction,
  -0, NaN, an index past the end, an uncounted string and any non-ASCII string all go out of line.
  - I mutated the bound to `position <= length`. The string sweep and ascii_scan.a both catch it, as an ASan
    heap-buffer-overflow, not through -Werror.
- **Cached length.** Every string starts with its caches empty: both allocators (string.c and input.c), every
  stack piece, every static string. Nothing changes a string's bytes or length after its units are counted.
  The one in-place growth, `builder_add`, is a separate builder struct, never a string.
  - I'm noting one risk for later, not finding it now. A future reuse in place for strings, like A2's reuse
    for objects and arrays, would have to reset `units` and `index`.
- **Whole numbers below 2^53 written directly.** Negatives, zero, NaN and infinities are handled before this
  path. A fraction fails `value == (double)(uint64_t)value`. The 16-byte buffer holds 2^53 - 1's 16 digits.
  No shorter digits read back as an integer there, so its own digits are the shortest.
- **Size classes with free lists.** Slots are 16-byte multiples carved from malloc's 16-aligned chunks. Under
  ASan the classes are off, and every value goes through malloc and free. With them on (ADAMIC_SLABS) free
  slots are poisoned.
  - Freed slots come back last in, first out, so a release build hands a dead target's address to the very
    next object of its size. Weak's side table is keyed by address.
  - I re-ran the round-three Weak probes as release builds: `a_reuse.a`, `a_churn.a`, `a_together.a`,
    `a_narrowfree.a`. All four runs agree on each, apart from `a_reuse.a`'s documented native difference.
    That one reads `gone`; no stale handle ever reads the recycled object.
  - The oracle now also holds the release build's output to the sanitized run's.
- **Gate on 3e98c99.** gofmt and vet are clean; native, lower and load pass. I didn't run the oracle package on
  this branch.

## Not covered

- The values mutants were screened on Node. Only the 11 confirmed ones ran natively through the test, and the
  media query mutant ran on Node only.
- The values gaps tests and GAPS.md weren't reviewed.
- Nothing ran on macOS or arm64.
- No oracle run of the alloc branch, as above.
