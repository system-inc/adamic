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
