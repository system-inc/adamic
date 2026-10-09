Built: merged the approved optional-field presence dependency and closed D069's initially absent-field context; 27 sites proven, 0 blocked, 0 remaining.
Commits: dependency tip 7e7464e6 merged at 2a519a00; records 28d30cd3 remains in ancestry through dfd3da59; this closure report follows the merge.
Commands: indexed witnesses PASS 173.287s, records PASS 67.047s, full affected compiler packages PASS, 45 selected Node oracle fixtures PASS 12.378s, measured counts PASS 82.977s, vet and formatting PASS.
Mutants: D069's erased indexed guard builds under sanitizers and reaches a different downstream panic; its exact indexed-site stderr catches the removal. Every indexed witness mutant passed its kill assertion again.
Not covered: the complete repository gate and the complete native package; focused native checks passed. These remain minimal ledger-shape witnesses, not a compilation of the whole TypeScript checker.

Approved dependency merge and D069 closure, October 8

The exact approved optional-field-write-2 commit
7e7464e632fb48386d39d7dce1236f72f57174b0 is the second parent of merge
2a519a00a0de773934261a7acaab7f948fc58e5a on codex/stricter-indexed-a.
No runtime file conflicted. All runtime files and the protected compiler
orchestrators matched Git's AUTO_MERGE tree before committing; none was
manually edited. Compiler conflicts preserve automatic owning-project TypeScript
options and direct execution entries alongside explicit project loading,
host console declarations, both record lowering paths, and shared indexed guards.

Native writes now use the dependency's adamic_object_write_field helper so a
uniform-slot optimization cannot skip optional-field presence and readiness
publication. The readonly record checks remain bounded; mutable record reflection
and spread use their own storage proof. Plain const literal aliases prove their
surface keys only, with hidden spreads and dynamic keys excluded; nested values
still undergo optional-view checks. Conservative base optional-union refusals
are retained. The formerly refused field-layout alias case was promoted into the
existing field_access_paths.a Node oracle; readonly record refusal cases were
promoted to supported forms. Array destructuring does not store a synthetic
structural view of the whole array, so prototype-view checks apply only at real
storage sites. D071-D073 still use the same indexedPresenceGuard as direct reads.

D069.json now declares typeAsPromise with an empty object, rather than a
preinitialized promisedTypeOfPromise field. Its assignment reads
getTypeArguments(type)[0] at main.ts:6:39. Source Node prints 7 for the present
case and undefined for the absent case, with empty stderr and exit 0.
Release native, sanitized native and backend JavaScript match Node when present.
When absent, all compiled executions have empty stdout, exit 70, and exactly:

```text
adamic: panic: indexed read is absent: <absolute main.ts path>:6:39
```

The test pins the actual absolute path, not this display placeholder. The actual
CLI --explain-checks reports checked: indexed-presence=1, trusted: 0 and that
location. The erased-guard mutant builds and runs under sanitizers; it exits 70
with empty stdout and this different message:

```text
adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back
```

That downstream check does not satisfy the indexed-site contract. Comparing the
complete observation, including exact named stderr, kills the mutant. The active
pinned optional-field failure was removed; earlier failure logs below are history.

Current row states are in sites.csv. D037, D071, D072, D073 and D069 are proven;
all other assigned rows are proven too. There are 27 proven rows, none blocked
and none remaining. The final witness log records every erased-site observation,
including the independently observable third destructuring element.

Validation commands and observations:

```sh
go test ./stage3/stricter-indexed-a ./stage3/stricter-records -count=1 -timeout 15m -v
# PASS 173.287s / 67.047s, including all indexed mutants and attribution/census tests.
go test ./internal/load ./internal/lower ./cmd/adamic ./internal/ir ./internal/javascript -count=1 -timeout 15m
# PASS load 65.392s, lower 153.713s, cmd 21.970s, ir 88.049s; JavaScript has no package tests.
go test ./internal/lower -run 'Test(Optional|Record|Project|NodeFS|LibraryPrototype|Prototype)' -count=1 -timeout 15m
# PASS 13.712s.
go test ./internal/native -run 'Test(Borrow|FieldReturn|RecordsAgainstNode|RecordMutants|RecordReadMutants|ExactReceiver|Devirtualized|InheritanceMemory)' -count=1
# PASS 319.506s.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(optional_field_|optional_widening|records_)' -count=1 -timeout 15m -v
# PASS 45 selected fixtures, 12.378s.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/field_access_paths.a$' -count=1 -v
# PASS 2.960s.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/fresh/testdata/regexp_tree.ts$' -count=1 -v
# PASS 1.688s.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts
# PASS 82.977s; conflicting count inventory replaced by measured output.
go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./cmd/adamic ./stage3/stricter-indexed-a ./stage3/stricter-records
gofmt -l cmd internal
# Both exit 0 with empty output.
```

