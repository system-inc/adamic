# First increment: validation and measurements

Branch `codex/stage1-ts-printer`, from `origin/codex/typescript-scanner` at
`ed2477e538f54e772c62593de9bab4ddeca0d4ab`. Only `stage1/cohere/tsprinter/` changes.
This increment is the complete shared doc layout engine and a partial expression printer. It does
not claim whole-file formatting, the complete expression slice, statements, declarations or types.

## Setup

`bash cloud/setup.sh > /tmp/ts-printer-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`; `nproc` printed `5` (cgroup quota: four CPUs).
Setup printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (21s)
setup: done in 21s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node v24.19.0. Cohere is pinned to
`715ba94f3608a6500086b1076ce5cb7e51b836db`, npm Prettier to 3.9.6, and the TypeScript 6.0.3
source to `050880ce59e30b356b686bd3144efe24f875ebc8`. The scratch Prettier installation used here was
`/tmp/graphql-printer-prettier`; it already existed and its version is asserted in both npm oracles.
The TypeScript source was `/tmp/adamic-tsc-strictness/typescript`, with its commit asserted by the test.

## Slice gate

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-canonical-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-canonical-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  > /tmp/ts-printer-final-green.log 2>&1
```

Exit 0. Relevant output:

```text
197 files, 0 full-file parse refusals, 149852 supported maximal expression fragments
17 unported shapes return NotYet on native, Node and backend; Go and Prettier format every proving input
149852 expression fragments byte-identical
--- PASS: TestExpressionsAgainstGoAndPrettier (325.43s)
5072 docs identical to Go and Prettier on native, Node and backend; sanitizer and leak checks pass
--- PASS: TestDocumentsAgainstGoAndPrettier (12.06s)
--- PASS: TestMutants (4.81s)
PASS
ok github.com/system-inc/adamic/stage1/cohere/tsprinter 577.660s
```

The compiler suite walks all 77 compiler source files and all 120 Adamic `.ts` files, including other
stage-1 ports and gap programs. The independent Go selector formats maximal supported expressions
as standalone expression statements. Unsupported parents are traversed for supported children.
The 149,852 count includes 1,203 generated cases. It is not a count of formatted complete files.
[results/coverage.json](results/coverage.json) records accepted root kinds and rejected candidates.

Source on Node, native under ASan/UBSan, the JS backend and native release all match Go byte for byte;
the npm comparison of accepted expressions is strict and has no exceptions. The separate unported
proving corpus records one exact anonymous-function spacing difference between Go and npm Prettier,
with its source and both byte strings in [testdata/prettier-differences.json](testdata/prettier-differences.json).
That input is `NotYet` in the port. The document comparison is independently constructed from Go's
existing doc generator and from npm's actual doc builders; it includes all document kinds, shared
identity, conditional states, width boundaries, Unicode, tabs, alignments, suffixes and trims.

The boundary audit added numeric receivers, update distinctions, Boolean coercion and long member
contexts to the generated cases; optional stopping boundaries and spaced blank lines are explicit
NotYet proofs. The doc-kind whitelist is immutable module data, not rebuilt for each doc node.

The initial compiler command took several minutes; final expression validation took 325 s, including
lowering, two native builds, all executions, leak checking and npm formatting. The parentheses
mutant's compile/run subtest took 247 s. Those are observed command/subtest durations, not a profile
attributing cost to any specific compiler analysis. No compiler or runtime performance claim is made.

## Three mutants

Every mutant is compiled successfully and must exit 0 with empty stderr before a differing stdout
counts as caught. All three were caught on native with ASan/UBSan and on Node.

| Mutant | First independent Go mismatch | Native / Node |
|---|---|---|
| Group ignores its width fit result (`&&` becomes `||`) | Doc line 18: `\n` instead of `tab\tbctab\tbc` | Both caught, normal exit |
| Fill never packs a pair (`separatorMode = 1`) | Doc line 6: `xéé\n\nword\n\n ` instead of `xéé\nword\n ` | Both caught, normal exit |
| Required expression parentheses disappear | Expression line 81: `0.5 * body.mass * body.vx * body.vx + body.vy * body.vy + body.vz * body.vz;` instead of `0.5 * body.mass * (body.vx * body.vx + body.vy * body.vy + body.vz * body.vz);` | Both caught, normal exit |

`TestCompilerGaps` also passes the three recorded Node/diagnostic proving programs in `gaps/`.
`GAPS.md` distinguishes two stage-0 NotYet cases from an intentional 0.1 refusal.

## Default width

An additional complete-corpus run used width **120**, the driver's default. Go's compiled batch
helper was run with `ADAMIC_TS_BENCH_WIDTH=120`; its independently formatted answers became the
expected outputs. Native release and source Node were invoked without a width argument; npm's
printer was invoked at 120. All stderr files were empty, all processes exited 0, and all 149,852
outputs matched byte for byte. Output:

```text
149852 texts byte-identical at default width 120 on native release, Node, Go and Prettier
```

This additional run did not repeat the sanitizer/backend executions at width 120. Those execute
at width 80 in the main slice test; the doc corpus independently varies widths and indentation.

## Throughput

All runs used the same 149,852 accepted fragments at width 80. Three complete batch processes per
implementation, sequentially with no other test running. Each sample includes startup, input
decoding, parsing, layout, output encoding and file output. Every sample's full output is checked
against Go's bytes. Native is a release `-O2` build; Node runs the Adamic source. Go uses a compiled
overlaid test binary whose benchmark entry formats the same batch protocol. Prettier is the actual
pinned npm library. These are isolated-fragment rates, not whole-TypeScript-file throughput.

| Side | Seconds, three runs | Median texts/s |
|---|---|---:|
| Native | 1.803, 1.800, 1.782 | 83,244 |
| Node | 1.327, 1.431, 1.419 | 105,573 |
| Go cohere | 1.875, 1.817, 1.775 | 82,490 |
| Prettier | 42.654, 42.869, 42.123 | 3,513 |

Native is 23.7 times Prettier's rate on this fragment corpus and 0.79 times Node. Native and Go
are approximately equal here (their run ranges overlap); the measured median ratio is 1.01.
This is a baseline; no profile was taken in this increment.
[results/throughput.json](results/throughput.json) keeps unrounded samples. Reproduce after the main
test has retained its corpus and native artifacts:

```sh
python3 - <<'PYOVERLAY' > /tmp/ts-printer-bench-overlay.json
import json
from pathlib import Path
root = Path.cwd()
print(json.dumps({"Replace": {
    str(root / "cohere/internal/format/javascript/adamic_expressions_test.go"):
    str(root / "stage1/cohere/tsprinter/testdata/expressions_side_test.go")
}}))
PYOVERLAY
(cd cohere && go test -c -overlay=/tmp/ts-printer-bench-overlay.json \
  ./internal/format/javascript -o /tmp/ts-printer-go-oracle) > /tmp/ts-printer-go-build.log 2>&1
python3 stage1/cohere/tsprinter/testdata/measure.py \
  /tmp/ts-printer-canonical-corpus /tmp/ts-printer-canonical-artifacts/port \
  /tmp/ts-printer-go-oracle /tmp/graphql-printer-prettier \
  > /tmp/ts-printer-throughput.json 2> /tmp/ts-printer-throughput.err
