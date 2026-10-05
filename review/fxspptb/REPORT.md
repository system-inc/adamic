# Stream R, task fxspptb: today's merges, read again

What landed on main from 8800ed0 to 7055396 (`git log --first-parent 8800ed0..origin/main`), read for what
an oracle can't see. Everything below was run on Linux x86-64 (4 cores, FMA present): Go 1.27.0, Ubuntu clang
18.1.3 (with libclang-rt-18-dev installed for ASan), Node v24.21.0 (V8 13.6.233.17-node.53, the exact V8 the
ports name). The gate on main 7055396 was green here before I changed anything.

`probe.sh <file.a> [arguments]` runs a probe three ways (Node on the source, native under ASan, UBSan and
LeakSanitizer, and the JavaScript backend on Node) and prints each one's output and exit code.

## Confirmed findings

### 1. Native strings have no length limit, and Node's do (today's toUpperCase and readTextFile reach it)

V8's longest string is 2^29 - 24 UTF-16 units (0x1FFFFFE8). Past it, Node throws `RangeError: Invalid
string length`, which is a panic to Adamic. Nothing in `internal/native/runtime` checks it: the only length
check is the size_t overflow in `adamic_string_concat` ("string too long"). Native keeps going and prints a
different answer, which looks right. The JavaScript backend agrees with Node in every case, since it is Node.

| Probe | Node and the JavaScript backend | Native |
|---|---|---|
| `long_upper.a`: `"ß".repeat(300000000).toUpperCase()` (today's case.c) | prints 300000000, then `adamic: panic: RangeError: Invalid string length`, exit 70 | prints 300000000, then 600000000, exit 0 |
| `big_file.a /tmp/claude-0/big.txt` (`truncate -s 600M`; today's input.c) | `cannot read /tmp/claude-0/big.txt: failed`, exit 0 (readFileSync throws ERR_STRING_TOO_LONG) | `/tmp/claude-0/big.txt: Ok, 629145600 code units`, exit 0 |
| `long_repeat.a`: `"x".repeat(600000000)` (older code, for contrast) | `adamic: panic: RangeError: Invalid string length`, exit 70 | prints 600000000, exit 0 |

The repeat row shows this is runtime-wide and older than today's merges. Today's code adds two new ways
to reach it, one of which (a file) comes from outside the program. A file over 2 GiB would also fail on Node
with ERR_FS_FILE_TOO_LARGE ("failed"). I didn't run that one.

A fix would have to count UTF-16 units, not bytes: a 4-byte character is 2 units, a lone surrogate 1. It would
panic with `RangeError: Invalid string length` wherever a string is made over 0x1FFFFFE8 units (allocate,
concat, repeat, pad, case mapping), and readTextFile would return its `failed` error instead. I haven't
written that fix: it touches every string-making path, and the decision is Ahra's.

### 2. The ieee754 bit sweep can't certify the port to the last constant: 6 of 8 one-change mutants survive it

The sweep compares bits, not text, and catches fused arithmetic: with `-ffp-contract=off` and both
`#pragma STDC FP_CONTRACT OFF` lines removed, and `-march=haswell` added, 2174 of 1281787 answers differ.
But one-ulp changes to constants mostly get through it. Each mutant below changes one line of ieee754.c on main
7055396, and was run with `go test -count=1 -run TestIeee754 ./internal/native` (driver: /tmp/claude-0/mutate.sh):

| Mutant | Sweep |
|---|---|
| kernel_cos C1 one ulp up | caught, 238 answers (cos 126, sin 112) |
| kernel_sin S1 one ulp up | caught, 1609 answers (cos 716, sin 893) |
| kernel_cos C6 one ulp up | survives |
| log Lg7 one ulp up | survives |
| log ln2_lo one ulp up | survives |
| rem_pio2 pio2_1t one ulp up | survives |
| rem_pio2 medium-range bound 0x413921FB to 0x413921FA | survives |
| rem_pio2 third iteration dropped (`i > 49` to `i > 4900`) | survives |

Most survivors are probably equivalent or nearly so. A one-ulp change in a tiny polynomial tail or a
low-order constant flips a result only when the exact value sits within about 2^-70 of a rounding
boundary. pio2_1t is recomputed away by the second iteration exactly where cancellation is large, and
moving the medium/large boundary by one high word sends those inputs to kernel_rem_pio2, which gets the same
answer. I didn't prove any of them equivalent.

The third iteration is not equivalent, and I proved it: `near_pio2.py` finds the doubles nearest n·π/2
(n < 2^19) that cancel 68 to 73 bits. On those, the mutant prints `sin 8.85920166919226e-17` where Node prints
`8.859201669192259e-17` (642615.9188844458). They're now `internal/oracle/testdata/trig_reduction.a`. It agrees
with Node on main, and it fails the mutant through the real oracle test
(`go test -run 'TestNativeAgreesWithNode/internal/oracle/testdata/trig_reduction' ./internal/oracle`: exit 1,
"stdout differs"; exit 0 unmutated). That's the one change I propose for the gate, in its own commit.

So what guards the constants is the source comparison against V8, not the sweep. That comparison is below
(section "The fdlibm port"), and it found every constant identical.