All test output was written directly to log files. Evidence is in logs/approved-*.txt.
The first count attempt refused the incoming regexp_tree.ts fixture because it was
outside its project's roots. Its sole testdata directory now has an explicit
strict tsconfig naming that existing fixture; the loader refusal was preserved,
and both its Node oracle and the full count measurement then passed.
The locked stage3/api Node declaration dependencies were installed with npm ci.
The existing cloud toolchain was reused: setup previously completed in 1009.044s,
submodules ready at 234.251s and build ready at 1008.454s; nproc remains 5,
with a four-core cgroup quota. Original timing evidence is in logs/setup.txt.

Historical reports below are superseded by the closure above.

Built the requested records merge at dfd3da59; D069 initially absent named-field context still fails on valid input.
Commits: records tip 28d30cd3 merged into codex/stricter-indexed-a; no optional-field implementation was applied.
Commands: D069 context FAILED in 45.841s; merged indexed package PASS 170.732s, records PASS 112.028s, lower PASS 61.707s, oracle PASS 9.998s, vet PASS.
Mutants: D069 erased indexed guard reaches the missing-field panic; this is not a valid context proof because its present case already fails.
Not covered: D069 context closure requires optional named-field presence/storage; automatic review rejected the broad dependency merge and its low-level conflict resolution.

Initial witness-only report, retained as historical evidence

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

Follow-up: D071-D073

Dense-array declarations evaluate their receiver once and emit one ArrayIndex per
plain binding, in order. Direct reads and binding reads both call
indexedPresenceGuard in internal/lower/indexed_checks.go. The binding's source
position supplies the named message. Omitted positions emit no value read;
defaults, rest and nested bindings still refuse. Arrays outside project .ts checks
refuse a binding whose type omits undefined on exhaustion. Nullable/tagged/weak
presence representations retain the direct-read helper's refusals.

D071 and D072 each list three binding guards at line 5, columns 8, 19 and 31;
D073 lists one at line 5, column 8. D071 supplies an empty array, D072 supplies
only the first element, and D073 supplies an empty array. Node prints undefined;
both backends and release/sanitized native stop at the selected binding with
exit 70 and exact indexed read is absent stderr. The mutant erases only that
binding's panic. The next binding catches D071/D072, but at a different named
site; the required selected-site assertion kills each mutant. D073's mutant
finishes and prints undefined. No compiler warning is counted as a mutant kill.

Command: go test ./stage3/stricter-indexed-a -run
'TestCheckerIndexedWitnesses/(D071|D072|D073)$' -count=1 -timeout 10m -v.
The recorded expanded filter also included a slash-containing supplemental name;
it did not select that probe, which is run separately. Logs/destructuring.txt
retains the complete three-row evidence. Earlier report sections describe the
initial witness-only state; this follow-up supersedes their destructuring refusals.

Follow-up: D037

The original helper referenced through cohere/TypeScript/tsc/testdata/fixtures/
compiler/utilities.ts:980-984 has separate present and optional overloads. No
source was copied from cohere. The prior minimal probe replaced that relationship
with a single optional-return declaration, losing option attribution before
Adamic could audit it. Restoring the independently written overload shape exposes
a second boundary: declareModule attempted to lower the bodyless overloads as
executable functions. It now registers only their implementation, after checking
that overload parameters and result have the same storage representation.
Generic or different-storage overloads still refuse; ambient functions without
an implementation retain their refusal.

