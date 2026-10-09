# ECMAScript matcher reference

`Compile` parses UTF-8 or WTF-8 patterns to immutable backtracking bytecode.
`CompileUTF16` preserves exact JavaScript pattern strings, including lone
surrogates. `Program.New` makes the
mutable RegExp instance. `Exec` accepts UTF-16 code units, including lone
surrogates, and returns capture spans (capture zero is the whole match).
`Capture{-1, -1}` is undefined; an equal nonnegative start/end is empty.
`ExecString` is a convenience for well-formed UTF-8 input.

The embedding runtime supplies lastIndex after ToLength. Global and sticky
execution update it on success and reset it on failure; other executions leave
it unchanged. The embedding exposes indices only when `d` requests them. This
package does not implement JavaScript object descriptors, custom exec methods,
coercion, or the String method wrappers. The execution oracle observes the
native exec calls those wrappers make.

## Representation and semantics

Instructions implement character sets, ordered branches, jumps, capture saves,
assertions, backreferences, repetition initialization/choice/end, lookaround,
and acceptance. Branch frames copy input position, captures and repetition
registers. Repetition resets every capture in its atom before entering the
body. The zero-progress check applies only once the minimum has been satisfied.
Quantifier bounds and counters retain arbitrary precision without unrolling.

Lookaround runs a compiled assertion subprogram and commits only its first
successful result. Negative lookaround restores the incoming captures. A
lookbehind subprogram emits terms in reverse order and consumes code points or
code units backwards; alternatives keep their source order. No syntax tree or
Go predicate callback is retained in the executor. NativeDeclarations serializes instruction targets, range sets, strings, registers,
flags and assertion subprograms for the same backtracking C interpreter in
internal/native/runtime/regexp.c. Stage 0 compiles patterns before emitting C.

Canonicalize uses generated Unicode 17.0.0 tables, matching Node 24: C/S simple
case-fold mappings for `u`/`v`, and full uppercase without expansions or
non-ASCII to ASCII conversion otherwise. `testdata/generate-canonicalize.py`
records source hashes and generates exact Script aliases. The Unicode data
license is in `testdata/unicode-license.txt`.

`PropertyProvider.Lookup` is the UCD integration seam. The default
`UnicodeProperties` provider uses `internal/unicodeproperties`, merged from
`origin/cloud/grok-regex-canonicalize`, for every ECMAScript property, including
binary properties, Script_Extensions and v properties of strings. Compiling
snapshots every range and string. The old `GoProperties` provider remains an
explicit opt-in test stand-in for proving unavailable-data errors are loud.

`TestSharedCanonicalize` compared the local sparse maps with Grok's
CanonicalizeUnicode over all 1,114,112 code points and CanonicalizeLegacy over
all 65,536 code units: zero differences. Both implement Unicode 17.0.0, so the
matcher keeps the identical sparse compile-time maps.

A zero `StepLimit` is unlimited. The oracle sets ten million instruction steps
per execution. `ErrStepLimit` is an error, never a failed match. A separate
catastrophic `(a+)+$` probe uses 1,000 steps and requires that error.

Matching exposed parser representation and Annex B cases beyond the original
acceptance corpus. The parser now preserves nested v negation, handles escaped
surrogate pairs as one Unicode atom, splits raw astral literals into legacy
code-unit atoms, and preserves legacy escapes and class ranges. Added Node
controls hold braced surrogates, legacy octal bounds and identity/control
escapes, and v class-operation syntax.

## Reproduction

Source `/workspace/adamic-tools/env.sh` in this environment, or the env.sh path
printed by `bash cloud/setup.sh` elsewhere. Test output must go to files.

```sh
go test -v -count=1 -timeout 5m ./internal/regexp > /tmp/regex-matcher-package.log 2>&1
python3 internal/regexp/testdata/run-mutants.py > /tmp/regex-matcher-mutants.log 2>&1
gofmt -l cmd internal > /tmp/regex-matcher-gofmt.log 2>&1
go vet ./... > /tmp/regex-matcher-vet.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/regex-matcher-full-gate.log 2>&1
```

The execution corpus uses test262 commit
`7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`, the same pin as the parser corpus:

```sh
git clone https://github.com/tc39/test262.git /tmp/test262-regexp
git -C /tmp/test262-regexp checkout 7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd
node internal/regexp/testdata/extract-matches.js /tmp/test262-regexp internal/regexp/testdata/matches.json.gz > /tmp/regex-matcher-extraction.log 2>&1
```