### 3. contract_test.go skips silently, and always on a Mac

`TestArithmeticIsNeverFused` is the only `t.Skip` in the repository. `go test` without `-v` prints nothing for
a skipped test, so the gate looks the same whether or not it ran. On macOS there is no `/proc/cpuinfo`, and
`-march=haswell` is x86-only, so on every Apple silicon Mac (the machine it was written to protect) it always
skips. The merge message for 9a4b334 says the fix is "held by contract_test.go on a machine with fused
multiply-add". On a Mac it isn't held by that test.

What does hold it on a Mac, simulated here with `-ffp-contract=off` removed and `-march=haswell` added to every
build: `navigation.a` in the oracle ("Salt Lake City to Seoul: 9448.314548882256 km" against Node's
9448.31454888226) and `TestNumbersParseExactlyAsJavaScriptDoes` (parseInt with radix 36 is off in the last
bit). So today the Mac is covered, but by fixtures that weren't written for it, and the skip says nothing.

Here on Linux with FMA, the test does its job. Mutant M1 (delete the `-ffp-contract=off` line only) is caught
by it alone: `got "5.5511151231257827e-17 -5.5511151231257827e-17 -6.1679056923619804e-18\n"`, with sanitize
both false and true. The ieee754 sweep stays green under M1, since baseline x86-64 has no FMA instructions and
the pragma holds. That is not -Werror.

Suggested, not done: on `runtime.GOARCH == "arm64"`, build without `-march`, since every arm64 has FMA, and
run. Skip only on an x86 without FMA, and then say so where the gate's reader will see it.

### 4. THIRD_PARTY_NOTICES.md doesn't list every V8 file in the runtime

V8's BSD-3-Clause text is reproduced in full, so V8's terms travel with every binary. But the V8 entry's "In
Adamic" line names only `parse.c` and `number.c`. These also carry V8 notices and aren't named:

- `ieee754.c`: V8's modifications, "Copyright 2016 the V8 project authors", beside Sun's fdlibm notice. The
  fdlibm entry names it, but V8's BSD license covers V8's changes.
- `hypot.c`: from `src/builtins/math.tq`, "Copyright 2019".
- `dtoa.c`: from `bignum.cc`, `bignum-dtoa.cc`, `dtoa.cc` and `conversions.cc`, all "Copyright 2011", and
  `builtins-number.cc`, "Copyright 2016". I checked each year against V8's own files at the tag, and they match.

Two smaller differences:

- The V8 text in THIRD_PARTY_NOTICES.md says "Copyright 2006-2011, the V8 project authors" (V8's old LICENSE.v8
  wording). V8 13.6.233.17's top-level LICENSE says "Copyright 2014". The body text is the same.
- V8's LICENSE.fdlibm says "Copyright (C) 1993-2004 by Sun Microsystems", and the notices quote the 1993 line
  from ieee754.cc's header. That line is preserved verbatim in ieee754.c, which is what fdlibm's notice asks.

Every ported file keeps its original notice verbatim at its top: ieee754.c, hypot.c, dtoa.c, parse.c. The
Unicode entry covers case_tables.h. That entry gives unicode.org as the source, and case_generate.go fetches from
ICU's 78.2 tag. case_generate.go says why, and that the bytes are the same; they are pinned by SHA-256.

### 5. A comment in ieee754.c is wrong, not the code

The header says "C's libm is never called here except for sqrt and fabs". kernel_rem_pio2 also calls
`scalbn` and `floor` (ieee754.c:520, 521, 562, 594, 607), as V8 does: in C++ they resolve to std::scalbn and
std::floor. Both are exact, so there's no miscompile. The comment should name them.

## Suspicions that turned out fine, and what showed it

**The fdlibm port, line by line.** I made V8's src/base/ieee754.cc (13.6.233.17) into C with sed, applying only
the changes the header claims: `base::bit_cast` to the two memcpy helpers, `static_cast<T>` to `(T)`, the
leading underscores, `while (false)`, `V8_WARN_UNUSED_RESULT`, and the digit separators. Then I ran
`diff -b` against ieee754.c. 556 changed lines remain, and every one is one of these:

- The file header and includes; namespaces and `#undef`s gone; `V8_INLINE` to `static inline`; helpers made
  `static`.
- `legacy::pow` left out (313 lines, as the header says).
- `std::numeric_limits<double>::signaling_NaN()` and `quiet_NaN()` to `NAN`, and `infinity()` to `INFINITY`.
  A NaN's payload can't be seen without typed arrays, and the sweep treats every NaN as one.
- `*const_cast<volatile double*>(&x)` to `*(volatile const double *)(&x)`.
- Calls to V8's own functions (`log`, `log1p`, `exp`, `expm1`, `atan`) renamed `adamic_math_*`. I checked that
  no unrenamed call is left behind that C would quietly send to libm. The only libm calls left are `fabs`,
  `sqrt`, `scalbn` and `floor`, the same ones V8 calls.