```

## Other checks and limits

```sh
gofmt -l cmd internal stage1/cohere/tsprinter > /tmp/ts-printer-gofmt.log
go vet ./... > /tmp/ts-printer-final-vet.log 2>&1
/tmp/adamic-json-cohere --no-fix --no-cache stage1/cohere/tsprinter/*.ts \
  > /tmp/ts-printer-cohere-gate.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|numbers|objects|classes|sorting|indexing|updates|spreads)\.a$' \
  > /tmp/ts-printer-oracle-test.log 2>&1
```

All exit 0; gofmt and vet output are empty. Cohere reports `276 rules`, `8 checked`, `100% Adamic-ready`.
The filtered oracle passes 13 matching fixtures, plus its parent test, in 25.603 s, holding source Node,
sanitized native, release native and the JS backend, with leak checks. The regex also matches
optional/maybe/narrowed/undefined number/string fixtures, which are included in the 13.
An initial incorrectly shallow oracle filter matched no fixtures; it was corrected and the real
run above passed. The full repository test gate was not run: the new package alone took ten minutes,
and the selected oracle covers the runtime features this port uses. No `internal/` file changes.

Remaining expression families, comments, source normalization, whole files, statements, declarations,
types, JSX, embedded printers and non-default expression options remain outside this increment.
See [GAPS.md](GAPS.md). Test logs stay in `/tmp`, not in the commit.

## Sequence expression family

Continued on the same branch from `f798fd6`. Setup on this unit printed Go/clang/Node/submodules
ready in 0 s each, cache warm in 15 s, and done in 15 s on 5 processors (four-CPU quota, 17.6 GB).
`nproc` printed 5. Pins and options are unchanged.

The generator adds 1,212 sequence cases: 12 explicit-boundary/context cases, 240 long sequences,
and 960 compositions with binary/logical operators, arrays, calls, unary operands and member
receivers. All 197 files parse in Go; the expanded corpus contains 151,139 fragments including
2,415 generated cases. `results/sequence-coverage.json` records this increment separately from the
first increment's baseline. Whole files, unsupported expression families and statements are still
outside this claim.

Commands (each test's output goes to the named log):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-sequence-final-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-sequence-final-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-sequence-final.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/sequence' > /tmp/ts-printer-sequence-mutant.log 2>&1
```

The sequence mutant replaces its separator comma with an empty string. Native and Node both
compiled/finished normally (exit 0, empty stderr); the byte comparison caught the missing comma
between `getJsxNamespace(location)` and the following source-file expression in TypeScript's
compiler. The mutant run used the earlier 150,179-fragment corpus, before the 960 compositions
were added, and passed in 260.334 s (native mutant subtest: 255.61 s).

Two failed Node-first comparisons informed the port: top-level sequences require parentheses;
explicit nested sequences must not be flattened. Both are now generated regression cases, rather
than removed corpus inputs. The final cohere gate reports 276 rules, 8 checked, 100% Adamic-ready.
`go vet ./...`, `gofmt -l cmd internal stage1/cohere/tsprinter`, and `git diff --check` are clean.
The same filtered oracle command above passed all 13 fixtures in 24.062 s. The full repository
suite was not run; this unit changes only the printer slice and its independent corpus generator.

Final output (exit 0):

```text
197 files, 0 full-file parse refusals, 151139 supported maximal expression fragments
17 unported shapes return NotYet on native, Node and backend; Go and Prettier format every proving input
151139 expression fragments byte-identical
--- PASS: TestExpressionsAgainstGoAndPrettier (336.28s)
ok github.com/system-inc/adamic/stage1/cohere/tsprinter 336.285s
```

This holds the entire expanded corpus to Go and npm Prettier without accepted-case exceptions,
on source Node, sanitized native, native release and the JS backend; the separate leak run passes.

## Assignment printer family

Continued from sequence commit `fbed1b7`, with the same toolchain and pins. Only the printer slice
changes. Assignment strategies and chain selection follow `print_assignment.go`; short-argument
classification follows `utility_call_arguments.go`. Binary and member contexts use their existing
Go printers' assignment branches. The default string quote preference is unchanged; the lone-short
argument classifier also reproduces Go's double-quote default inside unary arguments.

The independent Go selector accepts every implemented assignment operator and supported target;
it traverses rejected parents as before. No file is omitted or fails to parse. The generator adds
6,855 cases: 6,000 operator/target/right-hand-side/outer-context combinations and 855 chain cases,
with chains up to 20 segments. The final corpus is 150,713 maximal fragments, including 9,286 generated
cases, from 197 files. Newly accepted larger assignment parents reduce the total fragment count.
`results/assignment-coverage.json` preserves the exact accepted and rejected shapes.

Commands:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-assignment-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-assignment-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-assignment-test.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/assignment' > /tmp/ts-printer-assignment-mutant.log 2>&1
```

The source Node and independent npm comparisons pass every fragment with no accepted-case
exceptions. Destructuring assignments are explicit gaps: `[a,b]=items` is held to its exact
`assignment-pattern` refusal, and Go/Prettier both prove the input formatable. There remain 17
proving inputs, with assignment support replacing the old `EqualsToken` gap by this narrower one.
Cohere reports 276 rules, 8 checked, 100% Adamic-ready. Vet, gofmt and `git diff --check` are clean.

A separate 16-input audit found that an optional call is not a lone short argument in Go: a
`ChainExpression` does not pass `isLoneShortArgument`. With a long assignment target and
`g(f?.()).x`, the initial port broke after `=`, whereas Go broke inside `g`'s arguments. The
classifier now excludes optional calls, and all 16 audit inputs are permanent generator regressions.
The final test uses `/tmp/ts-printer-assignment-final-corpus`,
`/tmp/ts-printer-assignment-final-artifacts`, and `/tmp/ts-printer-assignment-final.log`.

The assignment mutant drops the operator's text. Its run compared the 150,697-fragment corpus
before the optional-call audit; native and Node both finish normally with empty stderr and differ
at the very first case: `this.x  x;` instead of `this.x = x;`. Exit 0, mutant subtest 256.99 s,
complete mutant command 262.465 s. The optional-call correction does not affect this proving case.

The actual embedded Prettier fork is now an additional permanent oracle, through
`testdata/embedded.mjs`, alongside npm Prettier's unchanged strict comparison. It independently
parses the original source on V8 with the vendored TypeScript and ESTree plugins. The current
bundle set is the one committed at cohere's recorded submodule pin; its version is asserted as
3.9.6. Its full 150,713-case comparison was also run directly:

```sh
node stage1/cohere/tsprinter/testdata/embedded.mjs \
  /workspace/adamic/cohere/internal/format/prettier/bundles \
  /tmp/ts-printer-assignment-final-corpus/cases.json \
  > /tmp/ts-printer-assignment-embedded.txt 2> /tmp/ts-printer-assignment-embedded.err
cmp /tmp/ts-printer-assignment-final-corpus/answers.txt /tmp/ts-printer-assignment-embedded.txt
```

Both exit 0; stderr is empty. The running final native gate was compiled before this extra harness
assertion was added; this direct comparison supplies the same assertion for that run. Future slice
runs invoke both Prettier implementations automatically.

Final assignment output (exit 0):

```text
197 files, 0 full-file parse refusals, 150713 supported maximal expression fragments
17 unported shapes return NotYet on native, Node and backend; Go and Prettier format every proving input
150713 expression fragments byte-identical
--- PASS: TestExpressionsAgainstGoAndPrettier (334.67s)
ok github.com/system-inc/adamic/stage1/cohere/tsprinter 334.674s
```

Source Node, ASan/UBSan native, the JS backend, native release and the separate leak run all pass.

## Conditional increment

Conditional expressions compose with the existing expression families, including all three nested
positions and chained receivers. The generator adds 5,000 Cartesian compositions and 240 nested
cases (depths 1 through 20). The corpus contains 150,832 maximal fragments from the same 197 files,
with no full-file parse refusals. Counts are fragments, not whole-file formatting claims.

Two oracle failures guided corrections: `??` inside a conditional branch retains parentheses, and
flattened binary chains retain their ancestor stack so an enclosed ternary receives its own group.
Neither failure is excluded from the selector. `results/conditional-coverage.json` preserves the
complete accepted and rejected counts.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-conditional-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-conditional-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-conditional-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/conditional' > /tmp/ts-printer-conditional-mutant.log 2>&1
```

Cohere's source-only check (`--no-fix --no-cache stage1/cohere/tsprinter/*.ts`) reports 276 rules,
8 checked, 100% Adamic-ready; vet and `git diff --check` are clean. A directory-wide cohere check
also reports formatting in the existing independently maintained JS oracle and evidence files;
that is not the Adamic source lint scope used here. No compiler or parser file changes.

The ternary mutant changes the sole `? ` separator to `: `. Native and Node both compile/run
normally, exit 0 with empty stderr, and differ at corpus line 97: `index + (next() < 0.01 :
next() * 1000 : 0);` instead of the ternary using `?`. Mutant subtest 255.81 s; complete command
261.705 s. Its catch is an output mismatch, not a compiler or runtime failure.

Final ternary gate (exit 0, 392.433 s): 150,832 fragments byte-identical to Go cohere, npm
Prettier 3.9.6 and cohere's actual embedded fork; source Node, native ASan/UBSan, JavaScript backend,
leak detection and native release all pass. Sixteen remaining proving inputs return their exact
NotYet reasons on every execution. This gate does not cover arbitrary complete source files.

## Object value increment

Object literals compose with the accepted core through properties, shorthand values, computed
keys and spreads. The port uses Go's property assignment strategies and its short-key exception,
source-driven object wrapping and forced wrapping for homogeneous arrays of multi-item objects.
Its ES5 key-name predicate uses Go's Unicode categories directly, not TypeScript's different
identifier tables. Escaped spellings stay quoted when their printed text differs from the value.

The generator adds 1,680 object cases: nine keys by ten values by three source layouts by six
contexts, plus 60 growing-width cases. The final corpus has 152,387 maximal fragments from 198
files, no parse refusals. An extracted root object is explicitly parenthesized in the independent
Go selector; otherwise it would become a statement block when parsed outside its original return
or initializer context. No fragment is excluded for a formatting disagreement. Optional values
inside an object are accepted; the object itself cannot continue an optional chain. Methods and
expanded object arguments remain loud gaps. Source-only cohere: 276 rules, 9 checked, 100%
Adamic-ready; vet and whitespace checks pass.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-object-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-object-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-object-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/object' > /tmp/ts-printer-object-mutant.log 2>&1
```

The property mutant removes the sole property-colon doc. It compiled and ran normally with
empty stderr and exit 0 on native and Node; both differ at case 251: `({ left undefined, right
undefined });` instead of `({ left: undefined, right: undefined });`. Subtest 260.74 s, total
265.928 s. Its 152,220-case corpus preceded the object-spread parentheses correction and the
167 additional optional-value/boundary fragments; those later changes do not affect its proving
input or the colon printer. The current unmutated source agrees with Go on that input.

Final object gate: exit 0, 403.666 s; all 152,387 fragments byte-identical to Go, npm Prettier
3.9.6 and the embedded fork on source Node, native ASan/UBSan, JS backend and native release.
The separate leak check passes. Sixteen unported proving inputs retain their exact NotYet reasons.

## Call argument increment

Expanded object and array arguments now compose with every accepted expression family. Layout uses
cohere's flat, last-argument-expanded and all-arguments-broken states. Adjacent arguments of the
same type disable expansion, as do concise numeric arrays in multi-argument calls. Required broken
arguments propagate to outer groups; trailing commas, numeric fills and CommonJS/AMD special
layouts are held to the original printers.

The generator adds 2,784 call-layout cases: 2,304 combinations of twelve argument shapes and
480 growing-width cases through four callee spellings. The existing seed contributes all 500 random
expressions, including the 94 previously declined for array arguments. Total 154,822 maximal
fragments, 19,084 generated, from 198 files; no parse refusals. Fifteen explicit gaps remain.
`results/arguments-coverage.json` preserves counts. Source-only cohere reports 276 rules, 9 checked,
100% Adamic-ready; vet and whitespace checks pass.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-arguments-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-arguments-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-arguments-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/arguments' > /tmp/ts-printer-arguments-mutant.log 2>&1
```

The argument mutant replaces the call's conditional trailing comma by an empty doc. Native and
Node compile/run normally, exit 0 with empty stderr, and differ at case 10 (the repository's
n-body constructor array): long `new Body(...)` argument lists lose their final comma. Both use
the final 154,822-case corpus. Mutant subtest 263.21 s; complete command 269.398 s.

Final argument gate: exit 0, 417.315 s; all 154,822 fragments byte-identical to Go, npm Prettier
3.9.6 and the embedded fork on source Node, ASan/UBSan native, JS backend and native release.
The separate leak run passes. Fifteen unported proving inputs retain their exact NotYet reasons.

## Member-call increment

Every supported callee expression now composes with calls and constructors. Member-call chains
linearize their parser nodes and docs, segment the base/member/call groups, apply cohere's
factory and short-receiver heuristics, and choose flat or expanded layout. Curried calls prioritize
the inner argument group when appropriate. A constructor callee containing a call retains its
required parentheses. Optional lookup/call tokens are read from each link, not the inherited
optional-chain flag. The latter initially produced `tracing?.pop?.()` instead of `tracing?.pop()`;
the repository corpus caught it.

The generator adds 2,533 cases: 2,430 base/argument/suffix/context combinations, 90 chains up to
30 calls, and 13 constructor/curried/identifier boundaries. Both Go and npm Prettier drop the
TypeScript placeholder's redundant parentheses; Go's Babel-only `Parenthesized` branch is not
copied into the port. All 198 files parse, with no refusals. Accepted fragments decrease because
larger member calls replace their previously selected children.

Cohere required the deliberate chain output buffers' contract to be stated. `chainWalk` uses
`@mutates nodes` and `@mutates docs`, with reasons: it appends reverse-order nodes and their matching
docs for one final reversal. It mutates only fresh buffers created by its caller. The unused
callee lookup was deleted. Source-only cohere now reports 276 rules, 9 checked, 100% Adamic-ready;
vet and whitespace checks pass. A fresh final native gate recompiles this cleanup.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-member-final-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-member-final-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-member-final.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/member' > /tmp/ts-printer-member-mutant.log 2>&1
```

The member-chain mutant removes the method dot. Native and Node compile/run normally with empty
stderr and exit 0; both differ at case 35: `Mathsqrt(squared)` instead of `Math.sqrt(squared)` in
the repository's n-body expression. Subtest 264.33 s, complete command 270.894 s. It uses the
137,638-case corpus before deleting the unused lookup; that deletion removes one fragment from
the port's own source, not the proving input or the mutated dot printer.

Final member source gate: exit 0, 410.302 s; all 137,637 fragments byte-identical to Go, npm
Prettier 3.9.6 and the embedded fork on source Node, ASan/UBSan native, JS backend and native
release. The separate leak run passes. Fourteen unported proving inputs retain their exact
NotYet reasons. The pre-cleanup 137,638-case run also passed (412.697 s).

## Template increment

Untagged interpolated and multiline templates now preserve raw quasi bytes, literal-line break
propagation and absolute indentation. Interpolation expressions collapse to the unlimited-width
printer only when both source and resulting doc lack newlines. Leading and trailing interpolation
newlines are tested independently; parser literal positions include trivia, so the closing-brace
boundary uses the actual token start. Binary interpolations use the template's indentation rather
than adding a second level. Template assignment/property values remain attached to the operator.
The independent oracle caught each of these layout mistakes before the final gate.

Parentheses containing optional arguments or branches are accepted. The stopping-boundary guard
now follows only the active member/call/non-null chain; it does not confuse unrelated optional
subexpressions with the receiver chain. Actual chain-stopping parentheses retain their loud gap.

The generator adds 1,538 cases: 1,440 expression/quasi/newline/context combinations, 90 growing
interpolation lists and 8 boundaries. Total 138,241 maximal fragments from 198 files, with no
parse refusals. Thirteen explicit proving gaps remain. `results/template-coverage.json` records
counts. Source cohere: 276 rules, 9 checked, 100% Adamic-ready; vet and whitespace checks pass.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-template-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-template-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-template-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/template' > /tmp/ts-printer-template-mutant.log 2>&1
```

The template mutant removes `$` from `${`. Native and Node compile and finish with empty stderr
and exit 0, then mismatch at case 99: `console.log(\`{ordered} {checksum.toFixed(3)}\`);` instead
of the repository's interpolated log. Subtest 255.67 s, complete command 261.491 s. Its corpus
predates the final binary-interpolation indentation correction; that correction does not change
the proving input or the mutated dollar printer. The final unmutated corpus includes that input.

Final template gate: exit 0, 406.287 s; all 138,241 fragments byte-identical to Go, npm Prettier
3.9.6 and the embedded fork on source Node, ASan/UBSan native, JS backend and native release.
The separate leak run passes. Thirteen unported proving inputs retain their exact NotYet reasons.

## Arrow and basic body increment

Untyped arrows support defaults/rest, async prefixes, chains through 20 arrows, callee/member
parentheses and assignment-sensitive signatures. Call arguments implement expanded first/last
callbacks, multiple-function composition and dependency arrays. The basic body printer composes
blocks, expression statements, return/throw arguments, empty statements and debugger/jump tokens.
Protected string expressions preserve their directive status. Numeric ancestor paths include body
statements and flattened inner arrows; object receivers starting an arrow body keep parentheses.
Eleven targeted Go/Node probes caught and then closed the missing return-ternary group and missing
object-receiver parentheses. Their inputs are in the expanded generator.

The generator adds 2,672 cases: 27 body shapes, six parameter lists, two async variants and eight
contexts (2,592), plus 80 chains. Total 139,235 maximal fragments from 200 files, no parse refusals.
`results/arrow-coverage.json` records counts. Thirteen explicit proving gaps remain, replacing the
plain-arrow gap by a typed-signature gap. This is not whole-file statement formatting.

The first layout experiment used `functions.ts` and a third shared-contract module, `functionContext.ts`.
A type-only import cycle was refused. Viewing class methods through function-property slots then
revealed a native compiler bug; explicit callback properties work on all three executions. A callback
parameter inferred as `boolean | undefined` is NotYet, so known callers supply required arguments.
A `.some` closure capturing the layout class was also Refused; an equivalent loop reads the async
modifier without capturing that class. No internal, parser or language changes were made.
Earlier refused/aborted gates and mutants do not count as catches. That intermediate source passed cohere:
276 rules, 11 checked, 100% Adamic-ready, plus vet and whitespace checks.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-arrow-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-arrow-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-arrow-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/(arrow|return|statement|body|simple)' > /tmp/ts-printer-arrow-mutants.log 2>&1
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^Test(ClassInterfaceMethodGap|OptionalBooleanFunctionGap)$' \
  > /tmp/ts-printer-interface-gaps-final.log 2>&1
```

Intermediate callback-layout gate exit 0, 683.056 s, while the five mutants shared the four-CPU quota. All 139,235
fragments byte-identical to Go, npm Prettier 3.9.6 and the embedded fork on source Node, native
ASan/UBSan, JS backend and native release; the separate leak run passes. Mutants use this same
final corpus, compile and finish normally with exit 0 and empty stderr on both Node and native:

| Mutant | First mismatch | Subtest seconds |
|---|---|---:|
| Arrow token removed | Case 79, `random.sort((left, right)  left - right)` | 513.13 |
| Return keyword changed to throw | Case 278, the repository's counter callback | 506.86 |
| Expression-statement semicolon removed | Case 278, `count += 1` | 511.07 |
| Block opening brace changed to `[` | Case 278, the counter callback | 502.20 |
| Debugger removed | Case 114242, generated debugger/return body | 524.99 |

Complete mutant command exit 0, 530.843 s. The two separate execution-gap checks also pass,
18.166 s: Node prints `17`, class/interface native exits 70 with the missing-field panic; the
explicit callback prints `17` everywhere, leak-free. Node prints `absent` for the boolean/undefined
callback; stage 0 records NotYet. Its required-parameter form prints `absent` everywhere, leak-free.
Each gap check is proven able to reject the corresponding successful workaround as its mutant.

### Direct-layout correction

The callback-layout benchmark regressed: native 35,571, Node 63,922, Go 52,318 and npm Prettier
3,019 texts/s. `results/arrow-context-timing.json` retains all three samples. Rather than attributing
the regression to the compiler without a controlled comparison, the port now keeps stateful layout
in `Expressions`, splits pure AST helpers into `syntax.ts`, and moves document analysis into
`Documents`. No function-property interface is allocated for each expression. The async-modifier
loop remains. No internal file or parser changes are involved.

The final source passes cohere: 276 rules, 10 checked, 100% Adamic-ready. A fresh complete oracle
and the five family mutants are rerun on this source. Before/after throughput uses exactly the
intermediate 139,235-case corpus, including every expected output byte; fresh source coverage is
recorded separately because reorganizing the port changes its own maximal expression fragments.

Fresh direct-layout gate: exit 0, 641.856 s. All **139,234 maximal fragments from 199 files**,
including the 2,672 new arrow compositions, are byte-identical to Go, npm Prettier 3.9.6 and the
embedded fork on source Node, native ASan/UBSan, the JavaScript backend and native release.
The separate native leak run passes; all 13 loud NotYet proofs pass. The independent document
check passes all 5,072 documents in 21.008 s after the document-analysis move.
`results/arrow-coverage.json` is the final coverage; `arrow-context-coverage.json` preserves the
intermediate workload used by both throughput measurements.

Final-source output mutants, all exit 0 with empty stderr on both Node and native:

| Mutant | First mismatch | Subtest seconds |
|---|---|---:|
| Return becomes throw | Case 278, counter callback | 529.33 |
| Opening block brace becomes `[` | Case 278, counter callback | 532.78 |
| Expression statement loses semicolon | Case 278, counter callback | 544.01 |
| Debugger disappears | Case 114241, generated arrow body | 550.53 |
| Arrow token disappears | Case 79, random-sort comparator | 476.52 |

The combined command in `/tmp/ts-printer-arrow-direct-mutants.log` exited 1 in 556.741 s:
its arrow string anchor no longer matched the formatted source, while the other four subtests
passed. This harness failure is not a catch. The anchor was corrected and the arrow subtest
rerun separately, exit 0 in 489.029 s, recorded in
`/tmp/ts-printer-arrow-direct-arrow-mutant.log`. That normal-run mismatch is the arrow catch.
No refused lowering, crash or sanitizer failure is counted as an output mutant catch.

```sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-arrow-direct-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-arrow-direct-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-arrow-direct-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/arrow' > /tmp/ts-printer-arrow-direct-arrow-mutant.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestDocumentsAgainstGoAndPrettier$' > /tmp/ts-printer-arrow-direct-doc.log 2>&1
python3 stage1/cohere/tsprinter/testdata/measure.py /tmp/ts-printer-arrow-corpus \
  /tmp/ts-printer-arrow-direct-artifacts/port /tmp/ts-printer-go-oracle \
  /tmp/graphql-printer-prettier > /tmp/ts-printer-arrow-direct-timing.json \
  2> /tmp/ts-printer-arrow-direct-timing.err
```

The final arrow checks are package-filtered commands listed above, plus cohere source checks,
`go vet ./stage1/cohere/tsprinter` and `git diff --check`. No full-repository gate was run for this
port-only increment. Earlier sequence, assignment, conditional, object, argument, member and
template families were each committed and pushed after their own green oracle gates and output
mutants; their exact counts and commands remain in the preceding sections.

Controlled direct-layout throughput, three sequential complete processes on the unchanged
139,235-case intermediate corpus (all output samples byte-identical, exit 0, empty stderr):

| Printer | Median texts/s | Seconds, all three runs |
|---|---:|---|
| Native release | 35,415 | 3.932, 4.023, 3.778 |
| Source Node | 65,528 | 2.127, 2.117, 2.125 |
| Go cohere | 51,139 | 2.723, 2.761, 2.706 |
| npm Prettier | 3,063 | 45.461, 43.890, 47.768 |

`results/arrow-timing.json` contains the precise samples. The callback-layout native median was
35,571: the direct-layout median is 0.44% lower, and the timing ranges overlap. This experiment
**does not support the callback split as the cause of the native regression**, and no native speed
improvement is claimed. The final concrete layout avoids the observed interface implementation
bug and reduces the module coupling, but native is now slower than both Node and Go on this
expanded workload. Different fragment counts and newly supported larger parents also prevent a
causal comparison with earlier-family throughput. Diagnosing that cost needs a separate profile;
no compiler/runtime optimization was attempted in this port.

## Optional-chain stopping increment

The normalizer now records active optional-chain boundaries before removing parenthesized nodes.
Raw receiver traversal stops at a parenthesis, so a nested already-stopped chain does not force
extra outer parentheses. The printer preserves the boundary for ordinary members/calls, new
callees, non-null assertions and tags. Calls on a wrapped chain retain the Go adapter's generic
call choice even where an optional outer call can omit redundant parentheses.

The independent Go selector's stopping-chain exclusion is removed. The generator adds 480
compositions: eight chain bases, ten outer links, six contexts, including long paths and nested
parentheses. The first Node comparison exposed the long optional-call layout discrepancy at
case 114578; preserving the ChainExpression call-layout distinction fixed it. That failed gate is
not validation. The former stopping-boundary gap is replaced by `await value` / `AwaitExpression`.
Source cohere passes 276 rules, 10 checked, 100% Adamic-ready.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-optional-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-optional-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-optional-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/optional' > /tmp/ts-printer-optional-mutant.log 2>&1
```

Final optional-boundary gate: exit 0, 424.428 s. **139,716 maximal expression fragments from 199
files**, zero full-file parse failures, all byte-identical to Go cohere, npm Prettier 3.9.6 and
cohere's embedded fork on source Node, native ASan/UBSan, JavaScript backend and native release.
The separate native leak run passes. Thirteen loud NotYet proving programs remain and pass on all
three executions. `results/optional-coverage.json` contains the exact accepted/rejected counts.

The optional-boundary mutant bypasses the required-parentheses branch without discarding the
boundary metadata. Native and Node compile and finish normally with exit 0 and empty stderr,
then both mismatch at case 114115: `obj?.x.x;` instead of `(obj?.x).x;`. Subtest 268.97 s;
complete command exit 0, 274.782 s. Cohere source checks, vet and whitespace checks pass.
No full-repository gate was run; no parser or internal compiler file changed.

### Pushed families in this unit

| Commit | Family |
|---|---|
| `fbed1b7` | Sequence expressions |
| `3f7e19b` | Assignments and chains |
| `5305f40` | Conditional expressions |
| `e191d3d` | Object values and wrapping |
| `f2e5e8a` | Expanded call arguments |
| `80bca68` | Member and curried-call chains |
| `6ccd848` | Interpolated and multiline templates |
| `4990187` | Arrow signatures and basic statement bodies |

Each was pushed after its own green byte oracle and normal-run output mutant checks. This remains
an expression driver with basic statement bodies, not a whole-file statement formatter. Ordinary
function expressions, object methods/accessors, tagged templates, binding patterns, contextual
await/yield, comments/trivia, typed syntax and broader control-flow statements remain unported.
The implementation and proving gaps are explicit; no input is passed through as a fallback.

Final optional-boundary throughput, three sequential complete batch processes including input,
output and startup, all 139,716 expected outputs verified on every sample:

| Printer | Median texts/s | Seconds, all three runs |
|---|---:|---|
| Native release | 36,814 | 3.882, 3.795, 3.763 |
| Source Node | 65,190 | 2.084, 2.184, 2.143 |
| Go cohere | 52,324 | 2.873, 2.670, 2.590 |
| npm Prettier | 3,111 | 44.910, 45.557, 43.415 |

`results/optional-timing.json` retains the precise measurements. Command exit 0, empty stderr:

```sh
python3 stage1/cohere/tsprinter/testdata/measure.py /tmp/ts-printer-optional-corpus \
  /tmp/ts-printer-optional-artifacts/port /tmp/ts-printer-go-oracle \
  /tmp/graphql-printer-prettier > /tmp/ts-printer-optional-timing.json \
  2> /tmp/ts-printer-optional-timing.err
```

Native remains slower than Go and Node on this expanded corpus. The direct-layout controlled
comparison above did not demonstrate a native improvement; no unsupported causal claim is made.

## Named function increment

Untyped named function expressions now compose ordinary, async and generator forms with supported
block bodies. Names are recognized before the parameter list so identifier return types remain
explicitly unsupported. Function callees and expression-statement starts retain parentheses.
First and last argument signatures use the distinct Go expansion rules.

The generator adds 2,400 cases: four prefixes, six parameter lists, ten body shapes and ten outer
contexts. A long lone signature exposed a wrong expansion predicate. Adding an outer group did
not fix it. The independent Go `PrintToDoc` tree showed that the expanded argument signature was
flat text, while the all-broken alternative retained lines. The Go predicate expands a lone
signature when **all** parameters are plain identifiers, not when one is non-plain. Correcting
that predicate passed the full Node-to-Go comparison. Failed exploratory gates are not validation.
The temporary diagnostic overlay changed no cohere or internal files.

Cohere source checks pass: 276 rules, 10 checked, 100% Adamic-ready. Thirteen proving gaps remain,
including the existing anonymous-function npm spacing difference. Every accepted case remains a
strict npm and embedded-fork comparison, without exceptions.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-function-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-function-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-function-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/function' > /tmp/ts-printer-function-mutant.log 2>&1
```

The function-token mutant removes `function` from its generated prefix. Node and native compile
and finish normally, exit 0 and empty stderr, then both mismatch at case 114611: `( named() {});`
instead of `(function named() {});`. Subtest 268.97 s, complete command exit 0, 274.505 s.

Final named-function gate: exit 0, 422.215 s. **142,132 maximal expression fragments from 199
files**, zero full-file parse refusals, byte-identical to Go, npm Prettier 3.9.6 and the embedded
fork on source Node, native ASan/UBSan, JavaScript backend and native release. The separate leak
run passes, as do all 13 exact NotYet proving inputs. Coverage is in
`results/function-coverage.json`. Cohere, vet and whitespace checks pass. No full-repository gate
was run, and no internal compiler or parser file changed.

Named-function throughput: three sequential complete batches, 142,132 texts, every sample held
to all Go expected bytes, normal exit 0 and empty stderr:

| Printer | Median texts/s | Seconds, all three runs |
|---|---:|---|
| Native release | 35,331 | 4.040, 3.975, 4.023 |
| Source Node | 65,664 | 2.165, 2.156, 2.207 |
| Go cohere | 52,320 | 2.703, 2.717, 2.728 |
| npm Prettier | 3,122 | 44.958, 45.523, 47.095 |

Precise samples: `results/function-timing.json`. Command:

```sh
python3 stage1/cohere/tsprinter/testdata/measure.py /tmp/ts-printer-function-corpus \
  /tmp/ts-printer-function-artifacts/port /tmp/ts-printer-go-oracle \
  /tmp/graphql-printer-prettier > /tmp/ts-printer-function-timing.json \
  2> /tmp/ts-printer-function-timing.err
```

Native remains slower than Go and Node on this workload; no speed improvement is claimed.

## Object method and accessor increment

Object methods, getters and setters reuse key, parameter and block-body printers. The key offset
accounts for async and generator tokens. Typed signatures and binding parameters remain loud gaps.
The pure blank-line helper moves to `syntax.ts`; stateful layout stays in the concrete printer.
Source cohere passes 276 rules, 10 checked, 100% Adamic-ready; the expression file is 1,974 lines.

The generator adds 2,560 method combinations (eight names, four prefixes, five parameter lists,
four bodies, four contexts) and 64 accessor combinations. A separate audit of computed short keys
proved the recorded concern: `[x]` and `[1]` with long strings/member chains disagreed on both
source Node and the last green native release. Go's CleanDoc reduces text-only concats to text.
The port now classifies that cleaned text without changing the printed key or removing groups.
Another 64 boundaries cover eight keys, two values and four contexts. The native gate was restarted
on this corrected frozen source; the earlier gate does not count as final validation.

An initial Go selector build failure called the optional-token accessor as a field. Correcting it
to `QuestionToken()` fixes the selector; that failed build is not an oracle catch. The untyped method
gap is replaced by a typed method signature in the 13-input loud-gap corpus.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-method-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-method-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-method-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/(method|computed)' > /tmp/ts-printer-method-mutants.log 2>&1
```

The mutant regex also selects the previously implemented member-chain method-dot mutant. It is
rerun on this corpus alongside the method-key removal and computed-key cleanup reversal.

Intermediate method/accessor gate before the interpolation audit: exit 0, 477.358 s. **144,858 maximal fragments from 199 files**, zero
full-file parse failures, all byte-identical to Go cohere, npm Prettier 3.9.6 and the embedded fork
on source Node, native ASan/UBSan, JavaScript backend and native release. The separate leak run
passes, as do all 13 exact NotYet proofs. Coverage is in `results/method-coverage.json`.

All three selected mutants compile and finish normally, exit 0 and empty stderr on Node and
native, then mismatch only in output:

| Mutant | First mismatch | Subtest seconds |
|---|---|---:|
| Method loses key | Case 117049, `({ () {}, value: 1 });` | 339.46 |
| Computed-key text cleanup is undone | Case 119673, line break after `[x]:` | 336.03 |
| Member method dot disappears | Case 35, `Mathsqrt` | 335.95 |

Complete mutant command exit 0, 344.956 s. Source cohere, vet and whitespace checks pass.
No full-repository gate was run, and no internal compiler or parser file changed.

### Final method/accessor composition gate

The subsequent optional-chain interpolation audit exposed a real port bug that the intermediate
corpus missed. Both source Node and native flattened `(x?.y).z` into `x?.y.z`. The flat preview now
copies the optional-boundary set. There are 48 generated regression compositions plus two newly
extractable own-source fragments. The final frozen source supersedes the intermediate gate:

```sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-method-final-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-method-final-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-method-final-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/(template_preview|method|computed)' > /tmp/ts-printer-method-final-mutants.log 2>&1
```

Both exit 0. Final gate **566.021 s: 144,908 / 144,908** from 199 files, zero full-file parse
failures. All bytes match Go, npm Prettier and the embedded fork on Node, ASan/UBSan native,
JavaScript backend and release native; the separate leak pass and 13 exact NotYet proofs pass.
`results/method-coverage.json` is this final corpus.

The four mutants finish normally with empty stderr on both Node and native. The first mismatch
is line 119723 for computed-key cleanup (433.46 s), line 114171 for optional boundaries in the
preview (436.63 s), line 35 for member dots (440.33 s), and line 117099 for method keys (441.66 s).
The complete mutant command passes in **447.824 s**. Source cohere prints 276 rules, ten checked,
100% Adamic-ready; package vet and `git diff --check` pass.

The intermediate 144,858-text benchmark is retained in `results/method-timing.json`, explicitly
marked as preceding the preview correction. Three complete-process runs each produced Go's exact
bytes: medians native **33,956**, Node **62,494**, Go **50,703**, npm Prettier **3,080** texts/s.
Native is slower than Go and Node on this expanded workload. No full-repository gate was run.

## Ordinary tagged-template increment

The independent selector and generator add 504 tag/template/context combinations. Ordinary tags
reuse raw template and interpolation layout. Tag precedence, optional-chain stopping parentheses,
constructor callees, member chains and assignment values compose. A preliminary scratch probe
caught seven long assignment values breaking after the operator; Go's tagged-value exception fixes
those layouts. Jest `each` tables replace the ordinary-tag proving gap, still 13 exact NotYet inputs.
Pure unsupported-shape classification moved to `syntax.ts`, without callback or interface dispatch.

An initial mutant fixture had incorrectly quoted Go text and failed to build; that is a harness
failure, not a mutant catch. The first complete gate passes **436.142 s**, but two declaration-order
lint findings remained. Moving those pure declarations before their uses fixes lint. The final
source has its own complete gate below, superseding that run:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-tag-final-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-tag-final-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-tag-final-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/tagged' > /tmp/ts-printer-tag-mutant.log 2>&1
```

Final gate exit 0, **423.897 s: 145,440 / 145,440** maximal fragments from 199 files, zero full-file
parse failures, strict Go, npm Prettier 3.9.6 and embedded-fork equality. Source Node, ASan/UBSan
native, JavaScript backend, release native and the separate leak run pass. All 13 gap reasons match.
Coverage is in `results/tag-coverage.json`. The source cohere check prints 276 rules, ten checked,
100% Adamic-ready; package vet and whitespace checks pass. No internal or parser file changed.

The tag-removal mutant compiles and finishes normally, exit 0 and empty stderr on both Node and
native, then mismatches at line 114199: `` `raw`; `` instead of `` tag`raw`; ``. Its subtest passes
in **271.19 s**, complete command **277.182 s**. This was on the semantically identical source before
moving the two pure declarations; the tag printer and mutation anchor are unchanged by that move.
No full-repository gate was run.

### Tagged-template throughput

After both gates ended, the final source was frozen under `/tmp/ts-printer-tag-snapshot` and all
measurements completed before starting the next native compilation. This isolates source changes
and build CPU load from the benchmark. The corpus and release binary are from the final tag gate.

```sh
python3 /tmp/ts-printer-tag-snapshot/stage1/cohere/tsprinter/testdata/measure.py \
  /tmp/ts-printer-tag-final-corpus /tmp/ts-printer-tag-final-artifacts/port \
  /tmp/ts-printer-go-oracle /tmp/graphql-printer-prettier \
  > /tmp/ts-printer-tag-timing.json 2> /tmp/ts-printer-tag-timing.err
```

Exit 0, empty stderr, three sequential complete-process runs per side, every sample byte-identical
to Go. On **145,440 texts**, median texts/s: native **34,428**, Node **62,747**, Go **46,886**,
npm Prettier **3,089**. Native is 0.73x Go and 0.55x Node on this workload, and 11.15x npm Prettier.
This is a workload result, not a causal performance comparison with earlier corpora or machines.
Raw samples and source SHA are in `results/tag-timing.json`.

## Contextual await and yield increment

The generator adds 404 async/generator compositions and edge cases: nine operands, six await body
layouts, five yield layouts, four outer contexts, plus bare/delegated assignments and nested async
arrows. Context remains the parser's actual async/generator context. A preliminary corpus caught
60 extra yield-argument parentheses, and a separate valid bare-yield assignment probe caught a
missing-child panic. Both were fixed in scratch and their inputs retained before native validation.
The typed-await/cast input replaces the previous bare-await gap; there are still 13 proving inputs.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-await-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-await-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-await-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/(await|yield)' > /tmp/ts-printer-await-mutants.log 2>&1
```

Both exit 0. Full gate **444.419 s: 145,873 / 145,873** maximal fragments from 199 files, no full-file
parse failure. Strict Go, npm Prettier and embedded-fork equality holds on source Node, native
ASan/UBSan, JavaScript backend and native release. The separate leak pass and all 13 exact NotYet
proofs pass. Coverage is in `results/await-coverage.json`. Source cohere: 276 rules, ten checked,
100% Adamic-ready. Package vet and whitespace checks pass; no full-repository gate was run.

Both mutants compile and finish normally with exit 0 and empty stderr on Node and native, then
mismatch only in output. Await becomes `void`: first mismatch line 114228, **286.01 s**.
Yield delegation disappears: first mismatch line 114260, **284.74 s**. Complete mutant command
**291.807 s**, exit 0. No internal compiler or parser file changed.

## Variable statements and expression-statement composition

This family adds 646 variable-declaration cases and 40 retained statement-composition regressions.
An independent statement preview found receiver indentation, conditional receiver parentheses and
five two-segment assignment chains with the wrong wrapping. They were corrected before this gate.
A separate hashbang probe found header loss on both Node and native; the port now refuses it
rather than returning misleading output. The typed tagged-template gap now has a proving input.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-variables-composed-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-variables-composed-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-variables-composed-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/(variable|hashbang|assignment_chain)' > /tmp/ts-printer-variables-composed-mutants.log 2>&1
```

The mutant command exits 0 in **353.987 s**. All three mutants compile and finish normally,
exit 0 and empty stderr, on both Node and native before the byte comparison catches them:
variable keyword removal at line 114276 (**344.78 s**), hashbang refusal removal on proving-input
line 14 (**318.34 s**), and assignment-chain statement boundary removal at line 114296
(**348.21 s**). Hashbang uses the gap corpus deliberately: the accepted corpus excludes headers.

Source cohere prints 276 rules, ten checked, 100% Adamic-ready. Package vet and whitespace checks
pass. The full native oracle outcome follows below; no full-repository gate was run.

Full gate exits 0 in **483.463 s: 146,607 / 146,607**, from 199 files with zero full-file parse
failures. Source Node, native ASan/UBSan, JavaScript backend, native release and the separate leak
pass agree with Go cohere, npm Prettier 3.9.6 and the embedded fork. All **15** exact gap reasons
pass. The running gate printed the old hardcoded count of 13; only that log statement was changed
after its test binary was built, and the actual assertions already checked every proving input.
Coverage is in `results/variable-coverage.json`. No compiler or parser source changed.

## Complete-file statement composition

The new `statementsMain.ts` driver formats supported complete files. Program statement sequences,
directive prologues, empty statements/files and named untyped function declarations compose with
the expression printer. Pure precedence/parentheses logic moved to `parentheses.ts`; no callback
or compiler change was needed. Source cohere checks all 13 port files: 276 rules, 100% Adamic-ready.
Package vet and whitespace checks pass.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_STATEMENT_KEEP=/tmp/ts-printer-program-corpus \
ADAMIC_TS_STATEMENT_ARTIFACTS=/tmp/ts-printer-program-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestStatementsAgainstGoAndPrettier$' > /tmp/ts-printer-program-test.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRINTER_KEEP=/tmp/ts-printer-program-expression-corpus \
ADAMIC_TS_PRINTER_ARTIFACTS=/tmp/ts-printer-program-expression-artifacts \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestExpressionsAgainstGoAndPrettier$' > /tmp/ts-printer-program-expression-test.log 2>&1
```

The statement gate exits 0 in **401.238 s: 44,091 / 44,091** maximal statement/program fragments,
from 202 files including all 77 compiler sources, zero parse failures. Five complete files are
supported: `internal/load/testdata/0.1/compile/01_hello.ts`, three small tsprinter gap inputs
(`defaultSort.ts`, `numberConstructor.ts`, `prefixUpdateValue.ts`) and the parser's
`6_conditional_empty_array.ts` input. These are formatting inputs, not claims about their execution
or about complete compiler-file support. Source Node, native ASan/UBSan, JavaScript backend,
release and the separate leak check agree with Go, npm Prettier and the embedded fork. Five exact
statement refusal inputs agree on all three executions. Coverage: `results/statement-coverage.json`.

### Statement mutations and upstream evidence

The first combined mutation command selected program-separator, function-keyword and
optional-boundary mutations. It **exited 1 in 366.078 s** because the separator anchor matched
both the program and block paths. That harness failure is not a catch. The optional-boundary
mutation still completed normally and was caught on Node and native at line 115458:
``tag`a${x?.y.z}b` `` rather than ``tag`a${(x?.y).z}b` ``, **357.46 s**.
The generic function mutation completed normally and was caught at line 43831 (**337.87 s**),
but its first mismatch was an older function-expression case. It was therefore narrowed to
FunctionDeclaration before rerunning; the final declaration-specific result is recorded below.

```sh
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/program' > /tmp/ts-printer-program-separator-mutant.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestMutants$/declaration' > /tmp/ts-printer-program-declaration-mutant.log 2>&1
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
go test -v -count=1 -timeout 30m ./stage1/cohere/tsprinter \
  -run '^TestStatementUpstreamDifferences$' > /tmp/ts-printer-statement-upstream-test.log 2>&1
```

The external upstream-difference pin check exits 0 in **9.727 s**: five exact Go/fork pairs and
five exact npm outputs. It covers if/else, while, for, try/catch/finally and switch, all currently
unported. Accepted expression and statement corpora remain strict, with zero exceptions.
No full-repository gate was run, and no internal, parser, scanner or submodule source changed.

### Pushed expression-family commits in this unit

```text
fbed1b7 Port sequence expression layout
3f7e19b Port assignment expression layout and chains
5305f40 Port conditional expression layout
e191d3d Port object expression values and wrapping
f2e5e8a Port expanded call argument layout
80bca68 Port member call chains and curried calls
6ccd848 Port interpolated and multiline template layout
4990187 Port arrow signatures and basic statement bodies
d89ad60 Port optional chain stopping boundaries
9d8c9b6 Port named function expressions
a22bcf6 Port object methods and preserve template optional boundaries
43f35c4 Port ordinary tagged template expressions
384a593 Port contextual await and yield expressions
3dd85f2 Port variable statements and preserve composition boundaries
```

The corrected program-separator mutant exits 0 in **349.707 s**, subtest **338.12 s**. Both Node
and native finish normally, exit 0 and empty stderr, then disagree at line 3461: the formatted
`defaultSort.ts` program loses the newlines between its const declaration, sort and console call.
This specifically exercises the new complete-file program path.

An environment status update occurred during the final gates. The managed runtime reported current
observations and an enforced restricted network policy; the workspace and running test processes
were present in the process table, but a subsequent status check showed the two unfinished tests
were zombies. Their partial logs are not green results. The expression regression and
declaration-specific mutant are rerun below with explicit logged exit codes.

The declaration-only retry exits 0 in **259.245 s**, subtest **250.64 s**. Both Node and native
finish normally, exit 0 and empty stderr, then disagree at line 44023 on
`function named(x){const y=x+1;return y;}f(x);`: the declaration keyword is missing. Function
expressions are deliberately unchanged by this mutant, so this catch proves the new family.
Log: `/tmp/ts-printer-program-declaration-retry-mutant.log`. The retry used the same environment,
package and `-run '^TestMutants$/declaration'` flags above, with an explicit final `exit=0` line.
The interrupted earlier declaration attempt is not counted.

The complete expression regression retry exits 0 in **426.260 s: 146,678 / 146,678** fragments
from 202 files, zero parse failures. All 15 proving-input refusals match on Node, native and the
JavaScript backend. Strict Go, npm Prettier and embedded-fork equality holds on source Node,
native ASan/UBSan, JavaScript backend and native release; the separate leak pass is clean.
Log: `/tmp/ts-printer-program-expression-retry-test.log`, including explicit `exit=0`.
The retry used the same environment and flags above; the interrupted original attempt is not
counted as a completed gate. Coverage: `results/statement-expression-coverage.json`.

## Statement-family push and isolated throughput

Statement-family commit **b308bab67b3a836596d5a74ec734cf744d9e38bf** was pushed successfully:
`3dd85f2..b308bab codex/stage1-ts-printer -> codex/stage1-ts-printer`.
The worktree was clean before measuring. All validation jobs had ended; the source was frozen
under `/tmp/ts-printer-program-snapshot`, with parser/scanner and Node-oracle symlinks.

```sh
source /workspace/adamic-tools/env.sh
python3 /tmp/ts-printer-program-snapshot/stage1/cohere/tsprinter/testdata/measure.py \
  /tmp/ts-printer-program-corpus /tmp/ts-printer-program-artifacts/port \
  /tmp/ts-printer-go-oracle /tmp/graphql-printer-prettier --entry statementsMain.ts \
  > /tmp/ts-printer-program-timing.json 2> /tmp/ts-printer-program-timing.err
```

Exit 0, empty stderr. Three sequential complete-process runs per side, including startup,
input/output and formatting; **every timed sample is byte-identical to Go**. On **44,091 texts**,
median texts/s: native **29,135**, Node **37,933**, Go **42,278**, npm Prettier **2,825**.
Native is **0.69x Go**, **0.77x Node** and **10.31x npm Prettier** on this workload. This is a
measurement of a different corpus, not an inference about a regression from earlier increments.
Raw samples, source SHA, source-file hashes and method are in `results/statement-timing.json`.
The machine still reports **nproc 5**, with a four-core CPU quota.

This unit is a partial port, with 15 green printer-family commits from `fbed1b7` through `b308bab`.
Binding/assignment patterns, anonymous functions, Jest tables and broader control flow remain
unported, alongside comments, source trivia and later typed declarations/types. Their current
proving inputs and the five exact upstream control-flow differences are in GAPS.md. No complete
TypeScript compiler source file is claimed formatted. No internal, parser, scanner or submodule
file was edited. The port-specific gates above passed; the full repository gate was not run.

## Area merge-seat repair: stage1-format/ts-printer-2

Base: origin/stage1-format/ts-printer at 27acfbe8, merged with origin/main
b6b1538b (merge f489b16). Only this printer package is edited; no internal files change.

Toolchain: bash cloud/setup.sh, output /tmp/ts-printer-2-setup.log.
Timing lines: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 127s, total 127s. nproc 5; CPU quota four cores. Environment:
source /workspace/adamic-tools/env.sh, Go 1.27.1, clang 20.1.8, Node 24.19.0.

The before-repair filtered gap gate exited 1 in 4.970s, reproducing both obsolete
Number refusal and obsolete class-interface native-panic expectation.
The unchanged class proof now produces typed NotYet at gap.ts:6:28, before native
emission. Its callback workaround renames the concrete method readValue so main's
conservative compatible-shape scan does not erase its origin. The successful
workaround is the mutant rejected by the diagnostic check; Node, native and backend
print 17, with Linux leak checks passing. Number's former gap program is a positive
three-backend regression, and protocol/CLI conversions use Number again.

Pinned gate command (output /tmp/ts-printer-2-full.log):

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-tsc-strictness/typescript \
ADAMIC_TS_PRETTIER=/tmp/graphql-printer-prettier \
go test -v -count=1 -parallel 3 -timeout 60m ./stage1/cohere/tsprinter
```

349 files, zero parse refusals; 155,050 accepted expression fragments and 49,201
statement/program fragments, including five complete files, pass strict Go cohere,
embedded Prettier and npm Prettier comparisons. Both Prettiers remain pinned to
3.9.6. The ten new Number-call cases account for the expression-count increase.
5,072 document cases agree with Go and npm Prettier on Node, native and backend.
Hexadecimal widths in the independent Go protocol exercise Number's conversion:
its parseFloat mutant finishes normally on Node and native, then fails at output
line 5. Existing printer mutants retain successful-run byte-mismatch requirements.

Additional checks: go vet ./stage1/cohere/tsprinter (empty log); cohere --no-fix
--no-cache stage1/cohere/tsprinter/*.ts (exit 0, 13 checked, 100% Adamic-ready);
git diff --check; gofmt. Filtered oracle:
go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestInput'
passed in 6.394s, including read arguments/files, write files and UTF-8 sweep.
Output is /tmp/ts-printer-2-oracle.log. The full repository gate was not run.

Fix 3 waits on internal/leakcheck, absent from main b6b1538b when re-fetched.
The devtools/stage1-leaks branch is not merged. Existing ASan sites remain unchanged
until the shared helper lands; no macOS validation is claimed. Every named skip
now cites #xq2ecw6 (setup --gate-inputs), which installs pins and removes those skips.

Final pinned package result: exit 0, PASS, 827.352s. All 29 printer mutants
(including the new Number/parseFloat mutation) finished normally and were caught
by output mismatches on both Node and native. The class-interface and optional-
boolean gap checks also rejected their successful-lowering workaround mutants.
The final fetch/merge reports Already up to date at main b6b1538b; internal/leakcheck
is still absent. Fix 3 remains waiting, as authorized by the merge seat.
