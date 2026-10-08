Built: original outer accessor parenthesized update and host source-file callable checks.
Commits: follows pushed 19b8c74ab84edcfadc97d56adab8f96d68ceaa18 on codex/views-callables.
Checks: 32 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 32 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 33 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 102/2818 pairs and 2326/11063 candidate reads certified; 2716 pairs and 8737 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 191,192,193,194,196 add 45 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 191,192,193,194,196 certify complete original restoreOuterExpressions, updateGetAccessorDeclaration, updateParenthesizedExpression, updateSetAccessorDeclaration and TypeCheckerHost.getSourceFiles declarations/read expressions. OuterExpressionKinds is reduced to its numeric representation. Adjacent structural payloads and helper bodies are reduced. Both accessor fixtures retain all seven original PropertyName alternatives and kinds, validated against the original checker; arrays exercise supplied and empty parameters, absent modifiers and missing type/body values. Optional outer-expression and kinds parameters exercise absent, explicit undefined and supplied values. Returned source-file elements retain deferred value reads. The whole original compiler and narrower literal-subtype producers are not covered.

Seven runtime runs catch 33 pair-level assertions: native/JavaScript arity and result for all five, parameters for four, and reached innerExpression.value, parameter.value, node.value and sourceFile.value reads. Each parameter/result representation check runs before invoking its producer. The original alias/tag verifier command is node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 192,194. No production compiler or callback calling convention changes.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update, host-source-files
javascript-arity: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update, host-source-files
native-result: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update, host-source-files
javascript-result: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update, host-source-files
native-parameters: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update
javascript-parameters: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update
payload-reads: caught for outer-expressions, get-accessor-update, parenthesized-update, set-accessor-update, host-source-files
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/outer-accessor-parenthesized-host:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 191,192,193,194,196 > /tmp/lane5-outer-accessors-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(outer-expressions|get-accessor-update|parenthesized-update|set-accessor-update|host-source-files)$' -count=1 -timeout 5m > /tmp/lane5-outer-accessors-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py outer-expressions get-accessor-update parenthesized-update set-accessor-update host-source-files > /tmp/lane5-outer-accessors-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(outer-expressions|get-accessor-update|parenthesized-update|set-accessor-update|host-source-files)$' -count=1 -timeout 5m > /tmp/lane5-outer-accessors-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-outer-accessors-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-outer-accessors-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	15.642s
restored: ok  	github.com/system-inc/adamic/internal/oracle	15.609s
counts: ok  	github.com/system-inc/adamic/internal/oracle	76.670s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
