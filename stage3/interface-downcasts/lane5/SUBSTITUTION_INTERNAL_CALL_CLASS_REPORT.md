Built: original stored substitution hook, internal name, call update and class update callable member reads checked through views.
Commits: follows pushed 0040752305d151dff23d234213edc47aaddbd305 on codex/views-callables.
Checks: 27 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 27 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 28 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 85/2818 pairs and 2159/11063 candidate reads certified; 2733 pairs and 8904 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 136,156,157,158 add 46 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

onSubstituteNode keeps its original function-property declaration and stored context.onSubstituteNode read in previousOnSubstituteNode. The stored callback is invoked with the original scalar hint and Node carrier, and node.value is checked lazily. This preserves a property-style function rather than rewriting a method to allow detachment. Source verification records declarationKind property-function. A separate source mutant changes that property declaration to a method; the complete original-declaration comparison refuses it, and the fixture is restored. Its direct log is lane5-updates-property-style-mutant.log. This one source-verifier mutant is additional to seven runtime runs and 28 pair-level runtime checks.

getInternalName retains its original Declaration parameter, optional comment/source-map flags and Identifier result. Omitted/undefined, false, true and mixed flags agree with Node; node.value is checked lazily. updateCallExpression retains all four original parameters, covering absent/empty/supplied type-argument arrays and empty/supplied argument arrays. Its reached argument.value producer read receives the string check.

updateClassDeclaration retains all six original parameters and tests absent, empty and supplied modifier/type-parameter/heritage arrays, optional identifier names and required member arrays. The reached member.value field is checked lazily. Adjacent structural carriers and enum aliases remain reduced as documented; the original callable declarations and reads are complete.

The arrow family remains skipped as instructed. ForInitializer's broad union remains an uncertified original-read refusal witness with its prerequisite recorded in PREFIX_SYNTAX_STATIC_UNARY_REPORT.md. Tagged BindingName, ModuleExportName and PropertyName carriers preserve their original aliases and kind tags; their earlier untagged reductions are separate probes, not substitutes for the original certificates. No callback calling-convention guard changed and no new integrator decision on the Union exception was received.

The complete later-ranked harness additionally passes all previous supported families and retained refusal witnesses. Complete source verification and alias/tag verification are preserved in the all-fixtures and all-carriers logs. The property-style source mutant runs verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 136 against the changed fixture, requires original declaration changed and restores it in finally.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for substitution-hook, internal-name, call-update, class-update
javascript-arity: caught for substitution-hook, internal-name, call-update, class-update
native-result: caught for substitution-hook, internal-name, call-update, class-update
javascript-result: caught for substitution-hook, internal-name, call-update, class-update
native-parameters: caught for substitution-hook, internal-name, call-update, class-update
javascript-parameters: caught for substitution-hook, internal-name, call-update, class-update
payload-reads: caught for substitution-hook, internal-name, call-update, class-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/substitution-internal-call-class:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 136,156,157,158 > /tmp/lane5-updates-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(substitution-hook|internal-name|call-update|class-update)$' -count=1 -timeout 5m > /tmp/lane5-updates-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py substitution-hook internal-name call-update class-update > /tmp/lane5-updates-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(substitution-hook|internal-name|call-update|class-update)$' -count=1 -timeout 5m > /tmp/lane5-updates-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-updates-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-updates-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	13.837s
restored: ok  	github.com/system-inc/adamic/internal/oracle	12.725s
counts: ok  	github.com/system-inc/adamic/internal/oracle	66.183s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
