Built: lexical environment, module format, false node and parenthesized type callable member reads checked through views.
Commits: follows pushed fe969d268d4a2e3a01d182c63f6f9ac5039f0c7f on codex/views-callables.
Checks: 22 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 22 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 23 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 74/2818 pairs and 2031/11063 candidate reads certified; 2744 pairs and 9032 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 135,137,142,143 add 50 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

endLexicalEnvironment keeps its original Statement[] | undefined result and context read. Present/absent results agree with Node; returned statement.value is checked when reached. getEmitModuleFormatOfFile keeps its original SourceFile parameter and ModuleKind result, with the enum carrier reduced to number. Its string-typed sourceFile.value producer prints then returns a numeric ModuleKind on Node; both backends stop at the earlier string read before either output line. createFalse preserves its zero-argument FalseLiteral result contract. createParenthesizedType preserves its TypeNode parameter and returned node, with type.value checked inside the producer.

Source resolution now records Debug.formatSyntaxKind's original exported function header separately from the fixture method signature, following debug.ts. Future original evidence includes ranks through 158. Overload, predicate/rest and intrinsic member families encountered between these supported members remain uncertified; no single overload or plain-object intrinsic substitute is counted. The original onSubstituteNode method/value boundary remains outside these fixtures and existing callback conventions stay intact.

The arrow family remains skipped as instructed. BindingName's original tagged carrier is certified in TAGGED_VARIABLE_CONDITIONAL_BINARY_DIAGNOSTIC_TEXT_REPORT.md; its earlier untagged reduction remains only a refusal probe. The new ForInitializer probe is being retained separately pending its proved union-read contract, rather than replacing the original union with one narrower carrier.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for lexical-end, module-format, false, parenthesized-type
javascript-arity: caught for lexical-end, module-format, false, parenthesized-type
native-result: caught for lexical-end, module-format, false, parenthesized-type
javascript-result: caught for lexical-end, module-format, false, parenthesized-type
native-parameters: caught for module-format, parenthesized-type
javascript-parameters: caught for module-format, parenthesized-type
payload-reads: caught for lexical-end, module-format, parenthesized-type
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/lexical-module-false-type:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 135,137,142,143 > /tmp/lane5-lexical-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(lexical-end|module-format|false|parenthesized-type)$' -count=1 -timeout 5m > /tmp/lane5-lexical-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py lexical-end module-format false parenthesized-type > /tmp/lane5-lexical-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(lexical-end|module-format|false|parenthesized-type)$' -count=1 -timeout 5m > /tmp/lane5-lexical-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-lexical-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-lexical-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	9.658s
restored: ok  	github.com/system-inc/adamic/internal/oracle	9.780s
counts: ok  	github.com/system-inc/adamic/internal/oracle	54.030s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
