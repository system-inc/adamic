Built witnesses for all 18 selected ledger rows, all proven in both backends and release/sanitized native modes, including D155/D156 holes and original compound reads.
Commits: original head f5a2212c; sparse-support merge d62901cf from dfb82dab; completion is this branch-head commit.
Commands and outputs: complete witness package 31.793s, four uncached oracle fixtures 1.007s, compound explain checks pass, vet exit 0; original setup 598.141s, nproc 5.
Mutants: all 18 direct-read guards and both original compound-read guards erased independently; all 40 native builds/runs succeeded and were caught by named stderr/exit assertions.
Not covered: full AST payloads, whole-program native compiler build and full repository gate; blocked rows zero, remaining rows zero.

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
| D155, D156 | new Array<number>(2), numeric kind index, hole and compound bitwise read | proven after sparse-support merge |
| D157-D160 | array stacks, numeric offset index, array element | proven |
| D161 | object array, literal 0 | proven |
| D162 | readonly numeric superPath read used as an object-array index | proven |
| D163, D164 | readonly number array, superPathDepth | proven |
| D165 | object array, statementOffset | proven |
| D166 | constructor.body.statements array, statementOffset | proven |
| D178 | node.declarations array, literal 0 | proven |
| D179, D182, D183 | object array outerParameters, i | proven |
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
No hole guard or mutant was claimed at that initial checkpoint; the completion below supersedes these refusals.

D162 is an undefined-index diagnostic: its cause is the numeric superPath
read, not absence of the statementsIn element. Its witness keeps both reads,
requires the path result as number, and allows undefined from the subsequent
object-array observation. Exactly one guard and two lookups are required.
The D163/D164 minimal witnesses also require the path result as number; this
explicit sink prevents their Node observation from removing the stricter-option
diagnostic. The original downstream operations remain subtraction and addition.

## Historical half-site checkpoint

Eleven of 18 diagnostic rows assessed: nine proven, D155/D156 blocked by
hole-construction lowering, seven remaining (D166 and D178-D183). Group 2
passed all five rows, present and absent, in both backends and release/sanitized
native. Log: /tmp/stricter-indexed-c-group2.log.

All ten group-2 erase-check runs built and exited 0 with the following stdout.
D161 and D165 printed undefined; D162 printed present; D163 printed 0;
D164 printed 1. Every result lost the required named exit-70 stop, caught
by the exact output/exit assertion. Numeric missing results becoming zero
after erasure are an observation of the mutant, not correct JavaScript behavior.

```sh
ADAMIC_WITNESS_CLI=/tmp/stricter-indexed-c-adamic go test ./stage3/stricter-indexed-c -run 'TestLedgerWitnesses/D16[1-5]' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-group2.log 2>&1
```

## Historical third group

D166 and D178-D181 passed present and absent witnesses, exact named stderr,
IR location and CLI explain counts, backend Node and both native modes.
All ten erase-check mutant builds succeeded, then exited 0 with undefined on stdout. The exact named exit-70 assertion caught every one.
Log: /tmp/stricter-indexed-c-group3.log.

```sh
ADAMIC_WITNESS_CLI=/tmp/stricter-indexed-c-adamic go test ./stage3/stricter-indexed-c -run 'TestLedgerWitnesses/(D166|D178|D179|D180|D181)' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-group3.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-oracle.log 2>&1
go vet ./stage3/stricter-indexed-c > /tmp/stricter-indexed-c-vet.log 2>&1
```

The filtered oracle passed all four existing fixtures in 3.388s with zero
cache hits (native misses=10, Node misses=8). Vet exited 0 without diagnostics.
The full repository gate was not run under the worker-gate exception.

## Initial 16-site result before sparse support

All 18 rows are assessed. Proven: D157-D166 and D178-D183 (16).
Blocked: D155/D156, both arrays with holes created by new Array, refused
at main.ts:1:35 with `stage 0 can't lower new an Identifier yet`.
Remaining: zero. The blocked rows have Node observations and pinned refusals,
not runtime checks or successful native builds. Representation support remains
with the stricter-checks worker as instructed. No record, typed-array, string
or Map receiver occurs among this slice's traced reads.

