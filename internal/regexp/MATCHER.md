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
Go predicate callback is retained in the executor. Instruction targets, range
sets, strings, registers, flags and assertion subprograms can be encoded for a
future C executor. C emission and native-runtime integration are not built here.

Canonicalize uses generated Unicode 17.0.0 tables, matching Node 24: C/S simple
case-fold mappings for `u`/`v`, and full uppercase without expansions or
non-ASCII to ASCII conversion otherwise. `testdata/generate-canonicalize.py`
records source hashes and generates exact Script aliases. The Unicode data
license is in `testdata/unicode-license.txt`.

`PropertyProvider.Lookup` is the UCD integration seam. It returns inclusive
ranges and strings. `GoProperties` is explicitly a stand-in using Go's Unicode
General_Category and Script tables, including their exact aliases. Binary
properties, Script_Extensions, and properties of strings return a typed
`UnavailablePropertyError`; they never silently become empty sets. Tests inject
an emoji provider to exercise property strings with v set intersection,
subtraction, union and backwards consumption. Set operators operate on compiled
canonical ranges and strings; alternatives consume longest strings first.

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
The matcher corpus contains 127,369 distinct executions: 126,488 compared
without disagreements, and 881 explicitly refused unavailable properties.
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
