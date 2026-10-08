Built 22 minimal project .ts runtime witnesses proving 23 of 27 checker ledger rows; pinned four blocked rows.
Commits: b7214299, 59257d6a, 90342391, 0ebf5687 and a17a300f, all pushed only to codex/stricter-indexed-a.
Commands: setup passed in 1009.044s (nproc=5); complete witness package passed in 129.065s; filtered oracle passed in 8.125s; vet passed.
Mutants: all 22 single-panic erasures compiled and were caught by exact indexed-site stdout/stderr/exit assertions, including every supported array payload kind.
Not covered: D037 and D071-D073, the whole-program compiler build, hole/typed-array/record representations, and D069's initially absent optional-field context.

Group 1 proves D054, D056, D057, D067 and D068. Each source Node run has exit 0,
empty stderr, and stdout `7` when present or `undefined` when absent. Release
native, sanitized native and backend JavaScript agree with Node when present.
When absent they instead have empty stdout, exact named stderr
`adamic: panic: indexed read is absent: file:line:column` and exit 70. The IR and
actual CLI `--explain-checks` list exactly that read as checked, count
indexed-presence=1, and report trusted: 0.

Each mutant keeps the receiver and index lookup and erases only its emitted-C
panic. It must successfully compile with sanitizers. All five mutants exit 0;
D067 prints `0`, and the other four print `undefined`. The exact observation
assertion catches them. No build-warning or sanitizer failure is counted as a kill.

The manifests are input descriptions, not new Adamic programs. The harness writes
minimal .ts programs and their owning tsconfig in scratch directories, preserving
the receiver kind and index form while reducing TypeScript compiler payloads to
small object, callable or numeric values. This follows the base report's generated
.ts witness pattern and keeps the project's noUncheckedIndexedAccess=false. Source
Node runs the same .ts directly; no handwritten erased-source oracle is used.
Out-of-range is the absent array case. No compiler implementation file is edited.

Commands, all output written to logs before reading:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-a-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-a -run 'TestCheckerIndexedWitnesses/(D054|D056|D057|D067|D068)$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-a-group1-verified.log 2>&1
```

Setup: Node ready 0.125s, Go 0.156s, clang 0.955s, markdown dependencies 1.827s,
submodules 234.251s, Go build 1008.454s, cache warm 1008.701s, done 1009.044s.
Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc=5, CPU quota 4.
Environment: /workspace/adamic-tools/env.sh. Checked-in logs retain the observations.

Group 1 pushed as b7214299. Group 2 also proves D060/D061's shared stack[0]
read, D069's returned-array read and D077's declarations[0] read. Its gate passes
in 24.944s, including the refusal probes for D037 and D071 through D073.
The D037 probe retains an ordinary optional-return TS2322 error, before lowering.
The previous adaptation inventory independently calls checker.ts:12034:30 "Not
an indexed read: getSourceFileOfNode has an optional return". This is an attribution
question, not a proven indexed guard. D071, D072 and D073 refuse array destructuring
with `stage 0 can't lower destructuring anything but a tuple into [names] yet`.
The source Node present and absent observations still pass; CLI refuses and never
counts these witnesses as checked.

D069 keeps the optional receiving field already present. Its exact enclosing
assignment expression first refused `a BinaryExpression with a value and a value`.
A statement assignment with an initially absent field then compiled but its valid
native run stopped with `compiler bug: a field the checker proved is there is missing`.
That separate optional-field bug has its input in gaps/optional-field.json and
its observations in logs/optional-field-gap.txt, and
is not fixed here. Initializing the receiving field makes the indexed-read witness
pass without changing its receiver or constant-zero index. This proves the array
check, not support for initially absent optional-field writes or the exact full
checker expression. Its erased check mutant exits 70 with a different panic,
which fails the exact indexed-site stderr assertion.

Group 2 pushed as 59257d6a. Group 3 adds independently observable parameter-use
variants D082 through D086; all five pass in 24.516s. Each variant has exactly one
node.parameters[0] guard. These ledger rows diagnose later uses of the same
original parameter read; they are not five different original indexed expressions.

Group 3 pushed as 90342391. Group 4 proves D087 through D091 in 31.004s.
All five erased-panic mutants run successfully and print undefined, losing the
required named stop. The full observations are in logs/group4.txt.

Group 4 pushed as 0ebf5687. Group 5 proves D092 through D095 and pins the three
array-destructuring refusals. Its gate and exact 27-row completeness test pass
in 25.886s. Every assigned row now has a manifest: 23 rows proven by 22 runtime
programs, four blocked rows, zero unexamined rows. D060/D061 share one read.
Thirteen parameter variants cover thirteen downstream diagnostic rows on one
original read. These counts are witness coverage, not whole-program guard counts.


Final result and limits

