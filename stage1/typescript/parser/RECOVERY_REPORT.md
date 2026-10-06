Fixed the four parser sources underlying all eight reported lint failures.
Go, Node and sanitized native agree on their recovered trees and diagnostics.
Three recovery mutants prove diagnostic, child-tree and timeout comparisons.
All 77 compiler files yield 22,497 planned cutoff/removal/duplication inputs.
General recovery is unfinished: that comparison fails on a saved bare export.

## Built

Branch `codex/parser-recovery`, based on `origin/codex/typescript-scanner`
at `0090256e607c3f2de7d5b67cef680ec010f95c1d`, as the unit explicitly
requested. The checkout's narrow fetch omitted that branch; fetching its
explicit ref succeeded. No compiler, runtime, scanner, lint or submodule
source was changed. No pull request was opened.

The four source shapes are recorded in lint's `VOLUME.md` and `GAPS.md`,
rather than its `REPORT.md` on this branch:

```ts
interface I
interface I { m(a: string): void;
interface I { m<(a: string): void; }
interface I { m<T(a: T): T; }
```

Expected tokens now report diagnostic 1005 without consuming an unexpected
token. Type-member lists terminate at EOF and a missing opening brace creates
an empty member list. Generic parameter lists recognize Go's recovery
delimiters, including an opening parenthesis. Parameter lists also terminate
at EOF and retain an empty missing list when their opener is absent.
Parser speculation rolls diagnostics back with the scanner and node table.

`main.ts --whole --recovery` prints diagnostics and the recovered canonical
tree. The ordinary driver rejects diagnosed input instead of silently printing
an answer without its diagnostics. Diagnostic fields are code, byte start,
byte length, category and escaped English message. UTF-16 positions are
converted through the same byte-offset table used for trees. Tree fields
remain exactly the canonical protocol documented in `WHOLE_REPORT.md`.
This is not serialization of every Go AST field.

The independent Go oracle calls unmodified typescript-go. Recovery mode
allows its parse diagnostics and prints their actual localized messages.
It does not normalize their codes or positions. Twelve parser cases cover
the four originals, Unicode/CRLF prefixes and trailing CRLF/whitespace.
A separate integration test exercises all eight original method/property
lint combinations against real cohere findings and proposed repairs.
Cohere's edit engine refuses malformed source, so this is findings/proposed
repairs parity, not fixed-source parity.

## Wider comparison and remaining failure

`TestIncompleteCompilerAgrees` checks the TypeScript v6.0.3 pin
`050880ce59e30b356b686bd3144efe24f875ebc8` and enumerates every compiler file
before comparing. The real Go parser identifies literal locations for
regex/template rescans in the independent scanner, giving token boundaries
without slash or brace heuristics. Each file is cut after every Nth token,
with N selected for up to 256 cuts. Short files use N=1. An empty cutoff is
also included. Up to 32 regularly spaced tokens per file are removed and
then duplicated with a separating space. No diagnostic input is filtered.

Observed: all 77 files were enumerated, giving **22,497 planned inputs**.
Only **two inputs were compared**, because the test stops at its first
failure after running both ports. The empty cutoff passes. A namespace
wrapper cut after `export` fails on both Node and native with exit 70.
Go completes with diagnostic 1128, `Declaration or statement expected.`,
and a SourceFile containing only its EOF token. The exact input and all
three outputs are retained in `validation/recovery-export.*`.

Each incomplete-input process has an independent two-second deadline.
Input, stdout and stderr are written before reporting a failure and survive
test cleanup under `/tmp/adamic-parser-incomplete`, or the explicit
`ADAMIC_RECOVERY_ARTIFACTS` directory. Token-boundary enumeration is bounded
as well. The EOF-loop mutant proves a nonterminating case fails by this
deadline with its input saved.

General parity was **not achieved**. The next missing mechanism is Go's
outer-list element recognition and resynchronization: a bare export is not
a statement element, so Go reports and skips it; the port consumes it as a
modifier and attempts to parse an EOF expression. That mechanism is not
implemented by expected-token insertion or the signature list terminators.
Other malformed grammar remains unverified. Scanner-error integration,
missing identifier/expression/type nodes, related diagnostic information,
and recovery context flags are also outside this implementation.
The remaining 22,495 planned inputs, including the token edits, have not
been compared. The broader test deliberately remains failing when the
pinned corpus is supplied; its failure is not converted into a skip.

