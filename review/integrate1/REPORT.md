# Stream R, round two: what integrate-1 brought

This covers main 7055396 to 80c3098: B's radix.c, C's adversarial readers (the string limits, the stack check
and normalize), B2's one-word slots and counted unions, and A2's borrowed parameters, globals at exit and the
counts table. I looked for what an oracle can't see.

I ran everything on Linux x86-64 with FMA: Go 1.27.0, clang 18.1.3 and Node v24.21.0 (V8 13.6.233.17).
Three sub-reviews ran in parallel, one on borrowing, one on the limits and stack and normalize, and one on
radix.c. I stopped the radix one before it finished and did radix.c myself. Where I re-ran a sub-review's
finding myself, it says so; where I didn't, it says that too.

Every probe in this directory runs three ways with `review/fxspptb/probe.sh <file> [arguments]`: Node, native
under ASan, UBSan and LeakSanitizer, and the JavaScript backend. None of the probes is in the gate. Two new
oracle fixtures are, and they are on this branch: trig_reduction.a (round one) and normalize_long_marks.a.

## Confirmed findings, worst first

### 1. String() of some powers of two is wrong (silent; older than integrate-1, from 1afaa16)

`power_of_two_string.a`: `` `${2 ** 976}` `` gives these results:

| Run | Output | Exit |
|---|---|---|
| Node and the JavaScript backend | `6.386688990511104e+293` | 0 |
| Native | `6.3866889905111034e+293` | 0, sanitizers silent |

How widespread it is:

- 46 of the 2,098 powers of two differ, for example 2^-1017, 2^-957, 2^-808 and 2^-652, and their negatives.
- Of 1,000,000 random doubles, none differ.
- I found it because the radix sweep's radix-10 rows go through String(): 2 of its 842,896 lines.

The cause is in `shortest_digits` (number.c:17). For each p, it asks printf for the closest p-digit decimal
and stops at the first one that reads back. ECMAScript wants the fewest digits that read back, closest only
among those.

A power of two has half the gap below it that it has above. So the closest 16-digit decimal can fall outside
the narrow lower half-gap and fail to read back, while a farther 16-digit decimal inside the wide upper
half-gap does read back. Node writes that one; native goes on to 17 digits.

The fix is already in the runtime: dtoa.c's V8 BignumDtoa shortest mode, which `toExponential()` with no
argument uses and which matched Node over 9.4 million inputs in round one. String() could take its digits
from there.

### 2. Array.from passes 0 for undefined to a `number | undefined` parameter (silent; B2)

`from_undefined.a` runs `Array.from({ length: 3 }, (value: number | undefined, index: number) => ...)`:

| Run | Output |
|---|---|
| Node and the JavaScript backend | `undefined at 0`, `undefined at 1`, `undefined at 2` |
| Native | `0 at 0`, `0 at 1`, `0 at 2` (exit 0, sanitizers silent) |

The cause:

- `lower/from.go` marks the callback's first parameter always-undefined only when the checker types it
  unknown or undefined.
- Annotated `number | undefined`, the parameter is a packed slot instead, where undefined is the reserved NaN.
- `native/from.go:19` always passes `{.reference = NULL}` for the value, and those bits read as the double 0.

The fix: pass the undefined of the parameter's own type (`adamic_maybe_number_pack` of a missing value for
`number | undefined`), or say NotYet for any first parameter that isn't a reference.

### 3. The stack check never fires on a tail call (silent at -O1 and -O2; C)

Re-run myself: `stack_tail_call.a`, `count(100000)` and then `count(1e9)`, with `return count(n - 1)`:

| Run | Output | Exit |
|---|---|---|
| Node and the JavaScript backend | `adamic: panic: RangeError: Maximum call stack size exceeded` | 70 |
| Native | `0`, `0`, `end` | 0 |

`stack_forever.a`, the sub-review's run: Node panics at once, and native at -O2 never ends (it was killed after
60 s).

The cause: the check compares the frame's address, and clang turns a self tail call into a loop, so that
address never moves. A fix has to keep the call a call, or count depth instead of measuring the stack.

### 4. An undefined string argument reaches the runtime as NULL (crash; C, and older)

Re-run myself:

