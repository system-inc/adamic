This scout inventories postfix non-null assertions in stock TypeScript 6.0.3 and
observes their operands while stock tsc runs on Node. The original measurement changed no adaptation or
production compiler source; the ruled follow-up adds adaptation 49 as described below. The source pin is
050880ce59e30b356b686bd3144efe24f875ebc8, tag v6.0.3. The topic starts from
origin/main 45487a809f89885a3fc651cd590e7dabf31362dc.

The compiler API finds 1,123 assertions: four flow-proven erasure candidates and
1,119 check candidates. sites.json and sites.tsv record every file:line:column,
AST span, expression, operand type and disposition. files.json records all 80
stock closure files, including files with zero assertions and generated diagnostics.
The ledger's 81st file is adaptation 47's hostErrors.ts, with no stock counterpart.
Its AST template has zero assertions and its hash matches the ledger. closure.json
preserves membership and hashes from 855bcfaa. Prefix logical negation and
exclamation tokens on declarations are counted separately and excluded from the
assertion census.

The four erasure candidates are:

| Stock location | Expression | Operand type |
| --- | --- | --- |
| src/compiler/builder.ts:1793:46 | state.program! | Program |
| src/compiler/checker.ts:15960:125 | memberName! | __String |
| src/compiler/checker.ts:43368:22 | getSymbolOfDeclaration(node)! | Symbol |
| src/compiler/transformers/declarations.ts:717:48 | oldDiag! | GetSymbolAccessibilityDiagnostic |

Types are queried on the operand before asserting it, using typescript@6.0.3.
Every union constituent must be non-nullish; any, unknown, void and unconstrained
type parameters remain checked. Constraints are inspected recursively, and one
non-nullish intersection constituent proves the intersection non-nullish. The
inherited compiler project has zero semantic diagnostics. Adding
noUncheckedIndexedAccess and exactOptionalPropertyTypes gives 2,140 diagnostics on
the unadapted source but exactly the same 4/1,119 split. Those stricter options drive
the inventory. No options are weakened to change the result.

These rows are API erasure candidates, not an observed Adamic build of all tsc.
The area may conservatively retain checks for nullable storage or invalidated
narrowing. Actual lowering is observed for the three reduced controls through
--explain-checks and both emitted backends.

Runtime evidence

The stock control and final instrumented CLI each match all 301 existing goldens.
Of 1,119 distinct check candidates, 322 execute and 797 remain unvisited. Seventy-two
sites see undefined, totaling 61,952,884 evaluations; none sees null. All 301
processes first observe undefined at scanner.ts:4097:24, the unconditional
Script_Extensions table initializer. Under the initial ruling, each process would
stop at that initializer before processing its project. The later placeholder ruling
below removes that consequence for literal initializers. The observer deliberately
continues, exposing later reset and optional-field cases too.

runtime-summary.json records the final 301-project experiment. NULLISH.tsv lists
every observed nullish site, expression, value, evaluation count and triggering
input set. input-sets.json expands those sets into exact case IDs;
runtime-projects.jsonl gives each case's upstream path, raw-input SHA256, first
nullish observation and complete site counts. runtime-sites.json includes every
checked site, including those never reached. Unvisited assertions are not proved
safe. The unchanged corpus is stage3/drivers/tsc's 300 upstream projects and its
three-file tiny project, with its existing materialization, options and goldens.

bundle.cjs bundles the stock tsc entry with esbuild@0.27.3. Compiler-API AST
transformations insert an identity observer at every check candidate; proven
candidates undergo ordinary TypeScript erasure. The helper records null/undefined
and returns the original operand, allowing Node to continue. Logs are separate
from CLI stdout/stderr. Inventory checks require every source insertion and every
emitted probe. IDs include both UTF-16 span endpoints because nested assertions
can share file:line:column. The instrumented bundle disables tree shaking, retaining
all 1,119 probes and some unused functions without calling them. The control keeps
normal bundler settings; its final bytes match the control run on all 301 projects.

An asserted compound-write reference cannot become a call on the assignment's
left side. The two |= sites and one ++ site hold receiver, old read and result
once in an AST-generated arrow, observe the held read, then perform the original
write. Postfix returns its old value. tests.cjs checks getter/setter calls, receiver
calls, RHS evaluation, output, nullish capture and nested IDs against stock Node.
Unreviewed asserted reference forms stop the transform.