The extractor wraps the realm's native exec method, retaining method name,
arity and non-constructibility. It replays each ordinary execution with `d`
and verifies captures and lastIndex before recording indices. Patterns and input strings
are encoded as UTF-16 arrays. Undefined captures and named keys are preserved.
Identical pattern/flags/input/lastIndex cases are deduplicated. The checked-in
corpus is gzip-compressed JSON; it contains actual calls reached through exec,
test, match, matchAll, replace and split, rather than guessed input strings.

## Observed results and limits

Environment: Node 24.19.0 (Unicode 17.0), Go 1.27.1, clang 20.1.8. `nproc` was 5;
cgroup CPU quota was four processors. Setup printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (24s)
setup: done in 24s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The parser baseline agreed on 5,746 test262 patterns and 630 generated cases.
The matcher corpus contains 127,369 distinct executions: 127,369 compared
without disagreements, including the 881 formerly refused properties. No
property executions are skipped.
These executions span 2,791 distinct pattern/flag combinations. Source
directories include exec (72 rows), test (15), match (32), matchAll (6),
replace (56), split (62), and search (2), in addition to the other RegExp tests.
Deduplication attributes identical executions to their first source.
The fixed-seed differential (seed `0x875`) compares 10,000 generated pairs.
Canonicalize tests compare 5,403 mapping probes; targeted controls cover 94
executions, 53 exact UTF-16 pattern/input probes, and the injected
property-string test covers four more. A provider-storage mutation probe checks
that compiled strings stay immutable.

All four requested real implementation mutants were run and restored. The
Node capture/index oracle, rather than compilation, caught them:

| Mutant | Oracle failures | Witness |
|---|---:|---|
| Greedy becomes lazy | 12 | `a+` on `aaa`: span `[0,1]` instead of `[0,3]` |
| Captures not reset on repetition | 1 | `(a\|(b))+` on `aba`: capture 2 retains `[1,2]` instead of undefined |
| Lookbehind runs forwards | 2 | `(?<=([ab]+)([bc]+))$` on `abc`: fails instead of captures `[0,1]`, `[1,3]` |
| Folding ignores the u/v distinction | 2 | `k` with `i` matches Kelvin sign, which Node rejects |

Three additional mutants were caught: aliasing provider string storage fails
`TestMatcherProviderSnapshot` against Node; treating an instruction-limit error
as a failed match fails `TestMatcherStepLimit`, and compiling an unavailable
property as an empty set fails `TestMatcherPropertyProviderStrings`. Each had
one oracle failure. The mutation script restores all changed files in `finally`.

The final touched-package run passed with all execution and parser probes,
with zero unexplained failures. `gofmt -l cmd internal` and `go vet ./...` produced
empty logs and exited zero. The repository-wide command exited 1: every package passed except
`internal/native`. Its normalization points sweep hit its 300-second context
deadline after 925,042 lines; native and Node were killed and the final line
was truncated. `internal/oracle` passed in 1,384.803s. The focused normalization
rerun passed all 1,114,112 code points with zero mismatches in 54.282s:

```sh
go test -v -count=1 -timeout 10m -run '^TestNormalizeMatchesNode/points$' ./internal/native > /tmp/regex-matcher-normalize-isolated.log 2>&1
```

The full extraction accounting, including every file and reason omitted, is
`testdata/matches-extraction.json`. Of 1,879 files, 1,670 completed, 192 were
parse-negative, 16 failed and one timed out. There were 1,930 oversized calls
(inputs or patterns longer than 4,096 units/characters) in 452 files and 76
exotic executions excluded from recording. This is not complete test262
matching coverage. Oversized Unicode sweeps and coercion/property-descriptor
cases require more harness work. A timeout is reported, not silently swallowed.

Twelve extraction failures require secondary realms. One is the poisoned-stdlib
test, which native Node passes but the recording wrapper cannot run. Three
Symbol.replace getter tests fail identically on uninstrumented Node 24.19.0:
`flags-tostring-error.js`, `get-flags-err.js`, `get-unicode-error.js`. That is an
observed Node/test262 disagreement outside Pattern evaluation. The single
one-second timeout is `character-class-escape-non-whitespace.js`.

