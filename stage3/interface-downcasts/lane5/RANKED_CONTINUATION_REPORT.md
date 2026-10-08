Built: ranks 44, 54 and 60 certified with 15 original-declaration/read fixtures.
Commits: continued from 5d0ad15863bbff16aa651cb0fa68dede5a41f0f7; each group is committed and pushed separately to codex/views-callables.
Checks: Node, release and sanitized native, JavaScript, leaks and lane counts pass; required whole-table counts updater still fails inherited cases.
Mutants: native/JavaScript arity, result and parameter checks and deferred aggregate payload registration all caught for rank 44.
Uncovered: conservative 23/2818 pairs, 1008/11063 candidate reads certified; 2795 pairs and 10055 reads remain. Static totals unchanged at 4/308 and 34/1503.

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