The complete package was run without ADAMIC_WITNESS_CLI, exercising its
automatic build of the current CLI. It passed in 30.020s. Its 34 source-Node
observations cover both variants of each supported witness plus two hole
probes. The supported witnesses pass 32 backend-Node runs, 64 native runs
(release and ASan/UBSan), exact named absent stderr/exit assertions, actual IR
source locations and 32 CLI explain assertions. Each supported variant lists
its source read as checked indexed-presence, counts one, and reports trusted: 0.

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-c -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-final.log 2>&1
# ok github.com/system-inc/adamic/stage3/stricter-indexed-c 30.020s
```

Every mutant below erases only the single indexed panic from emitted C.
Both release and sanitized builds succeeded for each. All 32 executions
exited 0, so the exact assertion requiring the named stderr and exit 70
caught them. Neither a build failure nor a sanitizer report is counted as
a mutant kill here. These are 16 independent per-row mutants, including
duplicate originating-read shapes; the receiver kind is array throughout.

| Mutant | Release exit / stdout | Sanitized exit / stdout | Catcher |
|---|---|---|---|
| erase-D157-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D158-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D159-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D160-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D161-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D162-presence | 0 / `present\n` | 0 / `present\n` | pinned named stderr and exit-70 assertion |
| erase-D163-presence | 0 / `0\n` | 0 / `0\n` | pinned named stderr and exit-70 assertion |
| erase-D164-presence | 0 / `1\n` | 0 / `1\n` | pinned named stderr and exit-70 assertion |
| erase-D165-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D166-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D178-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D179-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D180-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D181-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D182-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |
| erase-D183-presence | 0 / `undefined\n` | 0 / `undefined\n` | pinned named stderr and exit-70 assertion |

All outcomes appear in /tmp/stricter-indexed-c-final.log. Group logs preserve
the earlier checkpoint runs. The source-only preliminary log is
/tmp/stricter-indexed-c-source-node.log; final source observations come from
the complete harness, including the corrected D162 origin trace.

Only this unit's manifest, Go harness and report changed. Compiler implementation,
protected files and cohere remain unchanged. Pushed checkpoints went only to
codex/stricter-indexed-c; no pull request was opened. No merge into main or an
area branch was performed. The requested baseline was retained.

Scope limit: some ledger diagnostics share a read through a local. The 16
proven rows represent 11 distinct originating receiver/index expressions,
not 16 distinct checks emitted into the whole adapted compiler. Object AST
payloads are minimized, stack element arrays use a number payload, and explicit
required-type sinks keep the minimal witnesses rejected by stricter indexed
checking even where Node observation would otherwise narrow away the error.
The whole-program compiler date and native emission of all original AST types
are not established by this witness unit. The full repository gate was not run.


## Completion after sparse support

The published codex/stricter-options-checks dfb82dab sparse-array implementation
was merged into this branch as d62901cf, with no hand-copied compiler changes.
D155/D156 now have present cases initializing slot 0, and absent cases leaving
that in-range slot a hole. Their runtime shape remains new Array<number>(2)
with the numeric kind identifier indexing it. Node prints 7 or undefined; both
compiled backends match the present result and stop on the hole with the
site-named message and exit 70. CLI explain lists the required site, counts
indexed-presence=1 and trusted=0 for each input.

holes_test.go additionally reproduces the original |= operations: flag 1 for
D155 (Substitution), flag 2 for D156 (EmitNotifications). It observes the element
directly before and after the compound operation. On Node, the hole observation
prints undefined, then JavaScript's bitwise coercion creates flag 1 or flag 2.
Both compiled backends print the initial undefined observation, then stop at
the compound read with the exact indexed-presence message and exit 70. The
present slot starts at zero and becomes the flag, matching Node. Both original
compound reads have independently checked IR locations and CLI explain results.

All 18 direct-read witnesses pass together. Source Node/backend Node runs: 36
of each; native runs: 72, including release and sanitizers. The compound cases
add four source-Node and four backend-Node observations and eight native runs.
All successful sanitizer runs are leak checked. There are 36 direct-read explain
assertions plus four compound-read explain assertions.

The four new mutant types all build and exit 0 in both native modes:

| Mutant | Release / sanitized stdout | Catcher |
|---|---|---|
| erase-D155-hole-presence | 0 | exact named indexed stderr and exit 70 |
| erase-D156-hole-presence | 0 | same |
| erase-D155-compound-presence | undefined, then 1 | same; initial observation retained |
| erase-D156-compound-presence | undefined, then 2 | same; initial observation retained |

The compound mutants reproduce Node's unguarded result exactly. The 16 older
per-row mutants retain the outcomes in the table above. This gives 20 distinct
per-row/operation artifact mutations and 40 successfully built executions,
with no mutant surviving. The receiver kind is array throughout this unit.

An initial present-case source-position assertion selected the initializing
write because it had identical text to the read. Spelling the initializer as
slot [0] makes the independently selected [kind] read unambiguous. An early
compound probe held the observation in a typed temporary, whose checker
narrowing triggered a different undefined-use check. Direct observation fixes
that probe while keeping the original compound operation. Neither failed
probe is counted as a successful proof.

```sh
source /workspace/adamic-tools/env.sh
go test -count=1 -v ./stage3/stricter-indexed-c > /tmp/stricter-indexed-c-18-final.log 2>&1
go test -count=1 -run '^TestHoleCompoundReads$' -v ./stage3/stricter-indexed-c > /tmp/stricter-indexed-c-compound-explain.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-c-holes-oracle.log 2>&1
go vet ./stage3/stricter-indexed-c > /tmp/stricter-indexed-c-holes-vet.log 2>&1
```

The complete suite passed in 31.793s. After adding explicit CLI assertions to
the compound probes, that changed test was rerun separately and passed.
The four uncached oracle fixtures passed in 1.007s: native hits=0/misses=10,
Node hits=0/misses=8. Vet and whitespace checks pass. Full proof output is saved
in evidence/completed-witnesses.log and evidence/compound-explain.log.

Final disposition: proven 18, blocked 0, remaining 0. All rows' receiver/index
shapes are assessed and supported. Eighteen rows represent twelve distinct
originating receiver/index expressions; they do not establish eighteen distinct
whole-program guards. Whole-program lowering and its delivery date remain
outside this witness unit. This completion changes only the manifest, additional
Go witness and report; sparse representation changes come from the authorized
upstream merge. Only codex/stricter-indexed-c is pushed for this completion.