The exact assignment is the 27 noUncheckedIndexedAccess rows in checker.ts from
`a1a16427:stage3/ledger/checker-259/rows.csv`, retained in ledger.json. sites.csv
maps every original row and source location to its minimal read, receiver,
exact expected witness line/column and disposition. TestAssignedLedgerIsCovered
rejects missing, duplicate, extra or wrongly scoped rows. Remaining unexamined
rows: 0. Remaining rows without runtime proof: 4.

| Rows | Underlying read | Result |
|---|---|---|
| D054 | elementTypes[unionIndex] | proven |
| D056 | types[unionIndex] | proven |
| D057 | targets[i] | proven, callable payload |
| D060, D061 | stack[0] | one shared read proven, variadic tuple payload |
| D067 | tupleType.elementFlags[index], index=pos-paramCount | proven, numeric payload |
| D068 | typeParameters[typeArgumentPosition] | proven through a local |
| D069 | getTypeArguments(type)[0] | array read proven with receiving field already present |
| D077 | copy.declarations[0] | proven |
| D082-D094 | node.parameters[0] | 13 independently observable use variants proven |
| D095 | mergedSymbols[symbol.mergeId] | proven, property index |
| D037, checker.ts:12034:30 | file from getSourceFileOfNode | blocked: ordinary optional-return TS2322 in the probe; the earlier inventory disputes indexed attribution |
| D071, checker.ts:46114:128 | yieldType from three-name array destructuring | blocked: array destructuring refused |
| D072, checker.ts:46114:208 | returnType from three-name array destructuring | blocked: array destructuring refused |
| D073, checker.ts:46128:128 | yieldType from one-name array destructuring | blocked: array destructuring refused |

All assigned actual indexed receivers are arrays. No assigned read calls Map.get
or indexes a string, record or typed array. Out-of-range is the runtime absence
used here; hole representation is not tested or implemented. No work belonging
to the representation worker 01a118d0 is built. No TypeScript compiler source is
rewritten, no forbidden compiler implementation file is edited, and no code is
copied from cohere. This is minimal shape evidence, not proof that the full checker
functions or the whole compiler lower. In particular D060/D061 reduce the
surrounding reportError spread to observing the selected tuple; D057 observes
the callable before invoking it; D095 observes the raw read before its original
truthiness fallback. These preserve receiver kind and index form, rather than
claiming full surrounding control-flow coverage.

The parameter variants keep each ledger use separately observable. Their original
13 diagnostic rows all refer to one node.parameters[0] read. Combining their
witness counts must not be interpreted as 13 guards emitted in the original
function. Likewise D060 and D061 are two diagnostics on one original read.
No whole-program completion date follows from these witness results. Four
blocked rows require attribution resolution or array-destructuring support.

Every mutant actually run in the final gate:

| Erased indexed panic in witness | Exit | stdout | stderr | Catcher |
|---|---:|---|---|---|
| D054 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D056 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D057 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D060-D061 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D067 | 0 | `0\n` | empty | exact named-site exit-70 observation |
| D068 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D069 | 70 | empty | `undefined where the checker narrowed it away: a call since the narrowing put it back` | exact named-site exit-70 observation |
| D077 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D082 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D083 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D084 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D085 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D086 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D087 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D088 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D089 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D090 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D091 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D092 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D093 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D094 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |
| D095 | 0 | `undefined\n` | empty | exact named-site exit-70 observation |

All 22 mutants retain the lookup and change only one emitted adamic_panic call to
(void)0. Each successfully builds with ASan and UBSan; a build failure fails the
test and is never counted as a caught mutant. D069 still panics later, but its
stderr is the generic narrowing message shown above rather than the required
indexed read site. The final log retains every mutant's exact stdout and stderr.
Blocked rows have no runtime mutant claim.

Final commands and observations, with complete output first written to files:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-a -count=1 -timeout 10m -v > /tmp/stricter-indexed-a-final.log 2>&1
# PASS; ok github.com/system-inc/adamic/stage3/stricter-indexed-a 129.065s
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-a-oracle.log 2>&1
# PASS; four fixtures; ok github.com/system-inc/adamic/internal/oracle 8.125s
go vet ./stage3/stricter-indexed-a > /tmp/stricter-indexed-a-vet.log 2>&1
# exit 0, empty output
gofmt -l stage3/stricter-indexed-a
# exit 0, empty output
git diff --check
# exit 0, empty output
```

The witness package runs source Node directly on generated .ts, backend Node,
release native, ASan/UBSan native, the actual CLI explain command, and all mutants.
Each finishing sanitized witness reports no sanitizer or leak failure. The filtered
existing oracle uses its normal cache: Node hits=0/misses=8, native hits=2/misses=8.
It is not claimed as wholly uncached. The full repository gate and WASI were not
run. Setup's initial dependency build took about 17 minutes; the worker gate is
the entire touched package plus the filtered existing oracle. Five group pushes
were made without opening a PR. The branch remains based on 390af985; no landing
into main or an area branch is claimed. Final report commit SHA is in the worker's
final response.
