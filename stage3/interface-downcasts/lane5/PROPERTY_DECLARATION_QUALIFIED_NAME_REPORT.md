Built: original tagged property declaration and qualified-name callable checks.
Commits: follows pushed 490b24478c9155ac73d2c70d3043a1408a55cc9b on codex/views-callables.
Checks: 14 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 14 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 14 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 97/2818 pairs and 2281/11063 candidate reads certified; 2721 pairs and 8782 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 176,177 add 20 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 176 and 177 certify complete original createPropertyDeclaration and createQualifiedName declarations and factory reads. PropertyName retains all seven original alternatives and kinds; EntityName retains Identifier | QualifiedName with original kinds 80 and 167. QuestionToken and ExclamationToken retain original kinds 58 and 54. Adjacent structural payloads and helpers are reduced. All name alternatives, string-name arguments, empty/absent modifiers and both token alternatives run against Node and both backends. Qualified-name recursion and the original compiler call graph are not executed by these carrier fixtures.

Seven runtime runs catch fourteen pair-level assertions: native/JavaScript arity, result and parameter admission, plus lazy modifier.value and left.value reads. Original carrier validation: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 176,177. Existing unsupported calling conventions remain unchanged.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for property-declaration, qualified-name
javascript-arity: caught for property-declaration, qualified-name
native-result: caught for property-declaration, qualified-name
javascript-result: caught for property-declaration, qualified-name
native-parameters: caught for property-declaration, qualified-name
javascript-parameters: caught for property-declaration, qualified-name
payload-reads: caught for property-declaration, qualified-name
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/property-declaration-qualified-name:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 176,177 > /tmp/lane5-property-qualified-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(property-declaration|qualified-name)$' -count=1 -timeout 5m > /tmp/lane5-property-qualified-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py property-declaration qualified-name > /tmp/lane5-property-qualified-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(property-declaration|qualified-name)$' -count=1 -timeout 5m > /tmp/lane5-property-qualified-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-property-qualified-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-property-qualified-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	7.170s
restored: ok  	github.com/system-inc/adamic/internal/oracle	7.331s
counts: ok  	github.com/system-inc/adamic/internal/oracle	75.338s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