| Probe | Node | Native |
|---|---|---|
| `normalize_undefined_form.a`: `'é'.normalize(form)` with `form` undefined | `1`, since undefined means NFC | UBSan, member access within null pointer at normalize.c:239, exit 1. The sub-review's unsanitized -O2 build segfaults (exit 139). |
| `pad_undefined_fill.a`: `'x'.padStart(3, fill)` with `fill` undefined | `  x` | UBSan at string.c:440, exit 1 |

The pad one predates integrate-1. The cause is in `lower/object.go`: a `string | undefined` argument lowers to
ir.String, which is NULL when undefined, and these two methods' signatures accept one. These are the only
string methods whose TypeScript signature takes undefined for a string argument.

### 5. Native recursion depth differs from Node's, either way (C)

These are the sub-review's runs:

- `stack_deep.a`: Node panics at `sum(10000)`, while native prints `50005000` and `5000050000` first.
- `stack_heavy.a`, under `ulimit -s 1024`: native panics at depth 3000, where Node prints `156001`.
- `stack_over.a` with three 120 KB arguments: native segfaults (exit 139) where Node panics (exit 70). The
  limit is measured from a frame below argv and the environment, but RLIMIT_STACK counts those too. 360 KB of
  environment does the same.

The commit says the depths differ, and they do. But each of these is a program whose output differs between
Node and native, and no fixture can hold one.

### 6. Releasing globals at exit leaves them in place, so LeakSanitizer can't see a global's own leak (A2)

Re-run myself: `global_cycle.a` is `a.next = a` on a global `a`. All three runs print `a1` and exit 0, and
LeakSanitizer reports nothing, but `adamic build --count` says `allocations 3 frees 1`.

The cause: `releaseGlobals` (emit.go:413) releases each global but leaves the pointer in its static. Statics
are roots to LeakSanitizer, so what a global holds directly always looks reachable.

The sub-review's mutant on the generated C added a retain to each global, which is what a dropped release
looks like. That is silent as things are, and caught (127 bytes) once the globals are set to NULL after their
release. So the fix is to zero each global after releasing it.

### 7. counts.md's stack_overflow.a row depends on the machine (A2 and C)

Re-run myself, with the counted binary:

| `ulimit -s` | Allocations |
|---|---|
| 4 MiB | 82,920 |
| 8 MiB | 170,302 (the recorded row) |
| 16 MiB | 345,064 |

TestCountsAreRecorded fails wherever the limit isn't 8 MiB, which is logged in `counts_16m.log`. I expect a
different frame size (another clang, or arm64) to move it too, but I haven't seen that. A fix pins the limit
for counted runs, or records that fixture's counts without the parts that depend on depth.

### 8. readTextFile has no limit natively, and Node's limit is in bytes (C; corrects my round one)

The sub-review's runs: Node's `readFileSync(path, 'utf8')` fails from 0x1FFFFFE8 bytes on.

| File | Node | Native |
|---|---|---|
| 536,870,887 bytes | reads it | reads it |
| 536,870,888 bytes | `failed` | Ok |
| 600 MB of `é`, which is 300M units | `failed` | Ok |

So the check belongs on the byte count, not on UTF-16 units as I suggested in round one.

### 9. Two checks no test can fail (C)

- A canonical reordering that gives up after six moves passed the normalize sweep, whose strings are at most
  four letters. It is a real bug: `a` + U+0305 × 7 + U+0323 normalizes wrongly under it. I added
  **`internal/oracle/testdata/normalize_long_marks.a`**, which agrees with Node on main and fails that mutant
  through the oracle test ("stdout differs").
- Removing concatenation's string-length check (`if (0)`) passed all of `./internal/native` and
  `./internal/oracle`, the sub-review's run. No test drives a string past the limit, because a string that
  long takes about 1 GB per side. A C-level test that calls the check with lengths, not strings, would hold it.

### 10. A method read as a value compiles and then panics (loud; older)

`method_value.a`, `const g = k.f;`: native prints `adamic: panic: compiler bug: a field the checker proved is
there is missing` and exits 70, where Node prints `t1a1`. It should be NotYet when it's compiled.

## Suspicions that turned out fine

