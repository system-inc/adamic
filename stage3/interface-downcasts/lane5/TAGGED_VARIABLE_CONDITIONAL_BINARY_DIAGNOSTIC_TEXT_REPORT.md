Built: original-tagged variable update, conditional expression, binary update, config diagnostic and scanner text callable member reads checked through views.
Commits: follows pushed 1a3c524a5843a8eac3bc53b56cc2a556503e45cf on codex/views-callables.
Checks: 31 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 31 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 30 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 70/2818 pairs and 1981/11063 candidate reads certified; 2748 pairs and 9082 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 122,129,131,132,133 add 66 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

updateVariableDeclaration retains its complete five-parameter declaration, both original BindingName aliases, and original numeric kind tags: Identifier 80, ObjectBindingPattern 207 and ArrayBindingPattern 208. All three alternatives and supplied/undefined token, type and initializer carriers agree with Node. The reached name.value producer read is checked lazily. The earlier untagged reduction remains a refusal probe and is not evidence that this original tagged pair is blocked.

createConditionalExpression retains all five original parameters, covering absent and supplied question/colon tokens and the condition.value payload read. updateBinaryExpression preserves context.factory.updateBinaryExpression, the original scalar/token operator alternatives and node.value read. addConfigDiagnostic retains the original diagnostic creation call and checks diag.value when reached. setText retains string | undefined and both optional numeric parameters; absent text, omitted/undefined numeric values, positive values and zero agree with Node.

verify-tagged-callable-carriers.cjs reads the pinned original aliases and kind declarations directly, resolving numeric enum values through the original TypeScript checker. Seven variable-update fixtures pass. A kind mutant changes Identifier 80 to 79; a separate alias mutant replaces BindingPattern's original alternatives. The verifier refuses each at its corresponding source-preservation assertion, and both are restored. Direct logs are conditional-tagged-mutant-kind.log and conditional-tagged-mutant-alias.log. These two source-verifier mutants are additional to seven runtime mutation runs and 30 pair-level runtime checks.

The original-declaration fixture map now uses variable-update-tagged for rank 122. The untagged variable-update fixture remains only in the refusal test and outside native counts. No callable calling-convention or runtime code was changed. The arrow family remains skipped as instructed, with the original ConciseBody union-read prerequisite recorded in HELPERS_CASE_NEWLINE_REPORT.md.

Additional command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122 > /tmp/lane5-conditional-tagged.log 2>&1. The two source mutants each invoke that command against the changed fixture, require the specific verifier assertion, and restore the fixture in finally. Future verifier configuration also records original ModuleExportName, PropertyName and TypeOfTag carriers for subsequent groups.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for variable-update-tagged, conditional-expression, binary-update, config-diagnostic, scanner-text
javascript-arity: caught for variable-update-tagged, conditional-expression, binary-update, config-diagnostic, scanner-text
native-result: caught for variable-update-tagged, conditional-expression, binary-update
javascript-result: caught for variable-update-tagged, conditional-expression, binary-update
native-parameters: caught for variable-update-tagged, conditional-expression, binary-update, config-diagnostic, scanner-text
javascript-parameters: caught for variable-update-tagged, conditional-expression, binary-update, config-diagnostic, scanner-text
payload-reads: caught for variable-update-tagged, conditional-expression, binary-update, config-diagnostic
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/tagged-variable-conditional-binary-diagnostic-text:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 122,129,131,132,133 > /tmp/lane5-conditional-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(variable-update-tagged|conditional-expression|binary-update|config-diagnostic|scanner-text)$' -count=1 -timeout 5m > /tmp/lane5-conditional-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py variable-update-tagged conditional-expression binary-update config-diagnostic scanner-text > /tmp/lane5-conditional-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(variable-update-tagged|conditional-expression|binary-update|config-diagnostic|scanner-text)$' -count=1 -timeout 5m > /tmp/lane5-conditional-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-conditional-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-conditional-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	15.526s
restored: ok  	github.com/system-inc/adamic/internal/oracle	15.051s
counts: ok  	github.com/system-inc/adamic/internal/oracle	52.938s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