The initial ordinary quadratic search in `S15.10.2_A1_T1.js` exceeded a one-million
step budget. It completes with the documented ten-million budget. Results were
not changed to avoid backtracking. Production execution is unlimited by default.

## Complete property integration and the native door

Merged origin/cloud/grok-regex-canonicalize at 26e4e76 and switched the default
provider at 6766c9d. Both commits were pushed before starting native emission.
The Go matcher rerun compared all 127,369 extracted executions, including all
881 previously unavailable properties: zero disagreements, zero skips. Its
10,000 randomized pairs, 5,403 folding probes, 94 controls, 53 exact UTF-16
probes and parser baseline also passed. Seven Go mutants were caught again.

The imported unicodeproperties package passed in 603.433 seconds. It checked
1,715 property expressions over 1,114,112 code points (1,910,702,080 observations),
7,906 property-of-strings sequences, all legacy canonicalization values and
4,294,967,296 legacy equivalence comparisons. Its Unicode equivalence checks
examined 5,653,004,288 code points, including iu/iv classes. No disagreements.
These are observations from its Node oracle, separate from matcher coverage.

Stage 0 now lowers regexp literals and constant-pattern RegExp constructors
(including const string aliases and concatenation). Patterns compile to immutable
C descriptors at compile time. Constructors still evaluate their arguments, so
initialization checks are retained. Runtime instances own their lastIndex,
source and ordered flags. The C interpreter consumes UTF-16, snapshots captures
and repetition registers for backtracking, and executes compiled lookaround
subprograms. Unicode properties and v strings are compiled into descriptors;
the runtime does not load Unicode data or parse patterns.

Native test and exec, String.match, lazy matchAll and its next/for-of iteration,
string-replacement replace/replaceAll, split and search are implemented. Results
retain captures, index, input, named groups and d indices. Undefined captures
stay undefined, including split captures. Compiler library declarations are
corrected before checking, so unsafe use of a capture as a string is rejected.
Named dictionary keys not present in the pattern return undefined. Counted
ownership covers metadata, iterators and early exit from loops.

The sanitized C interpreter compared all 127,369 extracted executions and
10,000 generated pairs (seed 0x875): zero disagreements in capture values,
indices, named indices and lastIndex. The stage 0 regexp.a and failure fixtures compare
source Node, the JavaScript backend, sanitized C and release C, with a separate
leak run. It covers every added method, replacement tokens including named
and missing groups, Unicode and lone surrogates, null results, .source escapes,
iterator cloning, direct next values, early exits, non-global validation and
null narrowing. A separate methods sweep (testdata/sweeps/regexp_methods.a) covers 63,960
wrapper probes: 25 patterns, 10 inputs, six
lastIndex values, 16 replacements and seven split limits. Native budget failure is a loud exit 70;
production matching remains unlimited.

Reproduce the new checks with output redirected:

```sh
go test -v -count=1 -timeout 10m -run '^TestRegExp' ./internal/native ./internal/load ./internal/lower > /tmp/regex-native-all.log 2>&1
go test -v -count=1 -timeout 5m -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp.a' ./internal/oracle > /tmp/regex-native-fixture.log 2>&1
python3 internal/native/testdata/run-regexp-mutants.py > /tmp/regex-native-mutants.log 2>&1
python3 internal/regexp/testdata/generate-native-folds.py
```

Native scope is typed built-in operations with string inputs and string
replacements. Replacement callbacks, computed named-group keys, dynamic
patterns or flags, custom exec/iterator methods, property-descriptor/prototype
semantics and generic coercion are not implemented. Unsupported source uses
are refused; this is not a general JavaScript object runtime. C quantifier
bounds above uint64 are refused at compile time rather than truncated. Nullable
match arrays are supported; a union holding both null and undefined requires
a representation tag and is refused. The test262 extraction omissions described
above remain; comparing every recorded execution does not mean every test262
file or String wrapper assertion has been compiled by stage 0.

