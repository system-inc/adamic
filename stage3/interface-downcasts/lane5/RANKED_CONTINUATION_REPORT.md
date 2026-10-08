Built: eight original callable pairs certified, adding 325 candidate reads and 49 fixtures across three pushed groups.
Commits: continued from 5d0ad15863bbff16aa651cb0fa68dede5a41f0f7; each group is committed and pushed separately to codex/views-callables.
Checks: Node, release and sanitized native, JavaScript, leaks and lane counts pass; required whole-table counts updater still fails inherited cases.
Mutants: native/JavaScript arity, result and parameter checks and deferred aggregate payload registration all caught across their applicable pairs: 25 mutation executions, all restored.
Uncovered: conservative 28/2818 pairs, 1235/11063 candidate reads certified; 2790 pairs and 9828 reads remain. Static totals unchanged at 4/308 and 34/1503.

Reporting date October 12, execution environment October 8. TypeScript truth pin 050880ce59e30b356b686bd3144efe24f875ebc8. Original declarations, source hashes and exact read spans are retained in ranked-next-original-witnesses.json. Fixtures preserve complete declarations and property read expressions; adjacent carriers are reduced. Read counts are inventory counts, not measured production reachability. Existing calling-convention refusals and the pending Union exception are untouched.

Group r44: original createReturnStatement(expression?: Expression): ReturnStatement and factory.createReturnStatement from nodeConverters.ts:56:33. Tests cover omitted/explicit undefined/object arguments, noncallables, arity, narrower parameters, incompatible results and an actual deferred expression.value read. Oracle passed 2.307s, lane counts update passed 18.805s. Seven measured counts rows added. Seven executable mutants fail the dedicated tests: arity bypasses execute rather than refuse; result bypasses reach later invalid object reads (native ASan SEGV) or later JS field refusal; parameter bypasses admit incompatible string producers; payload registration bypass reaches an ASan SEGV instead of the descendant read check. Mutation scripts restore every source in finally.

Exact commands, each test redirected directly to the indicated log:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run '^TestCheckedViewCallableRankedFamilies$' -count=1 -timeout 5m > /tmp/lane5-r44-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-ranked-callable-mutants.py return-statement > /tmp/lane5-r44-mutants.log 2>&1
node stage3/interface-downcasts/lane5/verify-ranked-callable-fixtures.cjs /workspace/scratch/lane5-ts-pin 44 > /tmp/lane5-r44-original.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-r44-counts.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-r44-all-counts.log 2>&1
```

Whole-table counts still fail inherited predicate/overload/process/nominal-write refusals and native invalid frees, as reported in the previous batch. Raw logs are retained under logs/ranked-continuation. No full package test, full gate, PR, main push or other branch push.

Group hooks: ranks 54 enableEmitNotification (32 reads) and 60 enableSubstitution (28 reads). Exact original SyntaxKind-to-void declarations and context member calls retained, enum members reduced to numeric witnesses. Eight fixtures pass Node, release/sanitized native, JavaScript and leaks; oracle 6.219s. Void return values are discarded, so no incompatible-result refusal is asserted. Each pair has native/JavaScript arity and parameter mutants; all four mutations caught both pairs. Original verifier confirms all eight fixtures. Commands are the r44 commands above with log prefix hooks, verifier ranks 54,60 and mutant arguments emit-notification substitution. Lane counts updater and required whole-table updater ran; raw outputs retained. Rank 44 was pushed at 21e31b4e.

Group aggregates: rank 14 createBlock (88 reads), 34 createArrayLiteralExpression (44), 39 createObjectLiteralExpression (41), 58 createFunctionExpression (28), and 62 updateBlock (26), totaling five pairs and 227 candidate reads. Thirty-four additional fixtures retain complete original declarations and original member reads, including context.factory.updateBlock. Seven-parameter function calls exercise undefined and supplied array/object arguments and string/object name alternatives. Nested view reads and deferred body.value/statements element reads are checked. Both aggregate pairs passed all seven runtime/compiler check mutants, followed by all three optional-Boolean pairs passing the same seven mutants. Native result/payload bypasses were caught by the expected read refusal and ASan diagnostics; JavaScript bypasses either executed or failed at a later field instead of the pinned member/payload read.

Correction to the prior gap assumption: the exact optional-Boolean declarations for ranks 14, 34 and 39 lower and execute successfully with the existing runtime descriptors. The new fixtures use multiLine in the result and compare omitted/undefined, false and true to Node. They are now certified, not treated as unsupported calling conventions. No production calling-convention guard was changed. The attempted refusal probe demonstrated successful lowering and was replaced with behaviorful backend fixtures before certification. The initial block wrong-members fixture expectation was corrected from 4 to Node's observed 3 because a boolean true is not the producer's string "true".

Final group commands:

```sh
go test ./internal/oracle -run '^TestCheckedViewCallableRankedFamilies$' -count=1 -timeout 5m > /tmp/lane5-aggregates-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-ranked-callable-mutants.py function-expression update-block > /tmp/lane5-aggregates-mutants.log 2>&1
python3 stage3/interface-downcasts/lane5/run-ranked-callable-mutants.py block array-literal object-literal > /tmp/lane5-booleans-mutants.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableRankedFamilies$' -count=1 -timeout 5m > /tmp/lane5-aggregates-final-oracle.log 2>&1
node stage3/interface-downcasts/lane5/verify-ranked-callable-fixtures.cjs /workspace/scratch/lane5-ts-pin 14,34,39,44,54,58,60,62 > /tmp/lane5-aggregates-original.log 2>&1
go vet ./internal/oracle > /tmp/lane5-aggregates-vet.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-aggregates-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m > /tmp/lane5-aggregates-counts-check.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-aggregates-all-counts.log 2>&1
```

Rank 58/62 initial oracle passed 12.138s. All 49 fixtures passed the restored oracle; exact original verifier confirms 49 complete declarations and member reads. Mutant detail logs name every check and pair; arity/parameter bypasses are caught by the pinned early member refusal, result bypasses by the same member-read pin versus later object failure, and payload bypasses by the descendant-read pin versus ASan SEGV. All production sources are restored and unchanged in this continuation. The only changes are fixtures, oracle/counts registration, evidence and inventory. Hook group pushed at 5ecc56cf.

Final restored oracle passed 17.748s; oracle vet passed; lane counts update passed 26.353s and verification passed 26.182s. Whole-table counts updater failed 39.722s on the previously recorded outside-lane failures. Exactly 49 fixture counts rows were added across the three groups; no pre-existing row changed. Conservative totals include the four prior certified pairs/34 reads in addition to ranked flags. Rank 17/18 original method bindings remain uncertified (136 reads). Generic, predicate, overload, rest and intrinsic boundaries remain outside these certifications; the integrator's pending Union callable decision was not changed.

Mutant executions by group: r44 and function-expression/update-block each ran native-arity, javascript-arity, native-result, javascript-result, native-parameters, javascript-parameters and payload-reads. The optional-Boolean group ran those same seven. Hooks ran native-arity, javascript-arity, native-parameters and javascript-parameters. All 25 runs failed the intended member/descendant-read checks for every selected pair and restored their mutation. Durable raw logs preserve sanitizer and compiler diagnostics verbatim, including whitespace in generated excerpts.
