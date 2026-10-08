Built: source-file path lookup and emit resolver member contracts checked through views; original literal-union and environment-method reads retained as refusals.
Commits: follows pushed baa72933a1504a6d186e5f307a0a4c9684c95615 on codex/views-callables.
Checks: 12 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 12 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 12 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 56/2818 pairs and 1785/11063 candidate reads certified; 2762 pairs and 9278 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 103,107 add 32 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

getSourceFileByPath preserves realProgram.getSourceFileByPath(file), its original Path parameter and SourceFile | undefined result. Present and absent results agree with Node; a reached sourceFile.value read checks its string payload lazily. getEmitResolver preserves context.getEmitResolver() and checks its returned object's reached value read. Native and JavaScript accept compatible aggregate carriers without eager payload admission.

The first source-file-path wrong-result control expected 0 incorrectly. Source Node actually prints undefined after reading value from a numeric result; the expectation was corrected before certification. The initial observation is preserved in lane5-env-node-expectation.log.

Ranks 102 and 106 remain uncertified original-read witnesses. literal-type/good.a retains createLiteralTypeNode(literal: LiteralTypeNode["literal"]): LiteralTypeNode and the original factory.createLiteralTypeNode read. Node prints 3; lowering refuses the reached union-carrier value read with unsupported untagged object union contract. This needs a proved representation and descendant-read contract for the untagged union. environment-variable/good.a retains !system.getEnvironmentVariable from the original read context. Node prints abc; lowering refuses unbound-method. Certifying that original read needs an agreed method-as-value convention; no property signature or arrow replacement is counted. TestCheckedViewCallableLaterRankedOriginalReadRefusals pins both Node outputs and refusals, and both refusal witnesses are excluded from native counts and certification.

The oracle and restored commands also include |^TestCheckedViewCallableLaterRankedOriginalReadRefusals$ in their -run expression. Fixture verification included ranks 102,103,106,107 and checked 14 complete declarations/reads: 12 supported fixtures plus two retained refusal witnesses. Intrinsic Map/SymbolTable, RegExp and array families remain outside these certificates; no plain-object substitute is certified for them.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for source-file-path, emit-resolver
javascript-arity: caught for source-file-path, emit-resolver
native-result: caught for source-file-path, emit-resolver
javascript-result: caught for source-file-path, emit-resolver
native-parameters: caught for source-file-path
javascript-parameters: caught for source-file-path
payload-reads: caught for source-file-path, emit-resolver
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/source-path-resolver:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 103,107 > /tmp/lane5-env-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(source-file-path|emit-resolver)$' -count=1 -timeout 5m > /tmp/lane5-env-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py source-file-path emit-resolver > /tmp/lane5-env-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(source-file-path|emit-resolver)$' -count=1 -timeout 5m > /tmp/lane5-env-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-env-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-env-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	6.026s
restored: ok  	github.com/system-inc/adamic/internal/oracle	5.664s
counts: ok  	github.com/system-inc/adamic/internal/oracle	48.553s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
