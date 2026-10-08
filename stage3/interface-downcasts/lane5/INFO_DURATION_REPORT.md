Built: original Info canonical name and performance duration callable reads.
Commits: follows pushed 3729c616e5add52c1c1addab253717fe3754a562 on codex/views-callables.
Checks: 11 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 11 counts rows added with existing rows unchanged.
Mutants: 6 applicable runs caught and restored, 12 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 153/2818 pairs and 2681/11063 candidate reads certified; 2665 pairs and 8382 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 278,325 add 12 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 278 and 325 retain the original Info.readonly getCanonicalFileName property declaration and performance.getDuration exported-function signature plus their witnessed member reads. Info's original GetCanonicalFileName alias is independently verified in core.ts. The original member is passed to a reduced toPath helper with its original argument position and callable alias. That helper and surrounding source carriers are reduced. Zero/nonzero duration names are exercised against Node. No full module-specifier or performance implementation execution is claimed.

Changing the Info fixture alias parameter from string to number is caught by the original alias verifier with original alias changed GetCanonicalFileName; the fixture is restored. Runtime arity, result and parameter mutants each catch both pairs. Existing callback conventions and recorded Set/arrow requirements remain unchanged; no integrator Union exception decision received. Higher-ranked pending receiver/signature families remain listed in the preceding group reports. No class method was rewritten as a function property.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for info-canonical-name, performance-duration
javascript-arity: caught for info-canonical-name, performance-duration
native-result: caught for info-canonical-name, performance-duration
javascript-result: caught for info-canonical-name, performance-duration
native-parameters: caught for info-canonical-name, performance-duration
javascript-parameters: caught for info-canonical-name, performance-duration
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/info-duration:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 278,325 > /tmp/lane5-info-duration-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(info-canonical-name|performance-duration)$' -count=1 -timeout 5m > /tmp/lane5-info-duration-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py info-canonical-name performance-duration > /tmp/lane5-info-duration-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(info-canonical-name|performance-duration)$' -count=1 -timeout 5m > /tmp/lane5-info-duration-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-info-duration-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-info-duration-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	4.838s
restored: ok  	github.com/system-inc/adamic/internal/oracle	4.979s
counts: ok  	github.com/system-inc/adamic/internal/oracle	111.678s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