Fixtures and the .a boundary

The inspected and built feature is origin/codex/non-null-checked-next at
93ebcf520c87eb55b47ada1d39189796959b2778, following area c41c0e06. Its
checkedAssertionSource and refusal tests deliberately reserve checked ! for .ts.
The three committed .a programs therefore have truthful refusal a-check headers.
Main and the feature area refuse them exactly. observe.cjs writes temporary .ts
copies in scratch to exercise the feature:

| Fixture | Real stock site | Node | Area native and JS |
| --- | --- | --- | --- |
| 01_erased.a | builder.ts:1793, pending program read | 7 | same; proven 1, checked 0 |
| 02_checked_pass.a | core.ts:519, getOrUpdate map read | 17 then 17 | same; proven 0, checked 1 |
| 03_checked_stop.a | core.ts:1896, memoize reset | 42 then after | 42; named undefined! panic, exit 70 |

The reductions preserve the asserted operations with typed scaffolding, not entire
compiler owners. Program is a scalar; getOrUpdate specializes K=string, V=number;
the memoize driver isolates its callback reset without consuming the invalidated
callback. The failing control's exact expression, location, stderr, stdout prefix
and exit are checked independently of Node's erasure. counts.md is refreshed for
these external fixtures; central registered-oracle counts are outside the territory.

Mutants run

mutants.cjs edits actual emitted C, then builds with clang, ASan and UBSan:

- 01 inserts a spurious terminal check, caught by its Node output/exit golden.
- 02 inverts the real presence guard, caught by its Node output/exit golden.
- 03 removes the actual panic and finishes with Node's 42/after output, caught by
  the expected named panic, exit 70 and before-only stdout contract.

All three compile valid C and fail semantically. A compiler error or sanitizer
failure does not count as a kill. mutants.json retains commands and observations.
Five measurement mutants also run: claim every operand proven, claim none proven,
evaluate a receiver twice, omit observers, and collapse nested span IDs. Seven
independent API cases, Node output/evaluation comparisons, hit counts and distinct
IDs catch them. a-check.cjs compiles three actual checked-header mutants; each
is caught by the fast gate's refusal-header policy. All original headers pass on
main. No full package or full gate is run as confirmation.

Reproduction

