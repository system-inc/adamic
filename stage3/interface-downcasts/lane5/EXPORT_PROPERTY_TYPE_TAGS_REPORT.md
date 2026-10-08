Built: original-tagged export and property names plus TypeOfTag callable member reads checked through views.
Commits: follows pushed ad2e41d59e6b0ff5464585d263a9b564a94efd04 on codex/views-callables.
Checks: 21 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 21 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 21 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 77/2818 pairs and 2067/11063 candidate reads certified; 2741 pairs and 8996 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 141,144,145 add 36 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

createExportSpecifier retains its original Boolean and string | ModuleExportName parameters, with ModuleExportName = Identifier | StringLiteral preserved. Identifier and StringLiteral retain the original numeric kind tags 80 and 11. Undefined, scalar names and both object alternatives agree with Node; a string-typed propertyName.value producer receives its check at that field read.

createPropertySignature retains all four original parameters and the exact PropertyName alias. Every original alternative is exercised: Identifier, StringLiteral, NoSubstitutionTemplateLiteral, NumericLiteral, ComputedPropertyName, PrivateIdentifier and BigIntLiteral. Their readonly numeric kind tags are read from the original checker, not invented. Scalar names and omitted modifier/token/type carriers are covered too. The modifier array's reached modifier.value field is checked lazily.

createTypeCheck retains the exact original TypeOfTag alias and exercises all nine alternatives: null, undefined, number, bigint, boolean, string, symbol, object and function. Its reached value.value field receives the producer's string check. These fixtures test the declared parameter/result representations with reduced adjacent data carriers; they do not claim certification of every narrower literal-subtype producer.

Twenty-one fixtures pass both complete declaration/read verification and original-carrier alias/tag verification. The carrier verifier's shared alias and numeric-tag assertions were independently proved by the restored source mutants in TAGGED_VARIABLE_CONDITIONAL_BINARY_DIAGNOSTIC_TEXT_REPORT.md. The existing callable representation guards and callback conventions are unchanged; no broader union is replaced with a narrower synthetic tagged alias.

Additional command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 141,144,145 > /tmp/lane5-exportnames-tagged.log 2>&1. Original numeric tag/alias facts are preserved separately in original-carrier-source.json. Arrow remains skipped with the recorded ConciseBody read-contract prerequisite.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for export-specifier, property-signature, type-check
javascript-arity: caught for export-specifier, property-signature, type-check
native-result: caught for export-specifier, property-signature, type-check
javascript-result: caught for export-specifier, property-signature, type-check
native-parameters: caught for export-specifier, property-signature, type-check
javascript-parameters: caught for export-specifier, property-signature, type-check
payload-reads: caught for export-specifier, property-signature, type-check
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/export-property-type-tags:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 141,144,145 > /tmp/lane5-exportnames-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(export-specifier|property-signature|type-check)$' -count=1 -timeout 5m > /tmp/lane5-exportnames-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py export-specifier property-signature type-check > /tmp/lane5-exportnames-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(export-specifier|property-signature|type-check)$' -count=1 -timeout 5m > /tmp/lane5-exportnames-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-exportnames-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-exportnames-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	10.238s
restored: ok  	github.com/system-inc/adamic/internal/oracle	10.893s
counts: ok  	github.com/system-inc/adamic/internal/oracle	60.933s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
