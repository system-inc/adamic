Built an isolated 18-row manifest and runtime harness; fourteen rows proven and two hole rows refused.
Commits: checkpoints 5716a4ce and ed1d9041; this third group continues codex/stricter-indexed-c.
Commands and outputs: groups passed in 13.214s, 7.637s and 7.847s; filtered oracle passed in 3.388s; vet passed.
Mutants: D157-D166 and D178-D181 erase-panic mutants caught in release and sanitized builds, 28 runs.
Not covered yet: D182/D183 remain pending; D155/D156 hole construction refuses.

## Scope and provenance

The start is origin/codex/stricter-options-checks at
390af985caf2a81edb239cdf5e3238add50a7fcd, overriding the generic origin/main
start as requested. The ledger is origin/codex/stage3-checker-ledger at
a1a16427a46149435e25f847c45ad136f1f1e55c, specifically rows.csv.
The manifest retains each row's ID, file, diagnostic line/column, expression,
cause and evidence. The three selected files give exactly 18 rows.

Receiver tracing uses the pinned TypeScript v6.0.3 source at
050880ce59e30b356b686bd3144efe24f875ebc8, fetched to
/tmp/stricter-indexed-c-typescript. Only minimal witnesses are authored here;
no compiler or cohere source is copied or changed. The existing witness pattern
is stage3/stricter-options/indexed_test.go, read together with its REPORT.md.

D162/D163/D164 trace through superStatementIndex to superPath[superPathDepth].
D165/D166 trace through statement to statementsIn[statementOffset] and
constructor.body.statements[statementOffset]. D179/D182/D183 trace through
outerParameter to outerParameters[i]. D180/D181 trace through originalParameter
to node.parameters[i]. Those diagnostics share originating reads; each row has
an isolated witness to keep its outcome individually observable. This report
counts diagnostic rows, not 18 distinct originating reads in the whole compiler.

The conservative minimalization retains array versus readonly array, nested
array versus object element, property-chain receivers, literal zero versus
numeric identifier indexes, and the downstream numeric arithmetic distinction.
Full TypeScript AST interfaces are reduced to fixed object payloads; the
witnesses prove presence representation and index shape, not all AST payload
representations or a whole-program compiler build.

## Harness contract

sites.json holds TypeScript interoperability source templates. The harness
materializes each as main.ts under its own tsconfig with strict checking and
noUncheckedIndexedAccess omitted, exactly the project-option setting whose
runtime conversion is under test. These are .ts probes explicitly requested
by this unit, not new Adamic .a programs. Source Node runs the same materialized
file using Node's type stripping, with no handwritten oracle translation.

Present input must match Node's stdout, empty stderr and exit 0. Absent input
must print undefined on source Node and stop in both backends with exactly
`adamic: panic: indexed read is absent: <witness file>:<line>:<column>\n`,
empty stdout and exit 70. Location is calculated from the source read and checked
against the actual IR guard. The CLI --explain-checks must report exactly one
indexed-presence guard and trusted: 0. Native builds run in release and under
ASan/UBSan, with their normal leak checks on successful termination.

Each absent supported witness erases just its one emitted C panic. The mutant
must build with the same native options, then differ from the pinned output and
exit assertion. Compiler source remains untouched. Failed mutant compilation
cannot count as a caught mutant.

## Progress

| Rows | Traced receiver/read | Status |
|---|---|---|
| D155, D156 | new Array<number>(2), numeric kind index, hole | blocked: stage 0 can't lower new an Identifier yet |
| D157-D160 | array stacks, numeric offset index, array element | proven |
| D161 | object array, literal 0 | proven |
| D162 | readonly numeric superPath read used as an object-array index | proven |
| D163, D164 | readonly number array, superPathDepth | proven |
| D165 | object array, statementOffset | proven |
| D166 | constructor.body.statements array, statementOffset | proven |
| D178 | node.declarations array, literal 0 | proven |
| D179, D182, D183 | object array outerParameters, i | D179 proven; D182/D183 pending |
| D180, D181 | node.parameters object array, i | proven |

## Commands

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-c-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go build -o /tmp/stricter-indexed-c-adamic ./cmd/adamic > /tmp/stricter-indexed-c-build.log 2>&1
ADAMIC_WITNESS_CLI=/tmp/stricter-indexed-c-adamic go test ./stage3/stricter-indexed-c -run 'TestLedgerWitnesses/D15|TestLedgerWitnesses/D160' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-group1-final.log 2>&1
```

nproc reports 5. Setup completed Go readiness at 0.336s, Node at 0.418s,
clang at 0.949s, markdown dependencies at 2.240s and submodules at 313.836s.
The pinned submodule fetch accounts for the startup delay. Go build readiness: 598.021s; cache warm: 598.113s; done: 598.141s.
Go 1.27.1, clang 20.1.8, Node 24.19.0; cpu.max=400000 100000.

The initial group run failed two harness expectations, not compiler behavior:
the hole refusal wording and the explain per-site line. Both were corrected
to observed output; the final run above passed. The initial log remains
/tmp/stricter-indexed-c-group1.log. D155/D156 refuse at witness main.ts:1:35.
No hole guard or mutant is claimed.

D162 is an undefined-index diagnostic: its cause is the numeric superPath
read, not absence of the statementsIn element. Its witness keeps both reads,
requires the path result as number, and allows undefined from the subsequent
object-array observation. Exactly one guard and two lookups are required.
The D163/D164 minimal witnesses also require the path result as number; this
explicit sink prevents their Node observation from removing the stricter-option
diagnostic. The original downstream operations remain subtraction and addition.

## Half-site checkpoint

Eleven of 18 diagnostic rows assessed: nine proven, D155/D156 blocked by
hole-construction lowering, seven remaining (D166 and D178-D183). Group 2
passed all five rows, present and absent, in both backends and release/sanitized
native. Log: /tmp/stricter-indexed-c-group2.log.

All ten group-2 erase-check runs built and exited 0 with empty stderr.
D161 and D165 printed undefined; D162 printed present; D163 printed 0;
D164 printed 1. Every result lost the required named exit-70 stop, caught
by the exact output/exit assertion. Numeric missing results becoming zero
after erasure are an observation of the mutant, not correct JavaScript behavior.

```sh
ADAMIC_WITNESS_CLI=/tmp/stricter-indexed-c-adamic go test ./stage3/stricter-indexed-c -run 'TestLedgerWitnesses/D16[1-5]' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-group2.log 2>&1
```

## Third group

D166 and D178-D181 passed present and absent witnesses, exact named stderr,
IR location and CLI explain counts, backend Node and both native modes.
All ten erase-check mutant builds succeeded, then exited 0 with empty stderr
and undefined on stdout. The exact named exit-70 assertion caught every one.
Log: /tmp/stricter-indexed-c-group3.log.

```sh
ADAMIC_WITNESS_CLI=/tmp/stricter-indexed-c-adamic go test ./stage3/stricter-indexed-c -run 'TestLedgerWitnesses/(D166|D178|D179|D180|D181)' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-group3.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-oracle.log 2>&1
go vet ./stage3/stricter-indexed-c > /tmp/stricter-indexed-c-vet.log 2>&1
```

The filtered oracle passed all four existing fixtures in 3.388s with zero
cache hits (native misses=10, Node misses=8). Vet exited 0 without diagnostics.
The full repository gate was not run under the worker-gate exception.