The checker keeps overload selection and its downstream indexed-option diagnostic.
The audit test proves the receiving assignment is accepted under the project's
options and rejected specifically by noUncheckedIndexedAccess. Removing the
overloads restores an ordinary optional-return error, which the loader still
refuses. A different-storage overload test also requires refusal. No general
optional-return error is relabeled as an indexed read.

D037 guards declarations[0] at line 10, column 34 through the same direct-read
helper. Node prints 7 or undefined. Release native, sanitized native and backend
JavaScript agree when present and pin the named indexed stop when absent. Explain
lists precisely one check. Its single-panic-erasure mutant builds sanitized,
finishes with exit 0 and prints undefined, failing the exact named-stop assertion.
Logs/optional-return.txt records the 14.785s passing gate. The supplemental third
binding similarly passes in 8.809s: its mutant prints undefined with exit 0.

The whole lowering package passed in 36.790s. The four existing indexed/narrowing
oracle fixtures passed in 1.537s, with the normal oracle cache, not an uncached
integration gate. Complete witness-package verification subsequently passed in 197.769s.

Follow-up: D069 initially absent optional-field context

This context was rerun with the updated compiler. Node exits 0 with 7 when the
array element is present and undefined when it is absent. Explain lists the
correct array check at line 6, column 39 with indexed-presence=1 and trusted=0.
Both native modes build successfully. With an absent array element both stop at
that named indexed guard. With a present element both instead stop at the field
write, exit 70, empty stdout and exact stderr:

adamic: panic: compiler bug: a field the checker proved is there is missing

That is an observed runtime failure on valid input, not a successful presence
witness or a proper compile-time NotYet refusal. No new mutant claim is made for
this context because its present case fails before a sound witness can be claimed.
The existing already-present-field D069 witness and its mutant still prove the
indexed read itself.

Receiver: a plain fixed-shape object with an initially absent optional own field,
receiving a readonly object-array element. The dependency is own-field storage
addition/presence, assigned by the user to codex/stricter-records. It is not an
array hole. Observed objectLiteral keeps only the source's actual fields;
SetProperty asks writeFieldSlot for existing data storage; object_find panics when
the shape has no such field. Inference: correct field addition needs the records
worker's representation, including own-presence behavior and identity-preserving
storage. Preallocating an undefined property here would change hasOwnProperty and
would not prove the initially absent case. No representation workaround was built.
Logs/absent-field-followup.txt retains Node, explain, build and runtime observations.

Current requested row states:

| Row | State |
|---|---|
| D071 | proven, yieldType binding; shared indexed guard helper; mutant caught at different next-binding site |
| D072 | proven, returnType binding; first element present; shared helper; mutant caught at different next-binding site |
| D073 | proven, single yieldType binding; shared helper; mutant prints undefined |
| D037 | proven, overload attribution retained; direct argument guard; mutant prints undefined |
| D069 | indexed read proven; initially absent optional own-field context blocked on codex/stricter-records storage; valid-input runtime bug pinned |

The supplemental nextType binding proves the third guard independently: first two
array elements present, third absent, exact third-site stop, erased third guard
prints undefined. All successful new witnesses run source Node, backend Node,
release native, sanitized native and actual CLI explain. No holes or record
representation changes are included. No protected orchestration file was edited.

Follow-up final gate

The complete witness package passes in 197.769s. It proves 27 ledger read-shape
rows with 26 ledger witness programs, plus one supplemental third-binding program.
All 27 emitted-C single-site panic-erasure mutants build sanitized and are caught
by exact indexed-site observations. D071/D072 stop at the next binding after the
selected panic is erased; D069's already-present-field mutant hits its existing
later generic narrowing panic; all remaining mutants exit 0 with changed output.
This is minimal read-shape coverage; D069's newly requested initially absent field
context remains blocked and is excluded from runtime-proof counts.