All fourteen native-path mutants were run and restored. Greedy-as-lazy,
no capture reset, forward lookbehind and folding without the u distinction
failed Node stdout comparison. The runtime folding witness uses `(.)\\1`
on `KK` with i and iu; a literal-capture witness was masked by the compiled
set and was strengthened before claiming the check. Starting matchAll at zero
and corrupting restored search lastIndex also failed Node stdout comparison.
Leaking array metadata failed the leak check. Turning the native budget error
into a failed match failed its required exit-70 check. Removing undefined from
exec capture types failed TestRegExpCaptureTypes. Reading an iterator result
by the value slot failed its missing/reordered-field test. Treating a global
match index as always present failed Node stdout comparison. Escaping a slash
inside a class, or changing source bytes, failed the Node source oracle. Merging
distinct lone-surrogate pattern keys failed the C capture/index oracle. The mutation runner requires
the intended result witness, so a compiler error does not count as detection.

The first wrapper sweep timed out on Node. The minimal reproducer on Node
24.19.0 was `const r=/[\\q{ab|a|}]/gv; const [zero]=[0,NaN];
r.lastIndex=zero; "🌍a🌍".replace(r, "")`: timeout exited 124 after three
seconds. With a literal zero lastIndex it returned normally. This is an
observed Node replacement slow-path hang, not a matcher capture disagreement.
The completed wrapper sweep uses the nonempty v alternatives `ab|a`; empty v
strings remain covered by the Go and C execution oracles. This is an explained
Node-oracle coverage gap, not a skipped execution from the 127,369-case corpus.

The wrapper sweep lives below testdata/sweeps so the flow tests' single-directory
fixture glob does not instrument all 63,960 library probes. That instrumented sweep grew beyond 5 GB while collecting events, and the
flow run failed with missing temporary trace files; the ordinary regexp.a
fixture and failure fixtures still exercise the added IR in the flow suite.

Final regression commands (logs remain in /tmp):

```sh
go test -count=1 -timeout 15m -parallel 4 ./internal/load ./internal/lower ./internal/ir ./internal/javascript ./internal/regexp ./internal/flow > /tmp/regex-native-packages.log 2>&1
go test -count=1 -timeout 15m -parallel 4 ./internal/lower > /tmp/regex-native-lower.log 2>&1
go test -v -count=1 -timeout 15m -parallel 4 ./internal/native > /tmp/regex-native-package-full.log 2>&1
go test -count=1 -timeout 15m -parallel 4 ./internal/flow > /tmp/regex-native-flow-final.log 2>&1
go test -v -count=1 -timeout 5m -parallel 4 -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp' ./internal/oracle > /tmp/regex-native-oracle-final.log 2>&1
go test -v -count=1 -timeout 5m -run 'TestNativeAgreesWithNode/internal/oracle/testdata/sweeps/regexp_methods.a' ./internal/oracle > /tmp/regex-native-wrappers-final.log 2>&1
go test -v -count=1 -timeout 20m -parallel 4 -run '^TestCountsAreRecorded$' ./internal/oracle -args -update-counts > /tmp/regex-native-counts.log 2>&1
```

The first combined package command failed on test-source type errors while new
fixtures were being tightened. Those sources were corrected and lower/flow
reruns passed. Load and regexp passed in the combined run; ir and javascript
have no direct tests. The complete native package passed in 231.477 seconds,
including 1,114,112 normalization points with zero mismatches; flow passed in
133.198 seconds. The filtered stage 0 oracle passed every ordinary regex and
failure fixture. The wrapper sweep passed source Node, backend, sanitized C,
release C and leak checking with zero disagreements. The repository-wide gate
was not rerun for this native step; the changed packages, filtered oracle and
complete counted-fixture gate are the selected validation. Vet, gofmt and
whitespace checks produced empty logs.

The final ordinary regexp fixture also checks optional match index/input
metadata, global-match metadata absence, and user IteratorYieldResult objects
with missing or reordered done fields. The final count gate passed in 311.871
seconds: regexp.a allocated/freed 454 values, the wrapper sweep allocated/freed
448,046, and all prior recorded rows stayed unchanged. The final flow rerun
passed in 110.758 seconds. The full native package run remains the earlier
231.477-second run; after the final helpers, the RegExp-specific package checks,
Node fixtures, source oracle and counted-fixture gate were rerun.

TestRegExpSourceNode compares 13 source/flag combinations to Node and preserves
one raw lone-surrogate byte sequence, including solidus escaping within regular
and nested v classes and line terminators. The C execution harness keys patterns
by their original UTF-16 units when supplied, so lossy JSON string decoding
cannot merge different lone-surrogate patterns. A two-pattern C regression and
its mutation demonstrate that distinction.