- `NegateWithWraparound<int32_t>(lx)` to `(0u - lx)`, and `SubWithWraparound(hx, 0x3FF00000)` to
  `((uint32_t)hx - 0x3FF00000u)`. lx and ly are uint32_t (ieee754.c:1259, 1581). In C++ the int32 result of
  the negation is converted back to uint32 by `lx | ...` before the shift, and the subtraction is OR'd with
  the uint32 lx before it's compared with 0. So both sides compute the same 32 bits through the same unsigned
  shift and comparison.

There are no other differences: every constant, table entry, branch and statement is the same text.

**Node 24 really uses fdlibm for sin and cos.** V8 can be built with `V8_USE_LIBM_TRIG_FUNCTIONS`, which uses
glibc's sin and cos instead. Node 24.21.0's `node --v8-options` has no `--use-libm-trig-functions` flag, which
exists only in such builds, and the sweep agrees bit for bit.

**hypot.c against math.tq.** For 0 arguments the emitter writes `0.0`. For 1 argument: |a| against
sqrt((a/a)²)·a, which is exactly a. For 2 and 3 arguments, the fast path's arithmetic is the loop's, step for
step. With 3, the loop's compensation after the second term is ((pA+pB)-pA)-pB, which is the fast path's
`(powerA + powerB) - powerA - powerB`, and its last sum is (pA+pB)+(pC-comp). Infinity is checked before NaN on
every path, and +0 is returned for all zeros.

**Arguments evaluated in order.** C doesn't fix the order of function arguments or compound-literal
initializers, but the emitter snapshots each argument into a temporary first
(`adamic_math_hypot(3, (const double[]){adamic_temporary_12, ...})`). `order.a` agrees three ways.
`Math.hypot(...xs)` is refused as NotYet ("can't lower a SpreadElement yet").

**case.c.**
- Strings are canonical WTF-8: concat rejoins a lone high surrogate and a lone low one (string.c:61), and case
  mapping never makes a surrogate, so its lone-surrogate passthrough can't split a pair.
- The 12-byte scratch fits the longest mapping (3 code points of 4 bytes).
- Allocation failure panics in adamic_allocate.
- Final_Sigma is ICU's algorithm, and the sweep covers every code point in each of its context positions.
- `made_strings.a` agrees three ways under LeakSanitizer. An emitter mutant that hands toUpperCase's result to
  `e.snapshot` (never released) instead of `e.own` is caught by LeakSanitizer on it: "384 byte(s) leaked in 6
  allocation(s)", stdout unchanged, exit 1. The existing case_mapping.a already catches the same mutant, so I
  didn't promote made_strings.a.

**input.c.**
- Every failure path frees what it took: the name after open, the buffer on a read error, and the buffer on a
  failed realloc (then panic).
- The errno mapping is the same as adamic.mjs's error-code mapping, and a NUL in the path is `failed` on both.
- A directory opens on Linux, then fails its read with EISDIR, which is handled.
- Kind strings are immortal, and the message is made by concat.
- The one disagreement is the size limit in finding 1.

**dtoa.c** (read against V8 by a sub-agent, with harnesses it ran):
- Bignum, bignum-dtoa, the builtins' order of checks, and buffer sizes match V8. The longest output is 108
  bytes, `(-5e-324).toExponential(100)`, against a 128-byte buffer.
- 9.4 million inputs (every power of two, every 1eK and its neighbors, denormals, random bits, 385,082 exact
  ties) were byte-identical to Node 24 under ASan and UBSan.
- BignumDtoa alone equals FastDtoa-then-Bignum: FastDtoa gives up on exact ties in both modes.
- Its two mutants were both caught: precision ties (`>= 0` to `> 0` in generate_counted_digits) and shortest
  ties (round-half-even reversed).
- One weakness it measured: dtoa_test.go kills the shortest-tie mutant on only 2 lines, both from one sampled
  value (±1.7860489244912068e+15).

**A wrong Node is loud.** With the machine's default Node 22.22.0 on PATH, the native package fails two
tests: TestCaseTablesMatchNodesUnicode ("Unicode 17.0.0 and Node is Unicode 16.0") and the pow sweep (V8's pow
changed between the two).

## What I didn't cover

- Nothing ran on macOS or on arm64. Finding 3's Mac behavior is a simulation: `-march=haswell` on every build
  stands in for arm64's FMA.
- The other streams' mutants: the merge messages give counts (eleven, eleven and eighteen), but the mutants
  themselves aren't in the repository, so I couldn't check whether any was killed by -Werror. I ran my own
  instead (above). None of mine is a -Werror kill.
- I proved none of the five surviving one-ulp mutants equivalent, beyond the reasoning above.
- The JavaScript backend's new code (toUpperCase, Math, toExponential and toPrecision, input) wasn't read line
  by line, only run in every probe.
- case_generate.go wasn't read line by line. Its output is swept against Node over every code point.
- The 3 GiB file case of finding 1 wasn't run.

## Notes for whoever runs this gate next on Linux

- The ASan runtime may be missing (`libclang_rt.asan-x86_64.a`). Install `libclang-rt-18-dev`.
- input_test.go drops to uid 65534 when run as root, so Node has to be installed where `nobody` can reach it
  (I used /opt/node24). The repository has to be reachable for `nobody` as well.