- **radix.c against V8's DoubleToRadixStringView**, read line by line by me. The logic, the comments, the
  2200-byte buffer, its midpoint, the delta, the round-to-even and the carry are V8's. The differences are
  mechanical:
  - the CHECKs are dropped;
  - the flush-denormals branch is dropped, since Node doesn't flush denormals;
  - Modulo becomes fmod, as V8's is away from Windows;
  - Double::Exponent and NextDouble are rewritten for positive finite values only;
  - number.tq's Smi path is left out, and it gives the same digits for those integers.
  - The order of checks is number.tq's: the radix first (ToIntegerOrInfinity, so NaN is 0 and a fraction
    truncates), then radix 10 as ToString, then 0, NaN and the infinities.
- **radix.c's mutants**:
  - The existing sweep catches 8 of the sub-review's 10.
  - The two survivors are the carry mutants: dropping `integer += 1`, and dropping the `digit + 1 < radix`
    test. Compiled side by side with the original, neither changed one output in about 7.5 million inputs.
    Those were every radix, k minus 1 to 64 ulps for k up to 2000, radix powers ±16 ulps, and 200,000 random
    fractions per radix.
  - Why I think no input can reach them: a carry into a digit equal to radix − 1 needs the previous remainder
    r with r + delta > 1. That would already have rounded the previous step and ended the loop, so a carry
    could start only at the first digit. There it needs the fraction within half an ulp of 1, which the exact
    split never gives.
  - So both are equivalent: argued and searched, not proven. That matches what B's merge message said of one
    of them.
- **The radix sweep against Node**: 842,896 answers, byte-identical apart from the two radix-10 lines that are
  finding 1.
- **B2's slots**:
  - Every slot read and write of `number | undefined` goes through slotted and unslotted, or through pop's and
    find's own unpack. A CheckedCast on a packed field would fail in clang, not miscompile.
  - Readonly covariance (`number[]` read as `(number | undefined)[]`) is safe, because packing changes only
    undefined and NaN, and the reserved NaN can't come from arithmetic.
  - Probes that agree three ways with no leaks: NaN as a Map key and in includes and indexOf
    (`nan_keys.a`; x86 makes 0/0 a negative NaN, which arm64 doesn't), readonly covariance and optional class
    fields (`covariance_and_fields.a`), and `??` unions with NaN and -0 (`counted_unions.a`).
  - Type predicates are refused. Generics, optional parameters on function values and `boolean | undefined`
    in slots are NotYet.
- **A2's borrowing**, about 30 probes from the sub-review. A callee that overwrites the storage its argument
  came from (a global, a field, an element, a map value, a cell, a class field through this, a module export)
  stays safe, because each of those reads is retained for the statement. A callee that reassigns its
  parameter keeps it owned. Borrowed values that escape (returned, stored, captured, boxed) are retained.
  - One inference, not observed: borrow.go exempts only closures. Once named functions can be values, map
    will hand them uncounted elements while their parameters are borrowed, the case borrow_map_overwrite.a
    guards for closures.
- **C's string limit elsewhere**, from the sub-review. Exactly 0x1FFFFFE8 units succeeds and one more fails,
  as on Node, across repeat (`x`, `é` and `😀`), template and `+`, glued surrogate halves, pad with a fill cut
  through a pair, case mapping, replace and replaceAll, join, and normalize (NFKD of U+FDFA). One more thing
  from it: `'İ'.repeat(268435445).toLowerCase()` segfaults Node itself, where native panics with the
  RangeError.
- **C's normalize**: the Unicode 17.0.0 tables match Node, the sweep compares bytes over every code point,
  ownership is clean, and five of its seven mutants were caught.

## What I didn't cover

- Nothing ran on macOS or on arm64.
- Findings 5, 8 and the concatenation mutant in 9 are the sub-review's runs, which I didn't repeat. Findings 1,
  2, 3, 4, 6, 7 and 10 I re-ran myself.
- I ran no mutant of the stack check itself.
- The JavaScript backend's new code (radix, normalize, unions, Array.from) I only ran, I didn't read it.
- The two carry mutants are equivalent by argument and search, not by proof.
- The sub-review noted, outside this range: native replaceAll is quadratic, taking 11.6 s on 80,000 characters
  where Node takes about 0.1 s.
