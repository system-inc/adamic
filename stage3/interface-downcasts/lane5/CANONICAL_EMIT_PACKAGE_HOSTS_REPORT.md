Built: original canonical name emit options and package-cache callable checks.
Commits: follows pushed 554cb1122f0d6e164f14a9b50848b7550177c2db on codex/views-callables.
Checks: 15 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 15 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 16 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 121/2818 pairs and 2476/11063 candidate reads certified; 2697 pairs and 8587 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 217,227,234 add 22 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 217,227,234 certify complete original Program.getCanonicalFileName, EmitHost.getCompilerOptions and ModuleResolutionCache.getPackageJsonInfoCache declarations and read expressions. GetCanonicalFileName retains the complete original callable alias from core.ts and the original getReferencedFiles helper argument read. Helper bodies and structural result carriers are reduced. The alias verifier can resolve configured original alias source files; changing the fixture alias parameter from string to number fails its original-alias assertion and is restored.

Seven runtime runs catch sixteen pair-level assertions: native/JavaScript arity and result for three, parameters for canonical names, and deferred value reads for both object results. No production compiler/runtime or calling convention changes.

Ranks 223 and 225 are retained as uncertified original method-read probes. realpath uses host.realpath?.bind(host); directoryExists uses host.directoryExists && directoryExists. Source Node prints abc/3 and true. Both corresponding CLI c commands refuse the cast with adamic/no-unchecked-cast. These are observed optional-method cast boundaries, not successful member-read certificates and not claims of an unbound-method diagnostic. They need checked-view admission for the original optional method contracts and preservation/proof of their original binding/condition contexts. They stay outside certification and counts. The directory probe's initial direct boolean console argument was corrected to interpolation; only the corrected source/read and lowering observations are used.

Exact supplemental commands: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 217; go run ./cmd/adamic c stage3/interface-downcasts/lane5/later-ranked-callables/realpath-binding/good.a and directory-condition/good.a, each redirected to its named probe log; node --disable-warning=ExperimentalWarning oracle/node.mjs with each probe path, each redirected to its named Node log. Original declaration/read verification also includes ranks 223 and 225; its 17 fixtures include the 15 certified fixtures and two excluded probes.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for canonical-file-name, emit-host-options, package-cache
javascript-arity: caught for canonical-file-name, emit-host-options, package-cache
native-result: caught for canonical-file-name, emit-host-options, package-cache
javascript-result: caught for canonical-file-name, emit-host-options, package-cache
native-parameters: caught for canonical-file-name
javascript-parameters: caught for canonical-file-name
payload-reads: caught for emit-host-options, package-cache
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/canonical-emit-package-hosts:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 217,227,234 > /tmp/lane5-canonical-optional-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(canonical-file-name|emit-host-options|package-cache)$' -count=1 -timeout 5m > /tmp/lane5-canonical-optional-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py canonical-file-name emit-host-options package-cache > /tmp/lane5-canonical-optional-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(canonical-file-name|emit-host-options|package-cache)$' -count=1 -timeout 5m > /tmp/lane5-canonical-optional-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-canonical-optional-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-canonical-optional-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	6.517s
restored: ok  	github.com/system-inc/adamic/internal/oracle	6.703s
counts: ok  	github.com/system-inc/adamic/internal/oracle	93.692s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