Use separate output folders and write test output directly to logs. Prepare an
external stock source checkout at the pin and generate diagnostics with upstream's
scripts/processDiagnosticMessages.mjs. Install typescript@6.0.3, esbuild@0.27.3,
@types/node@25.3.3 and @types/source-map-support@0.5.10 in external node_modules.
No upstream checkout is committed here.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/nonnull-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/path/to/stock-dependencies/node_modules
node stage3/scouts/step09/nonnull/tests.cjs > /tmp/nonnull-tests.log 2>&1
node stage3/scouts/step09/nonnull/census.cjs /tmp/stock > /tmp/nonnull-census.log 2>&1
node stage3/scouts/step09/nonnull/bundle.cjs /tmp/stock /tmp/bundles > /tmp/nonnull-bundle.log 2>&1
TSC_JOBS=1 TSC_RESULTS=/tmp/control python3 stage3/drivers/tsc/driver.py -- node /tmp/bundles/control.cjs > /tmp/nonnull-control.log 2>&1
mkdir -p /tmp/probes
ADAMIC_NONNULL_LOG=/tmp/probes TSC_JOBS=1 TSC_RESULTS=/tmp/instrumented python3 stage3/drivers/tsc/driver.py -- node /tmp/bundles/instrumented.cjs > /tmp/nonnull-instrumented.log 2>&1
node stage3/scouts/step09/nonnull/collect.cjs /tmp/control /tmp/instrumented /tmp/probes > /tmp/nonnull-collect.log 2>&1
node stage3/scouts/step09/nonnull/observe.cjs /tmp/area-adamic /tmp/fixtures > /tmp/nonnull-fixtures.log 2>&1
node stage3/scouts/step09/nonnull/mutants.cjs /tmp/fixtures /path/to/area-sanitized-runtime-cache > /tmp/nonnull-mutants.log 2>&1
node stage3/scouts/step09/nonnull/a-check.cjs /tmp/main-adamic /tmp/header-mutants > /tmp/nonnull-a-check.log 2>&1
```

Build area-adamic in an isolated checkout of the feature revision, with main's
cohere pin. A shared submodule symlink caused initial Go VCS stamping to fail:
“error obtaining VCS status: exit status 128; Use -buildvcs=false to disable VCS
stamping.” go build -buildvcs=false succeeded; observations.json pins the feature
SHA explicitly. No compiler-area code is merged into this topic.

Setup timings: Node 0.035s, Go 0.048s, submodules 0.107s, clang 0.268s, markdown
1.317s, build cache 63.715s, total 63.760s. nproc=5; CPU quota is four. All setup
timing lines are in evidence/setup.log. The control's 301 projects passed in
227.667s; final instrumentation timing is in runtime-summary.json.

Limits: this is the specified acceptance selection, not every upstream test,
watch/build/incremental mode, every generic instantiation or native tsc. The .a
refusal boundary prevents claiming the authored .a fixtures execute checked !;
their temporary .ts controls do. No adaptation of intentional undefined resets is
proposed by this unit. Stock snippets and reductions derive from Microsoft
TypeScript, copyright Microsoft Corporation, licensed under Apache-2.0.

The cloud environment restarted during the final span-ID run after 285 projects.
resume.py used the unchanged driver for the remaining 15 cases and tiny, then
recompared every byte of all 301 projects against the original goldens. The report
records that interruption and the resumed 16-project receipt rather than inventing
a continuous-run elapsed time. probe-contract.cjs independently requires exactly
one undefined observation of the unconditional Script_Extensions initializer in
each process; an actual probe-log omission mutant exercises that check.


Ruled follow-up: placeholders and real type lies

The October 8, 12:35 ruling makes literal `undefined!` and `null!` placeholder
initializers, including resets. Other assertions remain checked unwraps. The 72
observed sites split into 50 literal placeholders, seven nullish property/call
results, and 15 other nonliteral nullish operands. All 22 nonliteral operands
received `undefined`, never `null`; they are the ruling's category (3) type lies.
[SPLIT.tsv](SPLIT.tsv) lists all 72 stock locations and input sets. split.cjs uses
the stock compiler API and the previously measured span IDs, not source regexes.
split.json records the expression and its owning slot; placeholder-slots.json
resolves all 50 slots and their lexical binding references.

The slot experiment matches stdout, stderr and exits for all 301 projects. It
records 61,883,896 placeholder executions and 4,888,904 first reads after those
executions. Of 50 slots, 25 have a write before every observed first read, 19 have
at least one first read before a write, and six have no observed first read.
There are 344,015 reads before writes. None of the tracked arrays took an opaque
native-array operation in this run. These observations include presence tests and
reads returning undefined; they do not mean each such read is itself an asserted
unwrap or an error under the placeholder ruling.

[SLOTS.tsv](SLOTS.tsv) gives the answer for each file:line, including unread epoch
counts and every input observing a read before a write. slot-observations.json
also gives first-read locations and whether their values were nullish;
slot-projects.jsonl retains per-input slot counts. Unread epochs are coverage
limits, including under the 25 positive rows. The six entirely unread slots are
not proved safe. This is evidence on the 301 projects, not on every possible tsc
input.

slots-transform.cjs uses stock compiler API transforms. Lexical tokens distinguish
closures and per-call variable storage; WeakMaps identify aliased object slots.
A reset rearms the first-read observation. Writes stay in place and no metadata
is attached to compiler objects. Optional chains and destructuring are lowered by
stock tsc to ES5 before field-access instrumentation. The final build uses external
`tslib` helpers to avoid capture by tsc's imported `Symbol` type. The initial local
helper version passed 296 projects and failed five for that harness defect; those
observations were discarded. The corrected complete rerun passed 301/301 in
403.574 seconds. Reflective descriptor reads, spread, Object.assign, JSON and
inherited property reads are observed; unreviewed opaque native-array operations
are reported as unknown rather than safe.

slots-tests.cjs checks 18 helper observations and kills four mutants: omitted
first-read observation, omitted write observation, inverted write status, and a
getter writing the slot before the observer latches the read status. Every mutant
is caught by the first-read observation assertions.

The requested adaptation and the TypeScript bug log are in
[49-honest-optionals](../../../adapt/49-honest-optionals/README.md). Eleven of the
22 lies admit byte-preserving private type repairs. Eleven are blocked by the
simultaneous public API and JavaScript byte requirements; each is listed with its
own input and reason. The classifier does not quietly move those sites into the
placeholder category.