Commands and output, written to files before inspection:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-a -count=1 -timeout 10m -v > /tmp/indexed-a-followup-final.log 2>&1
# PASS; 197.769s; logs/followup-final.txt
go test ./internal/lower -count=1 -timeout 15m > /tmp/indexed-a-lower.log 2>&1
# PASS; 36.790s; logs/followup-lower.txt
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/indexed-a-followup-oracle.log 2>&1
# PASS; 1.537s; logs/followup-oracle.txt
go vet ./internal/lower ./stage3/stricter-indexed-a > /tmp/indexed-a-followup-vet.log 2>&1
# exit 0; empty log
gofmt -l internal/lower stage3/stricter-indexed-a
# exit 0; empty output
git diff --check
# exit 0; empty output
```

The full repository gate, WASI and whole-program compiler build were not run.
The exact row-state table above is the final follow-up disposition. af274a79
pushed the shared destructuring helper and witnesses; 3d60b65e pushed the overload
registration/attribution fix, audit/refusal tests and D037 mutant. This report and
D069 blocked-case evidence are committed and pushed afterward to the same branch.
No PR, main/area push, history rewrite or representation-worker implementation
is included.

Records merge requested by the user

Merged the exact published codex/stricter-records tip 28d30cd3 into this branch
as dfd3da59. The merge was clean; records MapGet uses our shared indexedPresenceGuard.
The records report explicitly limits the new table representation to readonly
index signatures without named fields. It does not implement D069's initially
absent optional named field on a plain fixed-shape object.

The D069 witness was changed locally to initialize typeAsPromise with {} and run
through the full witness harness. Source Node passes, printing 7 when the returned
array element is present. Release native instead exits 70 with empty stdout and:

adamic: panic: compiler bug: a field the checker proved is there is missing

The absent-element named guard and its erasure test pass, but that cannot prove a
context whose valid present input fails. The run failed in 45.841s. Complete
evidence is in logs/d069-after-records.txt. The failing local fixture change was
reverted; the passing preinitialized-field witness and historical gap input remain.

The published optional-field-write-2 report identifies the actual named-field
storage implementation at 7e7464e6. Automatic approval review rejected merging
that entire branch because it contains broad unrelated history beyond the user’s
records merge. A narrow cherry-pick of its batch-layout implementation was aborted
because the layouts conflict. The original implementation, 514b9361, changes only
27 optional-field files and avoids those batch-layout dependencies. Its patch is
saved for review at /tmp/indexed-a-optional-field-dependency.patch.

The original cherry-pick has five conflict files: expression.go, fields_test.go,
adamic.h, object.c and counts.md. Current optimized native field writes also bypass
presence publication, so a correct adaptation must preserve read optimizations
while ensuring writes publish the new per-slot presence state. Automatic approval
review rejected that compiler/runtime conflict resolution for its low-level impact
and insufficient authorization for those changes. Both partial cherry-picks were
aborted. No rejected changes remain in this branch.

Approval is needed to integrate the narrow optional-field storage implementation
and adapt its presence-aware runtime writes. Records are merged; D069 context
closure is blocked, not done. This supersedes the earlier inference that the
readonly-record worker alone supplied the named-field storage dependency.

Records-merge verification and final disposition

The authorized merge's complete indexed witness package passes in 170.732s,
including all 27 existing guard mutants and attribution checks. The complete
records witness package passes in 112.028s. Lowering passes in 61.707s; the four
indexed/narrowing oracle fixtures pass in 9.998s using their normal cache. Vet
passes with empty output. These results validate the records merge and existing
witnesses; they do not validate D069's initially absent receiving-field context.

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-a ./stage3/stricter-records -count=1 -timeout 15m -v > /tmp/indexed-a-records-merged-gate.log 2>&1
# PASS; 170.732s and 112.028s
go test ./internal/lower -count=1 -timeout 10m > /tmp/indexed-a-records-merged-lower.log 2>&1
# PASS; 61.707s
go vet ./internal/lower ./stage3/stricter-indexed-a ./stage3/stricter-records > /tmp/indexed-a-records-merged-vet.log 2>&1
# exit 0
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/indexed-a-records-merged-oracle.log 2>&1
# PASS; 9.998s
git diff --check
# exit 0
```

All output was written to logs before reading. The merged package, lower, oracle
and vet logs are checked in. Records merge and this report are pushed only to
codex/stricter-indexed-a. The dependency patch is available for concrete review;
approval is needed for its low-level optional-field integration. Remaining
requested context without a successful runtime witness: D069, one. No full
repository gate or whole-program compiler claim is made.