Lint's old `checkRecoveryRefusal` assertions still describe the earlier
parser behavior. They need migration to positive parity checks by the lint
unit. This unit leaves that worker's files untouched and provides the
positive eight-case integration test under the parser package instead.
The complete repository gate is not claimed green.

## Mutants

| Mutant | Independent check that catches it on both backends |
| --- | --- |
| Expected-token diagnostic 1005 changed to 1006 | Diagnostic bytes differ while the tree remains intact |
| Generic TypeParameter child list emptied | Recovered tree loses the Identifier child for T |
| Type-member EOF terminator removed | Each run reaches the two-second deadline |

The first two compile and finish normally on Node and ASan/UBSan native.
The third compiles and is caught specifically by the external deadline,
not by a compiler refusal or sanitizer crash. Successful executions retain
leak checking through the normal sanitized native build.
The first version of the child mutant used an untyped empty array and met
stage 0's `array of never` refusal. That attempt is **not credited**.
The final mutant uses `children.slice(0, 0)` and completes normally.

## Commands and observations

Every test command wrote its output to a log, never a pipe. Commands ran
from the repository after sourcing `/workspace/adamic-tools/env.sh`.
`nproc` printed **5**. `bash cloud/setup.sh` succeeded and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (21s)
setup: done in 21s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, Node 24.19.0, clang 20.1.8. The actual setup log is retained.

```sh
bash cloud/setup.sh > /tmp/parser-recovery-setup.log 2>&1
source /workspace/adamic-tools/env.sh

go test ./stage1/typescript/parser -run '^TestMethodRecoveryAgrees$' \
  -count=1 -v > /tmp/parser-recovery-baseline.log 2>&1
# Before the fix: FAIL, 18.563s. Three source shapes panic on both backends;
# the missing closing brace reaches both two-second deadlines.

go test ./stage1/typescript/parser \
  -run '^(TestMethodRecoveryAgrees|TestRecoveryMutants|TestRecoveredLintCasesAgree)$' \
  -count=1 -v -timeout 10m > /tmp/parser-recovery-focused-final.log 2>&1
# PASS, 70.908s. Twelve parser inputs, all eight lint combinations,
# and all three mutants on both backends.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -skip '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 30m > /tmp/parser-recovery-regression.log 2>&1
# PASS, 361.047s. All existing parser tests and recovery tests passed.
# 77 whole compiler files: 44,766,682 identical whole-tree bytes.
# The broader incomplete-input test is explicitly excluded here.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout 30m > /tmp/parser-recovery-incomplete-final.log 2>&1
# FAIL, 15.026s. Planned 77 files / 22,497 inputs; stopped after two inputs.

(cd cohere && go build -o /workspace/scratch/cohere ./command/cohere) \
  > /tmp/parser-recovery-cohere-build.log 2>&1
/workspace/scratch/cohere --format-only stage1/typescript/parser/*.ts \
  > /tmp/parser-recovery-format.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts \
  > /tmp/parser-recovery-cohere.log 2>&1
# Exit 0, 276 rules, no findings, 100% Adamic-ready.

go vet ./... > /tmp/parser-recovery-vet.log 2>&1
# Exit 0, no output.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' \
  -count=1 -timeout 10m > /tmp/parser-recovery-filtered-oracle.log 2>&1
# PASS, 3.225s.
gofmt -l stage1/typescript/parser
# No output.
git diff --check
# No output.
```

The first lint integration build succeeded but emitted Go dependency-download
chatter on stderr. The existing execute helper treats any stderr as failure;
rerunning after the successful downloads passed all eight combinations in
20.103s. It was a harness/build-output issue, not a lint comparison failure.
The final focused run repeats the checks on formatted source, including the
added Unicode and trailing-trivia variants. Full repository tests and lint's
full package were not run. Parser regressions, the focused integration, vet,
cohere source lint and the named filtered compiler oracle were run.

Implementation commit: `54024b1` (full SHA recorded by `git rev-parse` in
`validation/recovery-commits.log`). The subsequent documentation commit
retains this report, raw logs and the first broader failure. The requested
branch is pushed without rewriting history.
