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
