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
